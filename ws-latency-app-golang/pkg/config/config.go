package config

// Application mode constants
const (
	ServerPushMode = "push"
	RttMode        = "rtt"
)

// ServerConfig holds configuration for the WebSocket server in push mode
type ServerConfig struct {
	Port             int
	EventsPerSecond  int
	Mode             string // "push" or "rtt"
	EventIntervalMs  int    // Interval between events in milliseconds
	PayloadSizeBytes int    // Size of the message payload in bytes
}

// ClientConfig holds configuration for the WebSocket client in push mode
type ClientConfig struct {
	ServerURL          string
	TestDuration       int
	PrewarmCount       int
	InsecureSkipVerify bool
	Continuous         bool
}

// RttServerConfig holds configuration for the WebSocket server in RTT mode
type RttServerConfig struct {
	Port        int
	PayloadSize int
	Mode        string // Always "rtt"
}

// RttClientConfig holds configuration for the WebSocket client in RTT mode
type RttClientConfig struct {
	ServerURL          string
	TestDuration       int
	RequestsPerSecond  int
	PayloadSize        int
	PrewarmCount       int
	InsecureSkipVerify bool
	Continuous         bool
}
