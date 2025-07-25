package xyz.zzhe.wslatency.server;

import io.netty.bootstrap.ServerBootstrap;
import io.netty.buffer.ByteBuf;
import io.netty.buffer.Unpooled;
import io.netty.channel.*;
import io.netty.channel.epoll.EpollEventLoopGroup;
import io.netty.channel.epoll.EpollServerSocketChannel;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioServerSocketChannel;
import io.netty.handler.codec.http.*;
import io.netty.handler.codec.http.websocketx.*;
import io.netty.handler.logging.LogLevel;
import io.netty.handler.logging.LoggingHandler;
import io.netty.handler.timeout.IdleStateHandler;
import io.netty.util.CharsetUtil;
import io.netty.util.internal.PlatformDependent;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import xyz.zzhe.wslatency.common.util.JsonUtils;
import xyz.zzhe.wslatency.common.util.TimeUtils;

import java.time.Instant;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.TimeUnit;

/**
 * High-performance WebSocket server for latency testing using Netty.
 */
public class WebSocketServer {
    private static final Logger logger = LoggerFactory.getLogger(WebSocketServer.class);
    private static final String WEBSOCKET_PATH = "/ws";
    private static final String HEALTH_PATH = "/health";

    private final ServerConfig config;
    private final EventGenerator eventGenerator;
    
    private EventLoopGroup bossGroup;
    private EventLoopGroup workerGroup;
    private Channel serverChannel;

    /**
     * Create a new WebSocket server with the specified configuration.
     *
     * @param config The server configuration
     */
    public WebSocketServer(ServerConfig config) {
        this.config = config;
        this.eventGenerator = new EventGenerator(config);
    }

    /**
     * Start the WebSocket server.
     */
    public void start() {
        try {
            // Configure event loop groups based on available transports
            boolean useEpoll = false;
            try {
                // Check if we're on Linux and if Epoll is available
                Class.forName("io.netty.channel.epoll.Epoll");
                useEpoll = io.netty.channel.epoll.Epoll.isAvailable();
                if (useEpoll) {
                    logger.info("Using native epoll transport for optimal performance");
                }
            } catch (ClassNotFoundException e) {
                // Epoll not available
                logger.info("Native epoll transport not available");
            }
            
            if (useEpoll) {
                logger.info("Using native epoll transport for optimal performance");
                bossGroup = new EpollEventLoopGroup(1);
                workerGroup = new EpollEventLoopGroup();
            } else {
                logger.info("Using NIO transport");
                bossGroup = new NioEventLoopGroup(1);
                workerGroup = new NioEventLoopGroup();
            }

            // Configure server bootstrap
            ServerBootstrap bootstrap = new ServerBootstrap();
            bootstrap.group(bossGroup, workerGroup)
                    .channel(useEpoll ? EpollServerSocketChannel.class : NioServerSocketChannel.class)
                    .handler(new LoggingHandler(LogLevel.INFO))
                    .childHandler(new WebSocketServerInitializer())
                    .option(ChannelOption.SO_BACKLOG, 128)
                    .childOption(ChannelOption.SO_KEEPALIVE, true)
                    .childOption(ChannelOption.TCP_NODELAY, true); // Disable Nagle's algorithm for lower latency

            // Bind and start to accept incoming connections
            serverChannel = bootstrap.bind(config.getPort()).sync().channel();
            logger.info("WebSocket server started on port {}", config.getPort());
            logger.info("WebSocket endpoint available at ws://0.0.0.0:{}{}", config.getPort(), WEBSOCKET_PATH);
            logger.info("Health check endpoint available at http://0.0.0.0:{}{}", config.getPort(), HEALTH_PATH);

            // Start the event generator
            eventGenerator.start();
            logger.info("Event generator started with rate {} events/sec", config.getEventsPerSecond());

        } catch (Exception e) {
            logger.error("Error starting WebSocket server", e);
            stop();
            throw new RuntimeException("Failed to start WebSocket server", e);
        }
    }

