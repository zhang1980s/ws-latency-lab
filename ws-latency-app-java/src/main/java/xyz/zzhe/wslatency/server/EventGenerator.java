package xyz.zzhe.wslatency.server;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import xyz.zzhe.wslatency.common.model.Event;
import xyz.zzhe.wslatency.common.model.EventData;
import xyz.zzhe.wslatency.common.model.TestMetadata;
import xyz.zzhe.wslatency.common.util.JsonUtils;
import xyz.zzhe.wslatency.common.util.TimeUtils;

import io.netty.channel.Channel;
import io.netty.channel.group.ChannelGroup;
import io.netty.channel.group.DefaultChannelGroup;
import io.netty.handler.codec.http.websocketx.TextWebSocketFrame;
import io.netty.util.concurrent.GlobalEventExecutor;

import java.util.Random;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicLong;

/**
 * Generates and sends events to WebSocket clients at a configurable rate.
 */
public class EventGenerator {
    private static final Logger logger = LoggerFactory.getLogger(EventGenerator.class);
    private static final String EVENT_TYPE = "latency_test";
    private static final String ALPHABET = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";

    private final ServerConfig config;
    private final ChannelGroup clients = new DefaultChannelGroup(GlobalEventExecutor.INSTANCE);
    private final ScheduledExecutorService scheduler;
    private final AtomicLong sequence = new AtomicLong(0);
    private final AtomicLong eventsSent = new AtomicLong(0);
    private final Random random = new Random();

    /**
     * Create a new event generator with the specified configuration.
     *
     * @param config The server configuration
     */
    public EventGenerator(ServerConfig config) {
        this.config = config;
        this.scheduler = Executors.newScheduledThreadPool(1);
    }

    /**
     * Start generating and sending events.
     */
    public void start() {
        // Calculate the period in nanoseconds
        long periodNanos = TimeUnit.SECONDS.toNanos(1) / config.getEventsPerSecond();
        
        // Schedule event generation at the configured rate
        scheduler.scheduleAtFixedRate(
                this::generateAndSendEvent,
                0,
                periodNanos,
                TimeUnit.NANOSECONDS
        );
        
        logger.info("Event generator started with rate {} events/sec (period {} ns)",
                config.getEventsPerSecond(), periodNanos);
    }

    /**
     * Stop generating events.
     */
    public void stop() {
        scheduler.shutdown();
        try {
            if (!scheduler.awaitTermination(5, TimeUnit.SECONDS)) {
                scheduler.shutdownNow();
            }
        } catch (InterruptedException e) {
            scheduler.shutdownNow();
            Thread.currentThread().interrupt();
        }
        logger.info("Event generator stopped");
    }

    /**
     * Add a client to receive events.
     *
     * @param channel The Netty channel
     */
    public void addClient(Channel channel) {
        clients.add(channel);
        logger.debug("Client added: {}, total clients: {}", channel.id().asShortText(), clients.size());
    }

    /**
     * Remove a client.
     *
     * @param channel The Netty channel
     */
    public void removeClient(Channel channel) {
        clients.remove(channel);
        logger.debug("Client removed: {}, total clients: {}", channel.id().asShortText(), clients.size());
    }

    /**
     * Generate and send an event to all connected clients.
     */
    private void generateAndSendEvent() {
        if (clients.isEmpty()) {
            return;
        }

        try {
            // Create event
            Event event = createEvent();
            
            // Convert to JSON
            String message = JsonUtils.toJson(event);
            if (message == null) {
                logger.error("Failed to convert event to JSON");
                return;
            }
            
            // Create a WebSocket text frame
            TextWebSocketFrame frame = new TextWebSocketFrame(message);
            
            // Send to all clients
            clients.writeAndFlush(frame);
            eventsSent.addAndGet(clients.size());
        } catch (Exception e) {
            logger.error("Error generating event", e);
        }
    }

    /**
     * Create a new event.
     *
     * @return The event
     */
    private Event createEvent() {
        // Create event with next sequence number
        Event event = new Event(EVENT_TYPE, sequence.incrementAndGet());
        
        // Create event data with random payload
        EventData data = new EventData(generateRandomPayload(config.getPayloadSize()));
        event.setData(data);
        
        // Create test metadata with server timestamp
        TestMetadata metadata = new TestMetadata();
        metadata.setServerSendTimestampNs(TimeUtils.getCurrentTimeNanos());
        event.setTestMetadata(metadata);
        
        return event;
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
     * Get the total number of events sent.
     *
     * @return Number of events sent
     */
    public long getEventsSent() {
        return eventsSent.get();
    }
    
    /**
     * Get the number of connected clients.
     *
     * @return Number of connected clients
     */
    public int getClientCount() {
        return clients.size();
    }
}