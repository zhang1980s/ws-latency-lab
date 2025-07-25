package xyz.zzhe.wslatency.client;

/**
 * Configuration for the WebSocket client.
 */
public class ClientConfig {
    private String serverUrl = "ws://localhost:10443/ws";
    private int testDuration = 30;
    private int prewarmCount = 100;
    private boolean insecureSkipVerify = false;
    private boolean continuous = false;

    /**
     * Default constructor with default values.
     */
    public ClientConfig() {
        // Default values set in field declarations
    }

    /**
     * Get the WebSocket server URL.
     *
     * @return The server URL
     */
    public String getServerUrl() {
        return serverUrl;
    }

    /**
     * Set the WebSocket server URL.
     *
     * @param serverUrl The server URL
     */
    public void setServerUrl(String serverUrl) {
        this.serverUrl = serverUrl;
    }

    /**
     * Get the test duration in seconds.
     *
     * @return Test duration in seconds
     */
    public int getTestDuration() {
        return testDuration;
    }

    /**
     * Set the test duration in seconds.
     *
     * @param testDuration Test duration in seconds
     */
    public void setTestDuration(int testDuration) {
        this.testDuration = testDuration;
    }

    /**
     * Get the number of events to skip for warm-up.
     *
     * @return Number of events to skip
     */
    public int getPrewarmCount() {
        return prewarmCount;
    }

    /**
     * Set the number of events to skip for warm-up.
     *
     * @param prewarmCount Number of events to skip
     */
    public void setPrewarmCount(int prewarmCount) {
        this.prewarmCount = prewarmCount;
    }

    /**
     * Check if TLS certificate verification should be skipped.
     *
     * @return true to skip verification, false otherwise
     */
    public boolean isInsecureSkipVerify() {
        return insecureSkipVerify;
    }

    /**
     * Set whether to skip TLS certificate verification.
     *
     * @param insecureSkipVerify true to skip verification, false otherwise
     */
    public void setInsecureSkipVerify(boolean insecureSkipVerify) {
        this.insecureSkipVerify = insecureSkipVerify;
    }

    /**
     * Check if the client should run in continuous mode.
     *
     * @return true for continuous mode, false otherwise
     */
    public boolean isContinuous() {
        return continuous;
    }

    /**
     * Set whether the client should run in continuous mode.
     *
     * @param continuous true for continuous mode, false otherwise
     */
    public void setContinuous(boolean continuous) {
        this.continuous = continuous;
    }


    @Override
    public String toString() {
        return "ClientConfig{" +
                "serverUrl='" + serverUrl + '\'' +
                ", testDuration=" + testDuration +
                ", prewarmCount=" + prewarmCount +
                ", insecureSkipVerify=" + insecureSkipVerify +
                ", continuous=" + continuous +
                '}';
    }
}