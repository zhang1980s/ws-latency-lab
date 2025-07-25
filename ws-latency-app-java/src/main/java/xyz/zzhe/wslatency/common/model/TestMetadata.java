package xyz.zzhe.wslatency.common.model;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * Represents test metadata for latency measurements.
 */
public class TestMetadata {
    @JsonProperty("server_send_ts_ns")
    private long serverSendTimestampNs;
    
    @JsonProperty("client_recv_ts_ns")
    private long clientReceiveTimestampNs;

    /**
     * Default constructor.
     */
    public TestMetadata() {
        // Default constructor for Jackson
        this.serverSendTimestampNs = 0;
        this.clientReceiveTimestampNs = 0;
    }

    /**
     * Get the server send timestamp in nanoseconds.
     *
     * @return The server send timestamp
     */
    public long getServerSendTimestampNs() {
        return serverSendTimestampNs;
    }

    /**
     * Set the server send timestamp in nanoseconds.
     *
     * @param serverSendTimestampNs The server send timestamp
     */
    public void setServerSendTimestampNs(long serverSendTimestampNs) {
        this.serverSendTimestampNs = serverSendTimestampNs;
    }

    /**
     * Get the client receive timestamp in nanoseconds.
     *
     * @return The client receive timestamp
     */
    public long getClientReceiveTimestampNs() {
        return clientReceiveTimestampNs;
    }

    /**
     * Set the client receive timestamp in nanoseconds.
     *
     * @param clientReceiveTimestampNs The client receive timestamp
     */
    public void setClientReceiveTimestampNs(long clientReceiveTimestampNs) {
        this.clientReceiveTimestampNs = clientReceiveTimestampNs;
    }

    /**
     * Calculate the one-way latency in nanoseconds.
     *
     * @return The latency in nanoseconds
     */
    public long calculateLatencyNs() {
        // Calculate the difference between client receive time and server send time
        return clientReceiveTimestampNs - serverSendTimestampNs;
    }
    
    /**
     * For backward compatibility - calculate the one-way latency in microseconds.
     *
     * @return The latency in microseconds
     */
    public long calculateLatencyUs() {
        return calculateLatencyNs() / 1000;
    }

    @Override
    public String toString() {
        return "TestMetadata{" +
                "serverSendTimestampNs=" + serverSendTimestampNs +
                ", clientReceiveTimestampNs=" + clientReceiveTimestampNs +
                '}';
    }
}