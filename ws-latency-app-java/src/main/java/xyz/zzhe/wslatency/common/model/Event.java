package xyz.zzhe.wslatency.common.model;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * Represents a WebSocket event for latency testing.
 */
public class Event {
    @JsonProperty("event_type")
    private String eventType;
    
    @JsonProperty("sequence")
    private long sequence;
    
    @JsonProperty("data")
    private EventData data;
    
    @JsonProperty("_test")
    private TestMetadata testMetadata;

    /**
     * Default constructor.
     */
    public Event() {
        // Default constructor for Jackson
    }

    /**
     * Create a new event with the specified type and sequence.
     *
     * @param eventType The event type
     * @param sequence  The event sequence number
     */
    public Event(String eventType, long sequence) {
        this.eventType = eventType;
        this.sequence = sequence;
        this.data = new EventData();
        this.testMetadata = new TestMetadata();
    }

    /**
     * Get the event type.
     *
     * @return The event type
     */
    public String getEventType() {
        return eventType;
    }

    /**
     * Set the event type.
     *
     * @param eventType The event type
     */
    public void setEventType(String eventType) {
        this.eventType = eventType;
    }

    /**
     * Get the event sequence number.
     *
     * @return The sequence number
     */
    public long getSequence() {
        return sequence;
    }

    /**
     * Set the event sequence number.
     *
     * @param sequence The sequence number
     */
    public void setSequence(long sequence) {
        this.sequence = sequence;
    }

    /**
     * Get the event data.
     *
     * @return The event data
     */
    public EventData getData() {
        return data;
    }

    /**
     * Set the event data.
     *
     * @param data The event data
     */
    public void setData(EventData data) {
        this.data = data;
    }

    /**
     * Get the test metadata.
     *
     * @return The test metadata
     */
    public TestMetadata getTestMetadata() {
        return testMetadata;
    }

    /**
     * Set the test metadata.
     *
     * @param testMetadata The test metadata
     */
    public void setTestMetadata(TestMetadata testMetadata) {
        this.testMetadata = testMetadata;
    }

    @Override
    public String toString() {
        return "Event{" +
                "eventType='" + eventType + '\'' +
                ", sequence=" + sequence +
                ", data=" + data +
                ", testMetadata=" + testMetadata +
                '}';
    }
}