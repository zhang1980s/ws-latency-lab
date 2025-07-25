package xyz.zzhe.wslatency;

import picocli.CommandLine;
import picocli.CommandLine.Command;
import picocli.CommandLine.Option;
import xyz.zzhe.wslatency.rtt.WebSocketRttApp;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.concurrent.Callable;

/**
 * Main application entry point for WebSocket Testing Tool.
 * Dispatches to the appropriate application based on the application type.
 */
@Command(
    name = "ws-latency-app",
    mixinStandardHelpOptions = true,
    version = "WebSocket Latency App 1.0.0",
    description = "WebSocket Latency Testing Tool"
)
public class WebSocketApp implements Callable<Integer> {
    private static final Logger logger = LoggerFactory.getLogger(WebSocketApp.class);

    @Option(
        names = {"-a", "--app-type"},
        description = "Application type: 'push' for server-push model or 'rtt' for request-response model"
    )
    private String appType = "push";

    // Forward all other arguments to the specific application
    @Option(
        names = {"-m", "--mode"},
        description = "Mode to run: 'server' or 'client'"
    )
    private String mode;

    // Server options
    @Option(
        names = {"-p", "--port"},
        description = "Port for server to listen on"
    )
    private Integer port;

    @Option(
        names = {"-r", "--rate"},
        description = "Events/requests per second"
    )
    private Integer rate;

    // Client options
    @Option(
        names = {"-s", "--server"},
        description = "WebSocket server address"
    )
    private String serverUrl;

    @Option(
        names = {"-d", "--duration"},
        description = "Test duration in seconds"
    )
    private Integer testDuration;

    @Option(
        names = {"--payload-size"},
        description = "Size of the message payload in bytes"
    )
    private Integer payloadSize;

    @Option(
        names = {"--prewarm-count"},
        description = "Skip calculating latency for first N events"
    )
    private Integer prewarmCount;

    @Option(
        names = {"--insecure"},
        description = "Skip TLS certificate verification"
    )
    private Boolean insecureSkipVerify;

    // Common options
    @Option(
        names = {"--continuous"},
        description = "Run in continuous monitoring mode"
    )
    private Boolean continuous;

    /**
     * Main entry point for the application.
     *
     * @param args Command line arguments
     */
    public static void main(String[] args) {
        int exitCode = new CommandLine(new WebSocketApp()).execute(args);
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
            // Build the arguments array for the specific application
            String[] appArgs = buildAppArgs();

            // Dispatch to the appropriate application based on the app type
            if ("push".equalsIgnoreCase(appType)) {
                logger.info("Starting server-push model application");
                return new CommandLine(new WebSocketLatencyApp()).execute(appArgs);
            } else if ("rtt".equalsIgnoreCase(appType)) {
                logger.info("Starting request-response model application");
                return new CommandLine(new WebSocketRttApp()).execute(appArgs);
            } else {
                logger.error("Invalid application type: {}. Must be 'push' or 'rtt'", appType);
                return 1;
            }
        } catch (Exception e) {
            logger.error("Error running application", e);
            return 1;
        }
    }

    /**
     * Build the arguments array for the specific application.
     *
     * @return The arguments array
     */
    private String[] buildAppArgs() {
        // Create a StringBuilder to build the arguments
        StringBuilder argsBuilder = new StringBuilder();

        // Add mode
        if (mode != null) {
            argsBuilder.append("-m=").append(mode).append(" ");
        }

        // Add server options
        if (port != null) {
            argsBuilder.append("-p=").append(port).append(" ");
        }
        if (rate != null) {
            argsBuilder.append("-r=").append(rate).append(" ");
        }

        // Add client options
        if (serverUrl != null) {
            argsBuilder.append("-s=").append(serverUrl).append(" ");
        }
        if (testDuration != null) {
            argsBuilder.append("-d=").append(testDuration).append(" ");
        }
        if (payloadSize != null) {
            argsBuilder.append("--payload-size=").append(payloadSize).append(" ");
        }
        if (prewarmCount != null) {
            argsBuilder.append("--prewarm-count=").append(prewarmCount).append(" ");
        }
        if (insecureSkipVerify != null && insecureSkipVerify) {
            argsBuilder.append("--insecure ");
        }

        // Add common options
        if (continuous != null && continuous) {
            argsBuilder.append("--continuous ");
        }

        // Split the arguments string into an array
        return argsBuilder.toString().trim().split("\\s+");
    }
}