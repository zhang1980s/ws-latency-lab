package xyz.zzhe.wslatency.rtt.model;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * Represents a WebSocket message for RTT (Round-Trip Time) testing.
 * Uses microsecond precision for all timestamps.
 */
public class RttMessage {
    @JsonProperty("message_id")
    private String messageId;
    
    @JsonProperty("sequence")
    private long sequence;
    
    @JsonProperty("client_send_ts_us")
    private long clientSendTimestampUs;
    
    @JsonProperty("server_ts_us")
    private long serverTimestampUs;
    
    @JsonProperty("server_send_ts_us")
    private long serverSendTimestampUs;
    
    @JsonProperty("client_recv_ts_us")
    private long clientReceiveTimestampUs;
    
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
     * Get the client send timestamp in microseconds.
     *
     * @return The client send timestamp
     */
    public long getClientSendTimestampUs() {
        return clientSendTimestampUs;
    }

    /**
     * Set the client send timestamp in microseconds.
     *
     * @param clientSendTimestampUs The client send timestamp
     */
    public void setClientSendTimestampUs(long clientSendTimestampUs) {
        this.clientSendTimestampUs = clientSendTimestampUs;
    }

    /**
     * Get the server timestamp in microseconds.
     *
     * @return The server timestamp
     */
    public long getServerTimestampUs() {
        return serverTimestampUs;
    }

    /**
     * Set the server timestamp in microseconds.
     *
     * @param serverTimestampUs The server timestamp
     */
    public void setServerTimestampUs(long serverTimestampUs) {
        this.serverTimestampUs = serverTimestampUs;
    }

    /**
     * Get the client receive timestamp in microseconds.
     *
     * @return The client receive timestamp
     */
    public long getClientReceiveTimestampUs() {
        return clientReceiveTimestampUs;
    }

    /**
     * Set the client receive timestamp in microseconds.
     *
     * @param clientReceiveTimestampUs The client receive timestamp
     */
    public void setClientReceiveTimestampUs(long clientReceiveTimestampUs) {
        this.clientReceiveTimestampUs = clientReceiveTimestampUs;
    }
    
    /**
     * Get the server send timestamp in microseconds.
     *
     * @return Server send timestamp in microseconds
     */
    @JsonProperty("server_send_ts_us")
    public long getServerSendTimestampUs() {
        return serverSendTimestampUs;
    }

    /**
     * Set the server send timestamp in microseconds.
     *
     * @param serverSendTimestampUs Server send timestamp in microseconds
     */
    @JsonProperty("server_send_ts_us")
    public void setServerSendTimestampUs(long serverSendTimestampUs) {
        this.serverSendTimestampUs = serverSendTimestampUs;
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
     * Calculate the RTT (Round-Trip Time) in microseconds.
     *
     * @return The RTT in microseconds
     */
    public long calculateRttUs() {
        return clientReceiveTimestampUs - clientSendTimestampUs;
    }

    /**
     * Calculate the server processing time in microseconds.
     *
     * @return The server processing time in microseconds
     */
    public long calculateServerProcessingTimeUs() {
        return serverTimestampUs - clientSendTimestampUs;
    }

    /**
     * Calculate the client processing time in microseconds.
     *
     * @return The client processing time in microseconds
     */
    public long calculateClientProcessingTimeUs() {
        return clientReceiveTimestampUs - serverTimestampUs;
    }

    @Override
    public String toString() {
        return "RttMessage{" +
                "messageId='" + messageId + '\'' +
                ", sequence=" + sequence +
                ", clientSendTimestampUs=" + clientSendTimestampUs +
                ", serverTimestampUs=" + serverTimestampUs +
                ", serverSendTimestampUs=" + serverSendTimestampUs +
                ", clientReceiveTimestampUs=" + clientReceiveTimestampUs +
                ", payload='" + (payload != null ? payload.substring(0, Math.min(20, payload.length())) + "..." : "null") + '\'' +
                '}';
    }
}