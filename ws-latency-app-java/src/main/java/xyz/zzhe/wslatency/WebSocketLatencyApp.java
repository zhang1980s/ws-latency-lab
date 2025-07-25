package xyz.zzhe.wslatency;

import picocli.CommandLine;
import picocli.CommandLine.Command;
import picocli.CommandLine.Option;
import xyz.zzhe.wslatency.client.ClientConfig;
import xyz.zzhe.wslatency.client.WebSocketClient;
import xyz.zzhe.wslatency.server.ServerConfig;
import xyz.zzhe.wslatency.server.WebSocketServer;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.concurrent.Callable;

/**
 * Main application entry point for WebSocket Latency Testing Tool.
 * Supports both server and client modes with configurable parameters.
 */
@Command(
    name = "ws-latency-test",
    mixinStandardHelpOptions = true,
    version = "WebSocket Latency Test 1.0.0",
    description = "WebSocket Latency Testing Tool with server-push model"
)
public class WebSocketLatencyApp implements Callable<Integer> {
    private static final Logger logger = LoggerFactory.getLogger(WebSocketLatencyApp.class);

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

    @Option(
        names = {"-r", "--rate"},
        description = "Events per second for server to send"
    )
    private int eventsPerSecond = 10;

    // Client options
    @Option(
        names = {"-s", "--server"},
        description = "WebSocket server address"
    )
    private String serverUrl = "ws://localhost:10443/ws";

    @Option(
        names = {"-d", "--duration"},
        description = "Test duration in seconds"
    )
    private int testDuration = 30;

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
        int exitCode = new CommandLine(new WebSocketLatencyApp()).execute(args);
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
        logger.info("Starting server on port {} with event rate {} events/sec", port, eventsPerSecond);

        // Create server configuration
        ServerConfig config = new ServerConfig();
        config.setPort(port);
        config.setEventsPerSecond(eventsPerSecond);

        // Create and start server
        WebSocketServer server = new WebSocketServer(config);
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
        logger.info("Starting client connecting to {} for {} seconds", serverUrl,
                    continuous ? "continuous monitoring" : testDuration);

        // Create client configuration
        ClientConfig config = new ClientConfig();
        config.setServerUrl(serverUrl);
        config.setTestDuration(testDuration);
        config.setPrewarmCount(prewarmCount);
        config.setInsecureSkipVerify(insecureSkipVerify);
        config.setContinuous(continuous);

        // Create client
        WebSocketClient client = new WebSocketClient(config);

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