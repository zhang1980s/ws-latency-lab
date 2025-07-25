package xyz.zzhe.wslatency.rtt.client;

/**
 * Configuration for the WebSocket RTT client.
 */
public class RttClientConfig {
    private String serverUrl = "ws://localhost:10443/ws";
    private int testDuration = 30;
    private int requestsPerSecond = 10;
    private int payloadSize = 100;
    private int prewarmCount = 100;
    private boolean insecureSkipVerify = false;
    private boolean continuous = false;

    /**
     * Default constructor with default values.
     */
    public RttClientConfig() {
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
     * @return The test duration
     */
    public int getTestDuration() {
        return testDuration;
    }

    /**
     * Set the test duration in seconds.
     *
     * @param testDuration The test duration
     */
    public void setTestDuration(int testDuration) {
        this.testDuration = testDuration;
    }

    /**
     * Get the number of requests to send per second.
     *
     * @return Requests per second
     */
    public int getRequestsPerSecond() {
        return requestsPerSecond;
    }

    /**
     * Set the number of requests to send per second.
     *
     * @param requestsPerSecond Requests per second
     */
    public void setRequestsPerSecond(int requestsPerSecond) {
        this.requestsPerSecond = requestsPerSecond;
    }

    /**
     * Get the size of the message payload in bytes.
     *
     * @return Payload size in bytes
     */
    public int getPayloadSize() {
        return payloadSize;
    }

    /**
     * Set the size of the message payload in bytes.
     *
     * @param payloadSize Payload size in bytes
     */
    public void setPayloadSize(int payloadSize) {
        this.payloadSize = payloadSize;
    }

    /**
     * Get the number of initial samples to skip for warm-up.
     *
     * @return Prewarm count
     */
    public int getPrewarmCount() {
        return prewarmCount;
    }

    /**
     * Set the number of initial samples to skip for warm-up.
     *
     * @param prewarmCount Prewarm count
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
        return "RttClientConfig{" +
                "serverUrl='" + serverUrl + '\'' +
                ", testDuration=" + testDuration +
                ", requestsPerSecond=" + requestsPerSecond +
                ", payloadSize=" + payloadSize +
                ", prewarmCount=" + prewarmCount +
                ", insecureSkipVerify=" + insecureSkipVerify +
                ", continuous=" + continuous +
                '}';
    }
}