    /**
     * Stop the WebSocket server.
     */
    public void stop() {
        // Stop the event generator
        if (eventGenerator != null) {
            eventGenerator.stop();
            logger.info("Event generator stopped");
        }

        // Close server channel
        if (serverChannel != null) {
            serverChannel.close();
        }

        // Shutdown event loop groups
        if (bossGroup != null) {
            bossGroup.shutdownGracefully();
        }
        if (workerGroup != null) {
            workerGroup.shutdownGracefully();
        }
        
        logger.info("WebSocket server stopped");
    }

    /**
     * WebSocket server initializer.
     */
    private class WebSocketServerInitializer extends ChannelInitializer<SocketChannel> {
        @Override
        protected void initChannel(SocketChannel ch) {
            ChannelPipeline pipeline = ch.pipeline();
            
            // Add idle state handler to detect and close idle connections
            pipeline.addLast(new IdleStateHandler(0, 0, 300, TimeUnit.SECONDS));
            
            // HTTP codec
            pipeline.addLast(new HttpServerCodec());
            
            // Aggregate HTTP messages
            pipeline.addLast(new HttpObjectAggregator(65536));
            
            // Route based on the request URI
            pipeline.addLast(new WebSocketServerHandler());
        }
    }

    /**
     * WebSocket server handler.
     */
    private class WebSocketServerHandler extends SimpleChannelInboundHandler<Object> {
        private WebSocketServerHandshaker handshaker;

        @Override
        protected void channelRead0(ChannelHandlerContext ctx, Object msg) {
            if (msg instanceof FullHttpRequest) {
                handleHttpRequest(ctx, (FullHttpRequest) msg);
            } else if (msg instanceof WebSocketFrame) {
                handleWebSocketFrame(ctx, (WebSocketFrame) msg);
            }
        }

        private void handleHttpRequest(ChannelHandlerContext ctx, FullHttpRequest req) {
            // Handle a bad request
            if (!req.decoderResult().isSuccess()) {
                sendHttpResponse(ctx, req, new DefaultFullHttpResponse(
                        HttpVersion.HTTP_1_1, HttpResponseStatus.BAD_REQUEST));
                return;
            }

            // Handle health check request
            if (req.uri().equals(HEALTH_PATH)) {
                handleHealthCheck(ctx, req);
                return;
            }

            // Handle WebSocket handshake
            if (req.uri().equals(WEBSOCKET_PATH)) {
                WebSocketServerHandshakerFactory wsFactory = new WebSocketServerHandshakerFactory(
                        getWebSocketLocation(req), null, true);
                handshaker = wsFactory.newHandshaker(req);
                if (handshaker == null) {
                    WebSocketServerHandshakerFactory.sendUnsupportedVersionResponse(ctx.channel());
                } else {
                    handshaker.handshake(ctx.channel(), req);
                    
                    // Register the new client with the event generator
                    eventGenerator.addClient(ctx.channel());
                }
                return;
            }

            // Handle other HTTP requests with 404
            sendHttpResponse(ctx, req, new DefaultFullHttpResponse(
                    HttpVersion.HTTP_1_1, HttpResponseStatus.NOT_FOUND));
        }

        private void handleHealthCheck(ChannelHandlerContext ctx, FullHttpRequest req) {
            // Create health status response
            Map<String, Object> healthStatus = new HashMap<>();
            healthStatus.put("status", "healthy");
            healthStatus.put("timestamp", Instant.now().toString());
            healthStatus.put("version", "1.0.0");

            Map<String, Object> metrics = new HashMap<>();
            metrics.put("connections", eventGenerator.getClientCount());
            metrics.put("events_sent", eventGenerator.getEventsSent());
            healthStatus.put("metrics", metrics);

            // Convert to JSON
            String jsonResponse = JsonUtils.toJson(healthStatus);
            if (jsonResponse == null) {
                jsonResponse = "{\"status\":\"healthy\"}";
            }

            // Create HTTP response
            FullHttpResponse response = new DefaultFullHttpResponse(
                    HttpVersion.HTTP_1_1, 
                    HttpResponseStatus.OK,
                    Unpooled.copiedBuffer(jsonResponse, CharsetUtil.UTF_8));
            
            response.headers().set(HttpHeaderNames.CONTENT_TYPE, "application/json");
            response.headers().set(HttpHeaderNames.CONTENT_LENGTH, response.content().readableBytes());
            
            // Send response
            sendHttpResponse(ctx, req, response);
        }

