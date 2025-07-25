package xyz.zzhe.wslatency.rtt;

import picocli.CommandLine;
import picocli.CommandLine.Command;
import picocli.CommandLine.Option;
import xyz.zzhe.wslatency.rtt.client.RttClientConfig;
import xyz.zzhe.wslatency.rtt.client.WebSocketRttClient;
import xyz.zzhe.wslatency.rtt.server.RttServerConfig;
import xyz.zzhe.wslatency.rtt.server.WebSocketRttServer;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.concurrent.Callable;

/**
 * Main application entry point for WebSocket RTT Latency Testing Tool.
 * This implementation uses a request-response model to measure Round-Trip Time (RTT).
 */
@Command(
    name = "ws-rtt-test",
    mixinStandardHelpOptions = true,
    version = "WebSocket RTT Test 1.0.0",
    description = "WebSocket RTT Testing Tool with request-response model"
)
public class WebSocketRttApp implements Callable<Integer> {
    private static final Logger logger = LoggerFactory.getLogger(WebSocketRttApp.class);

    @Option(
        names = {"-m", "--mode"},
        required = true,
        description = "Mode to run: 'server' or 'client'"
    )
    private String mode;

    // Server options
    @Option(
        names = {"-p", "--port"},
        description = "Port for server to listen on"
    )
    private int port = 10443;

    // Client options
    @Option(
        names = {"-s", "--server"},
        description = "WebSocket server address"
    )
    private String serverUrl = "ws://localhost:10443/ws";

    @Option(
        names = {"-r", "--rate"},
        description = "Requests per second for client to send"
    )
    private int requestsPerSecond = 10;

    @Option(
        names = {"-d", "--duration"},
        description = "Test duration in seconds"
    )
    private int testDuration = 30;

    @Option(
        names = {"--payload-size"},
        description = "Size of the message payload in bytes"
    )
    private int payloadSize = 100;

    @Option(
        names = {"--prewarm-count"},
        description = "Skip calculating latency for first N events"
    )
    private int prewarmCount = 100;

    @Option(
        names = {"--insecure"},
        description = "Skip TLS certificate verification"
    )
    private boolean insecureSkipVerify = false;

    // Common options
    @Option(
        names = {"--continuous"},
        description = "Run in continuous monitoring mode"
    )
    private boolean continuous = false;


    /**
     * Main entry point for the application.
     *
     * @param args Command line arguments
     */
    public static void main(String[] args) {
        int exitCode = new CommandLine(new WebSocketRttApp()).execute(args);
        System.exit(exitCode);
    }

    /**
     * Executes the application based on the provided command line arguments.
     *
     * @return Exit code (0 for success, non-zero for errors)
     */
    @Override
    public Integer call() {
        try {
            // Run in server or client mode based on options
            if ("server".equalsIgnoreCase(mode)) {
                runServer();
            } else if ("client".equalsIgnoreCase(mode)) {
                runClient();
            } else {
                logger.error("Invalid mode: {}. Must be 'server' or 'client'", mode);
                return 1;
            }

            return 0;
        } catch (Exception e) {
            logger.error("Error running application", e);
            return 1;
        }
    }

    /**
     * Runs the application in server mode.
     */
    private void runServer() {
        logger.info("Starting RTT server on port {}", port);

        // Create server configuration
        RttServerConfig config = new RttServerConfig();
        config.setPort(port);
        config.setPayloadSize(payloadSize);

        // Create and start server
        WebSocketRttServer server = new WebSocketRttServer(config);
        server.start();

        // Keep the server running
        try {
            // Wait for the server to be stopped
            Thread.currentThread().join();
        } catch (InterruptedException e) {
            logger.info("Server interrupted, shutting down");
        } finally {
            server.stop();
        }
    }

    /**
     * Runs the application in client mode.
     */
    private void runClient() {
        logger.info("Starting RTT client connecting to {} for {} seconds with rate {} req/sec",
                serverUrl, continuous ? "continuous monitoring" : testDuration, requestsPerSecond);

        // Create client configuration
        RttClientConfig config = new RttClientConfig();
        config.setServerUrl(serverUrl);
        config.setTestDuration(testDuration);
        config.setRequestsPerSecond(requestsPerSecond);
        config.setPayloadSize(payloadSize);
        config.setPrewarmCount(prewarmCount);
        config.setInsecureSkipVerify(insecureSkipVerify);
        config.setContinuous(continuous);

        // Create client
        WebSocketRttClient client = new WebSocketRttClient(config);

        try {
            // Connect to server
            client.connect();

            // Run test
            client.runTest();
        } catch (Exception e) {
            logger.error("Error running client", e);
        } finally {
            client.disconnect();
        }
    }
}