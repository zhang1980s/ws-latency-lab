package xyz.zzhe.wslatency.rtt.server;

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
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import xyz.zzhe.wslatency.common.util.JsonUtils;
import xyz.zzhe.wslatency.common.util.TimeUtils;
import xyz.zzhe.wslatency.rtt.model.RttMessage;

import java.time.Instant;
import java.util.HashMap;
import java.util.Map;
import java.util.Random;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicLong;

/**
 * WebSocket RTT server for latency testing using Netty.
 * This server implements a request-response model for RTT measurement.
 */
public class WebSocketRttServer {
    private static final Logger logger = LoggerFactory.getLogger(WebSocketRttServer.class);
    private static final String WEBSOCKET_PATH = "/ws";
    private static final String HEALTH_PATH = "/health";
    private static final String ALPHABET = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";

    private final RttServerConfig config;
    private final Random random = new Random();
    private final AtomicLong messagesReceived = new AtomicLong(0);
    private final AtomicLong messagesSent = new AtomicLong(0);
    
    private EventLoopGroup bossGroup;
    private EventLoopGroup workerGroup;
    private Channel serverChannel;

    /**
     * Create a new WebSocket RTT server with the specified configuration.
     *
     * @param config The server configuration
     */
    public WebSocketRttServer(RttServerConfig config) {
        this.config = config;
    }

    /**
     * Start the WebSocket RTT server.
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
                    .childHandler(new WebSocketRttServerInitializer())
                    .option(ChannelOption.SO_BACKLOG, 128)
                    .childOption(ChannelOption.SO_KEEPALIVE, true)
                    .childOption(ChannelOption.TCP_NODELAY, true); // Disable Nagle's algorithm for lower latency

            // Bind and start to accept incoming connections
            serverChannel = bootstrap.bind(config.getPort()).sync().channel();
            logger.info("WebSocket RTT server started on port {}", config.getPort());
            logger.info("WebSocket endpoint available at ws://0.0.0.0:{}{}", config.getPort(), WEBSOCKET_PATH);
            logger.info("Health check endpoint available at http://0.0.0.0:{}{}", config.getPort(), HEALTH_PATH);

        } catch (Exception e) {
            logger.error("Error starting WebSocket RTT server", e);
            stop();
            throw new RuntimeException("Failed to start WebSocket RTT server", e);
        }
    }

    /**
     * Stop the WebSocket RTT server.
     */
    public void stop() {
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
        
        logger.info("WebSocket RTT server stopped");
    }

    /**
     * Generate a random payload of the specified size.
     *
     * @param size The payload size in bytes
     * @return The random payload
     */
    private String generateRandomPayload(int size) {
        StringBuilder sb = new StringBuilder(size);
        for (int i = 0; i < size; i++) {
            sb.append(ALPHABET.charAt(random.nextInt(ALPHABET.length())));
        }
        return sb.toString();
    }

    /**
     * WebSocket server initializer.
     */
    private class WebSocketRttServerInitializer extends ChannelInitializer<SocketChannel> {
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
            pipeline.addLast(new WebSocketRttServerHandler());
        }
    }

    /**
     * WebSocket server handler.
     */
    private class WebSocketRttServerHandler extends SimpleChannelInboundHandler<Object> {
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
                    logger.debug("WebSocket connection established: {}", ctx.channel().id().asShortText());
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
            metrics.put("messages_received", messagesReceived.get());
            metrics.put("messages_sent", messagesSent.get());
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
                // Increment received counter
                messagesReceived.incrementAndGet();
                
                // Parse message as JSON
                RttMessage rttMessage = JsonUtils.fromJson(message, RttMessage.class);
                if (rttMessage == null) {
                    logger.error("Failed to parse message: {}", message);
                    return;
                }
                
                // Add server timestamp
                rttMessage.setServerTimestampNs(TimeUtils.getCurrentTimeNanos());
                
                // If payload is null or empty, generate a random payload of the configured size
                if (rttMessage.getPayload() == null || rttMessage.getPayload().isEmpty()) {
                    rttMessage.setPayload(generateRandomPayload(config.getPayloadSize()));
                }
                
                // Set server send timestamp right before sending
                rttMessage.setServerSendTimestampNs(TimeUtils.getCurrentTimeNanos());
                
                // Send response back
                String response = JsonUtils.toJson(rttMessage);
                ctx.channel().writeAndFlush(new TextWebSocketFrame(response));
                
                // Increment sent counter
                messagesSent.incrementAndGet();
                
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
        }
    }
}