package xyz.zzhe.wslatency.rtt.model;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * Represents a WebSocket message for RTT (Round-Trip Time) testing.
 */
public class RttMessage {
    @JsonProperty("message_id")
    private String messageId;
    
    @JsonProperty("sequence")
    private long sequence;
    
    @JsonProperty("client_send_ts_ns")
    private long clientSendTimestampNs;
    
    @JsonProperty("server_ts_ns")
    private long serverTimestampNs;
    
    @JsonProperty("server_send_ts_ns")
    private long serverSendTimestampNs;
    
    @JsonProperty("client_recv_ts_ns")
    private long clientReceiveTimestampNs;
    
    @JsonProperty("payload")
    private String payload;

    /**
     * Default constructor.
     */
    public RttMessage() {
        // Default constructor for Jackson
    }

    /**
     * Create a new RTT message with the specified sequence number.
     *
     * @param sequence The sequence number
     */
    public RttMessage(long sequence) {
        this.sequence = sequence;
    }

    /**
     * Get the message ID.
     *
     * @return The message ID
     */
    public String getMessageId() {
        return messageId;
    }

    /**
     * Set the message ID.
     *
     * @param messageId The message ID
     */
    public void setMessageId(String messageId) {
        this.messageId = messageId;
    }

    /**
     * Get the sequence number.
     *
     * @return The sequence number
     */
    public long getSequence() {
        return sequence;
    }

    /**
     * Set the sequence number.
     *
     * @param sequence The sequence number
     */
    public void setSequence(long sequence) {
        this.sequence = sequence;
    }

    /**
     * Get the client send timestamp in nanoseconds.
     *
     * @return The client send timestamp
     */
    public long getClientSendTimestampNs() {
        return clientSendTimestampNs;
    }

    /**
     * Set the client send timestamp in nanoseconds.
     *
     * @param clientSendTimestampNs The client send timestamp
     */
    public void setClientSendTimestampNs(long clientSendTimestampNs) {
        this.clientSendTimestampNs = clientSendTimestampNs;
    }

    /**
     * Get the server timestamp in nanoseconds.
     *
     * @return The server timestamp
     */
    public long getServerTimestampNs() {
        return serverTimestampNs;
    }

    /**
     * Set the server timestamp in nanoseconds.
     *
     * @param serverTimestampNs The server timestamp
     */
    public void setServerTimestampNs(long serverTimestampNs) {
        this.serverTimestampNs = serverTimestampNs;
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
     * Get the server send timestamp in nanoseconds.
     *
     * @return Server send timestamp in nanoseconds
     */
    @JsonProperty("server_send_ts_ns")
    public long getServerSendTimestampNs() {
        return serverSendTimestampNs;
    }

    /**
     * Set the server send timestamp in nanoseconds.
     *
     * @param serverSendTimestampNs Server send timestamp in nanoseconds
     */
    @JsonProperty("server_send_ts_ns")
    public void setServerSendTimestampNs(long serverSendTimestampNs) {
        this.serverSendTimestampNs = serverSendTimestampNs;
    }

    /**
     * Get the payload.
     *
     * @return The payload
     */
    public String getPayload() {
        return payload;
    }

    /**
     * Set the payload.
     *
     * @param payload The payload
     */
    public void setPayload(String payload) {
        this.payload = payload;
    }

    /**
     * Calculate the RTT (Round-Trip Time) in nanoseconds.
     *
     * @return The RTT in nanoseconds
     */
    public long calculateRttNs() {
        return clientReceiveTimestampNs - clientSendTimestampNs;
    }

    /**
     * Calculate the server processing time in nanoseconds.
     *
     * @return The server processing time in nanoseconds
     */
    public long calculateServerProcessingTimeNs() {
        return serverTimestampNs - clientSendTimestampNs;
    }

    /**
     * Calculate the client processing time in nanoseconds.
     *
     * @return The client processing time in nanoseconds
     */
    public long calculateClientProcessingTimeNs() {
        return clientReceiveTimestampNs - serverTimestampNs;
    }

    // One-way latency calculation removed as it's affected by clock skew between client and server

    @Override
    public String toString() {
        return "RttMessage{" +
                "messageId='" + messageId + '\'' +
                ", sequence=" + sequence +
                ", clientSendTimestampNs=" + clientSendTimestampNs +
                ", serverTimestampNs=" + serverTimestampNs +
                ", serverSendTimestampNs=" + serverSendTimestampNs +
                ", clientReceiveTimestampNs=" + clientReceiveTimestampNs +
                ", payload='" + (payload != null ? payload.substring(0, Math.min(20, payload.length())) + "..." : "null") + '\'' +
                '}';
    }
}