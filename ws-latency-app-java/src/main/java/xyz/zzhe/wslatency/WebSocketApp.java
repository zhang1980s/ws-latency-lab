package xyz.zzhe.wslatency;

import picocli.CommandLine;
import xyz.zzhe.wslatency.rtt.WebSocketRttApp;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Main application entry point for WebSocket Latency Testing Tool.
 * This is a simplified version that only supports RTT (Round-Trip Time) measurement.
 */
public class WebSocketApp {
    private static final Logger logger = LoggerFactory.getLogger(WebSocketApp.class);

    /**
     * Main entry point for the application.
     *
     * @param args Command line arguments
     */
    public static void main(String[] args) {
        logger.info("Starting WebSocket RTT measurement application");
        int exitCode = new CommandLine(new WebSocketRttApp()).execute(args);
        System.exit(exitCode);
    }
}