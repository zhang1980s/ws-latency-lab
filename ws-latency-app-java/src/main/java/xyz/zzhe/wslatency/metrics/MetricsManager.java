package xyz.zzhe.wslatency.metrics;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpServer;
import io.micrometer.core.instrument.Counter;
import io.micrometer.core.instrument.Gauge;
import io.micrometer.core.instrument.Timer;
import io.micrometer.prometheus.PrometheusConfig;
import io.micrometer.prometheus.PrometheusMeterRegistry;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import xyz.zzhe.wslatency.stats.LatencyStats;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicLong;

/**
 * Manages metrics collection and exposure for the WebSocket latency testing application.
 */
public class MetricsManager {
    private static final Logger logger = LoggerFactory.getLogger(MetricsManager.class);

    private final MetricsConfig config;
    private final PrometheusMeterRegistry registry;
    private final HttpServer server;

    // Metrics
    private final Timer latencyTimer;
    private final Counter eventsSent;
    private final Counter eventsReceived;
    private final Counter errors;
    private final Map<String, AtomicLong> gaugeValues;

    /**
     * Creates a new metrics manager with the specified configuration.
     *
     * @param config The metrics configuration
     * @throws IOException If the HTTP server cannot be started
     */
    public MetricsManager(MetricsConfig config) throws IOException {
        this.config = config;
        this.registry = new PrometheusMeterRegistry(PrometheusConfig.DEFAULT);
        this.gaugeValues = new HashMap<>();

        // Initialize metrics
        this.latencyTimer = Timer.builder("ws.latency")
                .description("WebSocket one-way latency in microseconds")
                .tag("type", "server_to_client")
                .publishPercentiles(0.5, 0.9, 0.99)
                .register(registry);

        this.eventsSent = Counter.builder("ws.events.sent")
                .description("Total number of events sent")
                .register(registry);

        this.eventsReceived = Counter.builder("ws.events.received")
                .description("Total number of events received")
                .register(registry);

        this.errors = Counter.builder("ws.errors")
                .description("Total number of errors encountered")
                .register(registry);

        // Initialize gauge values
        initGauges();

        // Start HTTP server for metrics
        this.server = HttpServer.create(new InetSocketAddress(config.getPort()), 0);
        this.server.createContext(config.getEndpoint(), new MetricsHandler());
        this.server.setExecutor(null); // Use default executor
        this.server.start();

        logger.info("Metrics server started on port {} at endpoint {}", config.getPort(), config.getEndpoint());
    }

    /**
     * Initialize gauge metrics.
     */
    private void initGauges() {
        // Create atomic values for gauges
        gaugeValues.put("min", new AtomicLong(0));
        gaugeValues.put("max", new AtomicLong(0));
        gaugeValues.put("p50", new AtomicLong(0));
        gaugeValues.put("p90", new AtomicLong(0));
        gaugeValues.put("p99", new AtomicLong(0));
        gaugeValues.put("mean", new AtomicLong(0));

        // Register gauges
        Gauge.builder("ws.latency.min", gaugeValues.get("min"), AtomicLong::get)
                .description("Minimum latency in microseconds")
                .register(registry);

        Gauge.builder("ws.latency.max", gaugeValues.get("max"), AtomicLong::get)
                .description("Maximum latency in microseconds")
                .register(registry);

        Gauge.builder("ws.latency.p50", gaugeValues.get("p50"), AtomicLong::get)
                .description("P50 (median) latency in microseconds")
                .register(registry);

        Gauge.builder("ws.latency.p90", gaugeValues.get("p90"), AtomicLong::get)
                .description("P90 latency in microseconds")
                .register(registry);

        Gauge.builder("ws.latency.p99", gaugeValues.get("p99"), AtomicLong::get)
                .description("P99 latency in microseconds")
                .register(registry);

        Gauge.builder("ws.latency.mean", gaugeValues.get("mean"), AtomicLong::get)
                .description("Mean latency in microseconds")
                .register(registry);
    }

    /**
     * Record a latency value.
     *
     * @param latencyUs Latency in microseconds
     */
    public void recordLatency(long latencyUs) {
        latencyTimer.record(latencyUs, TimeUnit.MICROSECONDS);
    }

    /**
     * Update statistics based on calculated values.
     *
     * @param stats The latency statistics
     */
    public void updateStatistics(LatencyStats stats) {
        gaugeValues.get("min").set(stats.getMin());
        gaugeValues.get("max").set(stats.getMax());
        gaugeValues.get("p50").set(stats.getP50());
        gaugeValues.get("p90").set(stats.getP90());
        gaugeValues.get("p99").set(stats.getP99());
        gaugeValues.get("mean").set((long) stats.getMean());
    }

    /**
     * Increment the events sent counter.
     */
    public void incrementEventsSent() {
        eventsSent.increment();
    }
    
    /**
     * Increment the events sent counter by the specified amount.
     *
     * @param count The number of events sent
     */
    public void incrementEventsSent(int count) {
        eventsSent.increment(count);
    }

    /**
     * Increment the events received counter.
     */
    public void incrementEventsReceived() {
        eventsReceived.increment();
    }

    /**
     * Increment the errors counter.
     */
    public void incrementErrors() {
        errors.increment();
    }

    /**
     * Stop the metrics server.
     */
    public void stop() {
        if (server != null) {
            server.stop(0);
            logger.info("Metrics server stopped");
        }
    }

    /**
     * HTTP handler for exposing Prometheus metrics.
     */
    private class MetricsHandler implements HttpHandler {
        @Override
        public void handle(HttpExchange exchange) throws IOException {
            String response = registry.scrape();
            exchange.getResponseHeaders().set("Content-Type", "text/plain; charset=UTF-8");
            exchange.sendResponseHeaders(200, response.getBytes().length);
            try (OutputStream os = exchange.getResponseBody()) {
                os.write(response.getBytes());
            }
        }
    }
}