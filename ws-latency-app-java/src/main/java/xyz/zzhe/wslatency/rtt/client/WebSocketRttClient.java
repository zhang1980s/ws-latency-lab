package xyz.zzhe.wslatency.rtt.client;

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
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import xyz.zzhe.wslatency.common.util.JsonUtils;
import xyz.zzhe.wslatency.common.util.TimeUtils;
import xyz.zzhe.wslatency.rtt.model.RttMessage;
import xyz.zzhe.wslatency.stats.LatencyStats;
import xyz.zzhe.wslatency.stats.StatisticsCalculator;

import javax.net.ssl.SSLException;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.Random;
import java.util.UUID;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicLong;

/**
 * WebSocket RTT client for latency testing using Netty.
 * This client implements a request-response model for RTT measurement.
 */
public class WebSocketRttClient {
    private static final Logger logger = LoggerFactory.getLogger(WebSocketRttClient.class);
    private static final String ALPHABET = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";

    private final RttClientConfig config;
    private final StatisticsCalculator rttStatistics;
    private final StatisticsCalculator oneWayLatencyStatistics;
    private final CountDownLatch connectionLatch = new CountDownLatch(1);
    private final CountDownLatch completionLatch = new CountDownLatch(1);
    private final AtomicLong sequence = new AtomicLong(0);
    private final Random random = new Random();
    private final ScheduledExecutorService scheduler = Executors.newScheduledThreadPool(1);
    
    private EventLoopGroup group;
    private Channel channel;
    private WebSocketRttClientHandler clientHandler;

    /**
     * Create a new WebSocket RTT client with the specified configuration.
     *
     * @param config The client configuration
     */
    public WebSocketRttClient(RttClientConfig config) {
        this.config = config;
        this.rttStatistics = new StatisticsCalculator(config.getPrewarmCount());
        this.oneWayLatencyStatistics = new StatisticsCalculator(config.getPrewarmCount());
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
        clientHandler = new WebSocketRttClientHandler(
                WebSocketClientHandshakerFactory.newHandshaker(
                        uri, WebSocketVersion.V13, null, true, new DefaultHttpHeaders()),
                connectionLatch,
                completionLatch,
                rttStatistics,
                oneWayLatencyStatistics
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
        // Stop the scheduler
        scheduler.shutdown();
        try {
            if (!scheduler.awaitTermination(5, TimeUnit.SECONDS)) {
                scheduler.shutdownNow();
            }
        } catch (InterruptedException e) {
            scheduler.shutdownNow();
            Thread.currentThread().interrupt();
        }
        
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
     * Run the RTT latency test.
     *
     * @throws Exception If an error occurs
     */
    public void runTest() throws Exception {
        if (channel == null || !channel.isActive()) {
            throw new IllegalStateException("Not connected to WebSocket server");
        }
        
        logger.info("Starting RTT latency test with prewarm count: {}", config.getPrewarmCount());
        logger.info("Sending {} requests per second with payload size {} bytes", 
                config.getRequestsPerSecond(), config.getPayloadSize());
        
        // Calculate the period in nanoseconds
        long periodNanos = TimeUnit.SECONDS.toNanos(1) / config.getRequestsPerSecond();
        
        // Schedule sending requests at the configured rate
        scheduler.scheduleAtFixedRate(
                this::sendRequest,
                0,
                periodNanos,
                TimeUnit.NANOSECONDS
        );
        
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
        
        // Stop sending requests
        scheduler.shutdown();
        
        // Calculate and display statistics
        LatencyStats rttStats = rttStatistics.calculate();
        LatencyStats oneWayStats = oneWayLatencyStatistics.calculate();
        
        logger.info("Test completed with {} RTT samples and {} one-way latency samples ({} skipped for warm-up)",
                rttStatistics.getSampleCount(), oneWayLatencyStatistics.getSampleCount(),
                rttStatistics.getSkippedCount());
        
        // Output RTT latency metrics
        System.out.println("\n========== RTT LATENCY TEST RESULTS ==========");
        System.out.println(rttStats.toString());
        System.out.println("=============================================\n");
        
        // Output one-way latency metrics
        System.out.println("\n========== ONE-WAY LATENCY TEST RESULTS ==========");
        System.out.println(oneWayStats.toString());
        System.out.println("==================================================\n");
        
    }

    /**
     * Send a request to the server.
     */
    private void sendRequest() {
        if (channel != null && channel.isActive()) {
            try {
                // Create RTT message
                RttMessage message = new RttMessage(sequence.incrementAndGet());
                message.setMessageId(UUID.randomUUID().toString());
                message.setClientSendTimestampNs(TimeUtils.getCurrentTimeNanos());
                message.setPayload(generateRandomPayload(config.getPayloadSize()));
                
                // Convert to JSON
                String json = JsonUtils.toJson(message);
                
                // Send message
                channel.writeAndFlush(new TextWebSocketFrame(json));
            } catch (Exception e) {
                logger.error("Error sending request", e);
            }
        }
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
}