        private void handleWebSocketFrame(ChannelHandlerContext ctx, WebSocketFrame frame) {
            // Handle close frame
            if (frame instanceof CloseWebSocketFrame) {
                handshaker.close(ctx.channel(), (CloseWebSocketFrame) frame.retain());
                eventGenerator.removeClient(ctx.channel());
                return;
            }
            
            // Handle ping frame
            if (frame instanceof PingWebSocketFrame) {
                ctx.write(new PongWebSocketFrame(frame.content().retain()));
                return;
            }
            
            // Handle text frame
            if (frame instanceof TextWebSocketFrame) {
                // Process the received message
                String request = ((TextWebSocketFrame) frame).text();
                processMessage(ctx, request);
                return;
            }
            
            // Unsupported frame type
            throw new UnsupportedOperationException(
                    String.format("%s frame types not supported", frame.getClass().getName()));
        }

        private void processMessage(ChannelHandlerContext ctx, String message) {
            try {
                // Parse message as JSON
                Map<String, Object> data = JsonUtils.fromJson(message, Map.class);
                
                // Add server timestamp
                if (data.containsKey("_test")) {
                    @SuppressWarnings("unchecked")
                    Map<String, Object> test = (Map<String, Object>) data.get("_test");
                    test.put("server_ts_ns", TimeUtils.getCurrentTimeNanos());
                } else {
                    Map<String, Object> test = new HashMap<>();
                    test.put("server_ts_ns", TimeUtils.getCurrentTimeNanos());
                    data.put("_test", test);
                }
                
                // Send response back
                String response = JsonUtils.toJson(data);
                ctx.channel().writeAndFlush(new TextWebSocketFrame(response));
                
                // No metrics to update
            } catch (Exception e) {
                logger.error("Error processing message: {}", e.getMessage());
            }
        }

        private void sendHttpResponse(ChannelHandlerContext ctx, FullHttpRequest req, FullHttpResponse res) {
            // Send the response and close the connection if necessary
            if (res.status().code() != 200) {
                ByteBuf buf = Unpooled.copiedBuffer(res.status().toString(), CharsetUtil.UTF_8);
                res.content().writeBytes(buf);
                buf.release();
                HttpUtil.setContentLength(res, res.content().readableBytes());
            }

            // Send the response and close the connection if necessary
            ChannelFuture f = ctx.channel().writeAndFlush(res);
            if (!HttpUtil.isKeepAlive(req) || res.status().code() != 200) {
                f.addListener(ChannelFutureListener.CLOSE);
            }
        }

        private String getWebSocketLocation(FullHttpRequest req) {
            String location = req.headers().get(HttpHeaderNames.HOST) + WEBSOCKET_PATH;
            // Check if the request is over HTTPS/SSL
            // X-Forwarded-Proto is a standard header but not defined in HttpHeaderNames
            String xForwardedProto = "X-Forwarded-Proto";
            String scheme = req.headers().contains(xForwardedProto) ?
                req.headers().get(xForwardedProto) : "http";
            
            // Use wss:// for HTTPS requests, ws:// for HTTP requests
            return (scheme.equalsIgnoreCase("https") ? "wss://" : "ws://") + location;
        }

        @Override
        public void channelReadComplete(ChannelHandlerContext ctx) {
            ctx.flush();
        }

        @Override
        public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
            logger.error("WebSocket error: {}", cause.getMessage());
            ctx.close();
        }

        @Override
        public void handlerAdded(ChannelHandlerContext ctx) {
            logger.debug("Handler added: {}", ctx.channel().id().asShortText());
        }

        @Override
        public void handlerRemoved(ChannelHandlerContext ctx) {
            logger.debug("Handler removed: {}", ctx.channel().id().asShortText());
            eventGenerator.removeClient(ctx.channel());
        }
    }
}