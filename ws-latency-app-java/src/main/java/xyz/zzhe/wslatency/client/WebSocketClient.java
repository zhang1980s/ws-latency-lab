package xyz.zzhe.wslatency.client;

import io.netty.bootstrap.Bootstrap;
import io.netty.channel.*;
import io.netty.channel.epoll.EpollEventLoopGroup;
import io.netty.channel.epoll.EpollSocketChannel;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioSocketChannel;
import io.netty.handler.codec.http.DefaultHttpHeaders;
import io.netty.handler.codec.http.HttpClientCodec;
import io.netty.handler.codec.http.HttpObjectAggregator;
import io.netty.handler.codec.http.websocketx.*;
import io.netty.handler.ssl.SslContext;
import io.netty.handler.ssl.SslContextBuilder;
import io.netty.handler.ssl.util.InsecureTrustManagerFactory;
import io.netty.handler.timeout.IdleStateHandler;
import io.netty.util.internal.PlatformDependent;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import xyz.zzhe.wslatency.stats.LatencyStats;
import xyz.zzhe.wslatency.stats.StatisticsCalculator;

import javax.net.ssl.SSLException;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;

/**
 * WebSocket client for latency testing using Netty.
 */
public class WebSocketClient {
    private static final Logger logger = LoggerFactory.getLogger(WebSocketClient.class);

    private final ClientConfig config;
    private final StatisticsCalculator statistics;
    private final CountDownLatch connectionLatch = new CountDownLatch(1);
    private final CountDownLatch completionLatch = new CountDownLatch(1);
    
    private EventLoopGroup group;
    private Channel channel;
    private WebSocketClientHandler clientHandler;

    /**
     * Create a new WebSocket client with the specified configuration.
     *
     * @param config The client configuration
     */
    public WebSocketClient(ClientConfig config) {
        this.config = config;
        this.statistics = new StatisticsCalculator(config.getPrewarmCount());
    }

    /**
     * Connect to the WebSocket server.
     *
     * @throws Exception If an error occurs
     */
    public void connect() throws Exception {
        logger.info("Connecting to WebSocket server: {}", config.getServerUrl());
        
        URI uri = new URI(config.getServerUrl());
        String scheme = uri.getScheme();
        final String host = uri.getHost();
        final int port;
        
        if (uri.getPort() == -1) {
            if ("ws".equalsIgnoreCase(scheme)) {
                port = 80;
            } else if ("wss".equalsIgnoreCase(scheme)) {
                port = 443;
            } else {
                port = 80;
            }
        } else {
            port = uri.getPort();
        }

        // Configure SSL if needed
        final SslContext sslCtx;
        if ("wss".equalsIgnoreCase(scheme)) {
            if (config.isInsecureSkipVerify()) {
                sslCtx = SslContextBuilder.forClient()
                        .trustManager(InsecureTrustManagerFactory.INSTANCE)
                        .build();
                logger.warn("Using insecure SSL context that skips certificate verification");
            } else {
                sslCtx = SslContextBuilder.forClient().build();
            }
        } else {
            sslCtx = null;
        }

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
            group = new EpollEventLoopGroup();
        } else {
            logger.info("Using NIO transport");
            group = new NioEventLoopGroup();
        }

        // Initialize the client handler
        clientHandler = new WebSocketClientHandler(
                WebSocketClientHandshakerFactory.newHandshaker(
                        uri, WebSocketVersion.V13, null, true, new DefaultHttpHeaders()),
                connectionLatch,
                completionLatch,
                statistics
        );

        // Configure bootstrap
        Bootstrap bootstrap = new Bootstrap();
        bootstrap.group(group)
                .channel(useEpoll ? EpollSocketChannel.class : NioSocketChannel.class)
                .option(ChannelOption.TCP_NODELAY, true)
                .option(ChannelOption.SO_KEEPALIVE, true)
                .handler(new ChannelInitializer<SocketChannel>() {
                    @Override
                    protected void initChannel(SocketChannel ch) {
                        ChannelPipeline pipeline = ch.pipeline();
                        
                        // Add SSL handler if needed
                        if (sslCtx != null) {
                            pipeline.addLast(sslCtx.newHandler(ch.alloc(), host, port));
                        }
                        
                        // Add idle state handler to detect and close idle connections
                        pipeline.addLast(new IdleStateHandler(0, 0, 300, TimeUnit.SECONDS));
                        
                        // HTTP client codec
                        pipeline.addLast(new HttpClientCodec());
                        
                        // Aggregate HTTP messages
                        pipeline.addLast(new HttpObjectAggregator(65536));
                        
                        // WebSocket client handler
                        pipeline.addLast(clientHandler);
                    }
                });

        // Connect to server
        channel = bootstrap.connect(uri.getHost(), port).sync().channel();
        
        // Wait for handshake to complete
        clientHandler.handshakeFuture().sync();
        
        // Wait for connection to be established
        try {
            if (!connectionLatch.await(10, TimeUnit.SECONDS)) {
                throw new Exception("Timeout waiting for connection");
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw new Exception("Connection interrupted", e);
        }
        
        logger.info("Connected to WebSocket server");
    }

    /**
     * Disconnect from the WebSocket server.
     */
    public void disconnect() {
        if (channel != null) {
            try {
                // Close WebSocket connection
                channel.writeAndFlush(new CloseWebSocketFrame());
                
                // Wait for the server to close the connection
                channel.closeFuture().sync();
                
                logger.info("Disconnected from WebSocket server");
            } catch (InterruptedException e) {
                logger.error("Error closing WebSocket connection", e);
                Thread.currentThread().interrupt();
            } finally {
                // Shut down event loop group
                if (group != null) {
                    group.shutdownGracefully();
                }
            }
        }
    }

    /**
     * Run the latency test.
     *
     * @throws Exception If an error occurs
     */
    public void runTest() throws Exception {
        if (channel == null || !channel.isActive()) {
            throw new IllegalStateException("Not connected to WebSocket server");
        }
        
        logger.info("Starting latency test with prewarm count: {}", config.getPrewarmCount());
        
        try {
            if (config.isContinuous()) {
                // Run in continuous mode
                logger.info("Running in continuous mode");
                
                // Wait indefinitely
                completionLatch.await();
            } else {
                // Run for specified duration
                logger.info("Running for {} seconds", config.getTestDuration());
                
                // Wait for test duration
                if (!completionLatch.await(config.getTestDuration(), TimeUnit.SECONDS)) {
                    logger.info("Test duration completed");
                }
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            logger.info("Test interrupted");
        }
        
        // Calculate and display statistics
        LatencyStats stats = statistics.calculate();
        logger.info("Test completed with {} samples ({} skipped for warm-up)",
                statistics.getSampleCount(), statistics.getSkippedCount());
        
        // Output latency metrics directly
        System.out.println("\n========== LATENCY TEST RESULTS ==========");
        System.out.println(stats.toString());
        System.out.println("==========================================\n");
        
    }
}