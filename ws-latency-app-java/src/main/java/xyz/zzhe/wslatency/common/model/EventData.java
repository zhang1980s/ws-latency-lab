package xyz.zzhe.wslatency.common.model;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * Represents the data payload of a WebSocket event.
 */
public class EventData {
    @JsonProperty("payload")
    private String payload;

    /**
     * Default constructor.
     */
    public EventData() {
        // Default constructor for Jackson
    }

    /**
     * Create a new event data with the specified payload.
     *
     * @param payload The payload string
     */
    public EventData(String payload) {
        this.payload = payload;
    }

    /**
     * Get the payload.
     *
     * @return The payload string
     */
    public String getPayload() {
        return payload;
    }

    /**
     * Set the payload.
     *
     * @param payload The payload string
     */
    public void setPayload(String payload) {
        this.payload = payload;
    }

    @Override
    public String toString() {
        return "EventData{" +
                "payload='" + (payload != null ? payload.substring(0, Math.min(20, payload.length())) + "..." : "null") + '\'' +
                '}';
    }
}