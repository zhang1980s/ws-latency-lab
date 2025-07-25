package models

import (
	"encoding/json"
	"fmt"
)

// MessageType defines the type of WebSocket message
type MessageType string

const (
	// Server-push message types
	ServerPushEvent MessageType = "event"

	// RTT message types
	RttRequest  MessageType = "request"
	RttResponse MessageType = "response"
)

// BaseMessage is the common structure for all WebSocket messages
type BaseMessage struct {
	Type      MessageType `json:"type"`
	Timestamp int64       `json:"timestamp"` // Nanosecond timestamp
	Sequence  int64       `json:"sequence"`  // Sequence number for message ordering
}

// ServerPushMessage represents a server-pushed event message
type ServerPushMessage struct {
	BaseMessage
	Payload string `json:"payload"` // Optional payload data
}

// RttRequestMessage represents a client request in RTT mode
type RttRequestMessage struct {
	BaseMessage
	ClientTimestamp int64  `json:"clientTimestamp"` // Client-side timestamp in nanoseconds
	Payload         string `json:"payload"`         // Optional payload data
}

// RttResponseMessage represents a server response in RTT mode
type RttResponseMessage struct {
	BaseMessage
	ClientTimestamp int64  `json:"clientTimestamp"` // Original client timestamp from request
	ServerTimestamp int64  `json:"serverTimestamp"` // Server processing timestamp in nanoseconds
	ServerSendTime  int64  `json:"serverSendTime"`  // Server send timestamp in nanoseconds
	Payload         string `json:"payload"`         // Optional payload data (echo of request payload)
}

// NewServerPushMessage creates a new server push event message
func NewServerPushMessage(timestamp int64, sequence int64, payload string) *ServerPushMessage {
	return &ServerPushMessage{
		BaseMessage: BaseMessage{
			Type:      ServerPushEvent,
			Timestamp: timestamp,
			Sequence:  sequence,
		},
		Payload: payload,
	}
}

// NewRttRequestMessage creates a new RTT request message
func NewRttRequestMessage(timestamp int64, sequence int64, payload string) *RttRequestMessage {
	return &RttRequestMessage{
		BaseMessage: BaseMessage{
			Type:      RttRequest,
			Timestamp: timestamp,
			Sequence:  sequence,
		},
		ClientTimestamp: timestamp,
		Payload:         payload,
	}
}

// NewRttResponseMessage creates a new RTT response message
func NewRttResponseMessage(request *RttRequestMessage, serverTimestamp int64, responseTimestamp int64) *RttResponseMessage {
	return &RttResponseMessage{
		BaseMessage: BaseMessage{
			Type:      RttResponse,
			Timestamp: responseTimestamp,
			Sequence:  request.Sequence,
		},
		ClientTimestamp: request.ClientTimestamp,
		ServerTimestamp: serverTimestamp,
		ServerSendTime:  responseTimestamp, // Default to responseTimestamp, will be updated before sending
		Payload:         request.Payload,   // Echo back the request payload
	}
}

// ParseMessage parses a JSON message and returns the appropriate message type
func ParseMessage(data []byte) (interface{}, error) {
	// Parse the base message to determine the type
	var base BaseMessage
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}

	// Based on the message type, parse into the appropriate struct
	switch base.Type {
	case ServerPushEvent:
		var msg ServerPushMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse server push message: %w", err)
		}
		return &msg, nil

	case RttRequest:
		var msg RttRequestMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse RTT request message: %w", err)
		}
		return &msg, nil

	case RttResponse:
		var msg RttResponseMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse RTT response message: %w", err)
		}
		return &msg, nil

	default:
		return nil, fmt.Errorf("unknown message type: %s", base.Type)
	}
}
