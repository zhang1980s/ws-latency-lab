package xyz.zzhe.wslatency.metrics;

/**
 * Configuration for the metrics component.
 */
public class MetricsConfig {
    private int port = 9091;
    private String endpoint = "/metrics";

    /**
     * Constructor with default endpoint.
     *
     * @param port The port to expose metrics on
     */
    public MetricsConfig(int port) {
        this.port = port;
    }

    /**
     * Constructor with custom port and endpoint.
     *
     * @param port     The port to expose metrics on
     * @param endpoint The endpoint path for metrics
     */
    public MetricsConfig(int port, String endpoint) {
        this.port = port;
        this.endpoint = endpoint;
    }

    /**
     * Get the port for exposing metrics.
     *
     * @return The port number
     */
    public int getPort() {
        return port;
    }

    /**
     * Set the port for exposing metrics.
     *
     * @param port The port number
     */
    public void setPort(int port) {
        this.port = port;
    }

    /**
     * Get the endpoint path for metrics.
     *
     * @return The endpoint path
     */
    public String getEndpoint() {
        return endpoint;
    }

    /**
     * Set the endpoint path for metrics.
     *
     * @param endpoint The endpoint path
     */
    public void setEndpoint(String endpoint) {
        this.endpoint = endpoint;
    }

    @Override
    public String toString() {
        return "MetricsConfig{" +
                "port=" + port +
                ", endpoint='" + endpoint + '\'' +
                '}';
    }
}