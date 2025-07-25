package xyz.zzhe.wslatency.client;

import io.netty.channel.*;
import io.netty.handler.codec.http.FullHttpResponse;
import io.netty.handler.codec.http.websocketx.*;
import io.netty.util.CharsetUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import xyz.zzhe.wslatency.common.model.Event;
import xyz.zzhe.wslatency.common.model.TestMetadata;
import xyz.zzhe.wslatency.common.util.JsonUtils;
import xyz.zzhe.wslatency.common.util.TimeUtils;
import xyz.zzhe.wslatency.stats.StatisticsCalculator;

import java.util.concurrent.CountDownLatch;
import java.util.concurrent.atomic.AtomicLong;

/**
 * WebSocket client handler for latency testing.
 */
public class WebSocketClientHandler extends SimpleChannelInboundHandler<Object> {
    private static final Logger logger = LoggerFactory.getLogger(WebSocketClientHandler.class);

    private final WebSocketClientHandshaker handshaker;
    private final CountDownLatch connectionLatch;
    private final CountDownLatch completionLatch;
    private final StatisticsCalculator statistics;
    private final AtomicLong messagesReceived = new AtomicLong(0);
    
    private ChannelPromise handshakeFuture;

    /**
     * Create a new WebSocket client handler.
     *
     * @param handshaker      The WebSocket handshaker
     * @param connectionLatch Connection latch to count down when connected
     * @param completionLatch Completion latch to count down when test is complete
     * @param statistics      Statistics calculator
     */
    public WebSocketClientHandler(WebSocketClientHandshaker handshaker,
                                 CountDownLatch connectionLatch,
                                 CountDownLatch completionLatch,
                                 StatisticsCalculator statistics) {
        this.handshaker = handshaker;
        this.connectionLatch = connectionLatch;
        this.completionLatch = completionLatch;
        this.statistics = statistics;
    }

    /**
     * Get the handshake future.
     *
     * @return The handshake future
     */
    public ChannelFuture handshakeFuture() {
        return handshakeFuture;
    }

    @Override
    public void handlerAdded(ChannelHandlerContext ctx) {
        handshakeFuture = ctx.newPromise();
    }

    @Override
    public void channelActive(ChannelHandlerContext ctx) {
        handshaker.handshake(ctx.channel());
    }

    @Override
    public void channelInactive(ChannelHandlerContext ctx) {
        logger.info("WebSocket connection closed");
        completionLatch.countDown();
    }

    @Override
    protected void channelRead0(ChannelHandlerContext ctx, Object msg) {
        Channel ch = ctx.channel();

        // Handle handshake completion
        if (!handshaker.isHandshakeComplete()) {
            try {
                handshaker.finishHandshake(ch, (FullHttpResponse) msg);
                logger.info("WebSocket handshake completed");
                handshakeFuture.setSuccess();
                
                // Configure channel for low latency
                configureForLowLatency(ch);
                
                // Signal that connection is established
                connectionLatch.countDown();
            } catch (WebSocketHandshakeException e) {
                logger.error("WebSocket handshake failed", e);
                handshakeFuture.setFailure(e);
            }
            return;
        }

        // Handle WebSocket frames
        if (msg instanceof TextWebSocketFrame) {
            handleTextFrame((TextWebSocketFrame) msg);
        } else if (msg instanceof PongWebSocketFrame) {
            logger.debug("Received pong");
        } else if (msg instanceof CloseWebSocketFrame) {
            logger.info("Received close frame");
            ch.close();
        }
    }

    /**
     * Handle a text WebSocket frame.
     *
     * @param frame The text frame
     */
    private void handleTextFrame(TextWebSocketFrame frame) {
        try {
            // Record receive time
            long receiveTime = TimeUtils.getCurrentTimeNanos();
            
            // Parse message
            String message = frame.text();
            Event event = JsonUtils.fromJson(message, Event.class);
            if (event == null) {
                logger.error("Failed to parse message: {}", message);
                return;
            }
            
            // Update receive timestamp
            TestMetadata metadata = event.getTestMetadata();
            metadata.setClientReceiveTimestampNs(receiveTime);
            
            // Calculate latency
            long latency = metadata.calculateLatencyNs();
            
            // Add to statistics
            boolean added = statistics.addSample(latency);
            
            // Log periodically
            long count = messagesReceived.incrementAndGet();
            if (count % 100 == 0) {
                logger.debug("Received {} messages, last latency: {} µs", count, latency);
            }
        } catch (Exception e) {
            logger.error("Error processing message", e);
        }
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
        logger.error("Error on WebSocket connection: {}", cause.getMessage());
        
        // Signal that test is complete
        completionLatch.countDown();
        
        if (!handshakeFuture.isDone()) {
            handshakeFuture.setFailure(cause);
        }
        
        ctx.close();
    }

    /**
     * Configure the channel for low latency.
     *
     * @param channel The channel
     */
    private void configureForLowLatency(Channel channel) {
        try {
            // Set TCP_NODELAY option to disable Nagle's algorithm
            channel.config().setOption(ChannelOption.TCP_NODELAY, true);
            
            // Set send/receive buffer sizes
            channel.config().setOption(ChannelOption.SO_SNDBUF, 64 * 1024);
            channel.config().setOption(ChannelOption.SO_RCVBUF, 64 * 1024);
            
            logger.debug("Configured channel for low latency: TCP_NODELAY=true");
        } catch (Exception e) {
            logger.warn("Error configuring channel for low latency", e);
        }
    }
}