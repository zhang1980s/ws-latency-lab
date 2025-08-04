// Package client provides WebSocket client functionality for latency testing
package client

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"ws-latency-app-golang/pkg/stats"

	"github.com/gorilla/websocket"
)

// Config holds the configuration for the WebSocket client
type Config struct {
	ServerURL          string
	MessageRate        int
	TestDuration       int
	PrewarmCount       int
	InsecureSkipVerify bool
	Continuous         bool
}

// Client represents a WebSocket client for latency testing
type Client struct {
	config            Config
	conn              *websocket.Conn
	stats             *stats.LatencyStats
	baseMsg           map[string]interface{}
	done              chan struct{}
	expectedResponses int
	receivedResponses int
}

// NewClient creates a new WebSocket client with the given configuration
func NewClient(config Config) *Client {
	// Calculate expected responses, accounting for prewarm messages
	expectedResponses := config.MessageRate * config.TestDuration
	if config.Continuous {
		// For continuous mode, set a high number
		expectedResponses = 1000000000
	}

	// Log prewarm information if enabled
	if config.PrewarmCount > 0 {
		log.Printf("Will skip first %d messages for warm-up phase", config.PrewarmCount)
	}

	// No metrics server initialization

	// Create base message template
	baseMsg := map[string]interface{}{
		"arg": map[string]interface{}{
			"channel": "tickers",
			"instId":  "BTC-USDC",
		},
		"data": []map[string]interface{}{
			{
				"instType":  "SPOT",
				"instId":    "BTC-USDC",
				"last":      "105926",
				"lastSz":    "0.00016398",
				"askPx":     "105926.1",
				"askSz":     "0.34547131",
				"bidPx":     "105926",
				"bidSz":     "0.04848602",
				"open24h":   "103124.1",
				"high24h":   "106892.7",
				"low24h":    "102100.6",
				"sodUtc0":   "105619.9",
				"sodUtc8":   "104822.1",
				"volCcy24h": "78820585.423264371",
				"vol24h":    "755.56112024",
				"ts":        "1747721466604",
			},
		},
		"_test": map[string]interface{}{
			"client_send_ts_us": 0,
			"server_ts_us":      0,
			"client_recv_ts_us": 0,
		},
	}

	return &Client{
		config:            config,
		stats:             stats.NewLatencyStats(expectedResponses),
		baseMsg:           baseMsg,
		done:              make(chan struct{}),
		expectedResponses: expectedResponses,
		receivedResponses: 0,
	}
}

// Connect connects to the WebSocket server
func (c *Client) Connect() error {
	// No metrics updates

	// Set up WebSocket dialer with custom options for lower latency
	dialer := websocket.DefaultDialer
	dialer.NetDial = func(network, addr string) (net.Conn, error) {
		netDialer := &net.Dialer{
			Timeout: 5 * time.Second,
		}
		conn, err := netDialer.Dial(network, addr)
		if err != nil {
			return nil, err
		}

		// Set TCP_NODELAY to disable Nagle's algorithm
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.SetNoDelay(true)
		}
		return conn, nil
	}

	// Connect to WebSocket server
	log.Printf("Connecting to %s...\n", c.config.ServerURL)
	conn, _, err := dialer.Dial(c.config.ServerURL, nil)
	if err != nil {
		return fmt.Errorf("dial error: %w", err)
	}
	c.conn = conn
	log.Println("Connected to server")
	return nil
}

// Close closes the WebSocket connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// randomizeMessage updates the client's message with random values
func (c *Client) randomizeMessage() {
	msg := c.baseMsg
	data := msg["data"].([]map[string]interface{})
	item := data[0]

	// Randomize numeric values
	item["last"] = fmt.Sprintf("%d", 100000+rand.Intn(10000))
	item["lastSz"] = fmt.Sprintf("0.%08d", rand.Intn(100000000))
	item["askPx"] = fmt.Sprintf("%d.%d", 100000+rand.Intn(10000), rand.Intn(10))
	item["askSz"] = fmt.Sprintf("0.%08d", rand.Intn(100000000))
	item["bidPx"] = fmt.Sprintf("%d", 100000+rand.Intn(10000))
	item["bidSz"] = fmt.Sprintf("0.%08d", rand.Intn(100000000))
	item["open24h"] = fmt.Sprintf("%d.%d", 100000+rand.Intn(10000), rand.Intn(10))
	item["high24h"] = fmt.Sprintf("%d.%d", 100000+rand.Intn(10000), rand.Intn(10))
	item["low24h"] = fmt.Sprintf("%d.%d", 90000+rand.Intn(10000), rand.Intn(10))
	item["sodUtc0"] = fmt.Sprintf("%d.%d", 100000+rand.Intn(10000), rand.Intn(10))
	item["sodUtc8"] = fmt.Sprintf("%d.%d", 100000+rand.Intn(10000), rand.Intn(10))
	item["volCcy24h"] = fmt.Sprintf("%d.%09d", 70000000+rand.Intn(20000000), rand.Intn(1000000000))
	item["vol24h"] = fmt.Sprintf("%d.%08d", 700+rand.Intn(100), rand.Intn(100000000))
	item["ts"] = fmt.Sprintf("%d", time.Now().UnixNano()/1000000)
}

// RunTest runs the latency test
func (c *Client) RunTest() error {
	if c.conn == nil {
		return fmt.Errorf("not connected to server")
	}

	// Use the client's base message

	// Set up rate limiter
	interval := time.Second / time.Duration(c.config.MessageRate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Reset the client's state
	c.receivedResponses = 0
	c.done = make(chan struct{})

	// Set up response handler
	messageCount := 0
	go func() {
		defer close(c.done)
		for {
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}

			// Record receive time
			recvTime := time.Now().UnixNano() / 1000

			// Parse response
			var data map[string]interface{}
			if err := json.Unmarshal(message, &data); err != nil {
				log.Println("JSON parse error:", err)
				continue
			}

			// Extract timestamps
			if test, ok := data["_test"].(map[string]interface{}); ok {
				sendTime := int64(test["client_send_ts_us"].(float64))
				test["client_recv_ts_us"] = recvTime

				// Calculate RTT
				rtt := recvTime - sendTime

				// Increment message count
				messageCount++

				// Only add to statistics if we're past the warm-up phase
				if messageCount > c.config.PrewarmCount {
					c.stats.AddSample(rtt)

					c.receivedResponses++
					if c.receivedResponses >= c.expectedResponses {
						return
					}
				} else if messageCount == c.config.PrewarmCount {
					log.Printf("Warm-up phase complete. Skipped first %d messages.\n", c.config.PrewarmCount)
				}
			}
		}
	}()

	// Run test for specified duration or continuously
	sentMessages := 0
	testStart := time.Now()
	var testEnd time.Time

	if c.config.Continuous {
		log.Printf("Starting continuous test with rate %d msg/s\n", c.config.MessageRate)
		// Set testEnd to a far future time
		testEnd = testStart.Add(100 * 365 * 24 * time.Hour) // ~100 years
	} else {
		log.Printf("Starting test with rate %d msg/s for %d seconds\n", c.config.MessageRate, c.config.TestDuration)
		log.Printf("Will send approximately %d messages\n", c.config.MessageRate*c.config.TestDuration)
		testEnd = testStart.Add(time.Duration(c.config.TestDuration) * time.Second)
	}

	for time.Now().Before(testEnd) {
		<-ticker.C

		// Generate random values while keeping same format
		c.randomizeMessage()

		// Add client timestamp
		test := c.baseMsg["_test"].(map[string]interface{})
		test["client_send_ts_us"] = time.Now().UnixNano() / 1000

		// Send message
		message, err := json.Marshal(c.baseMsg)
		if err != nil {
			log.Println("JSON marshal error:", err)
			continue
		}

		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Println("Write error:", err)
			break
		}
		sentMessages++
	}

	actualDuration := time.Since(testStart)
	log.Printf("Test completed. Sent %d messages in %.2f seconds (%.2f msg/s)\n",
		sentMessages, actualDuration.Seconds(), float64(sentMessages)/actualDuration.Seconds())

	// Wait for all responses with a timeout
	log.Println("Waiting for all responses...")
	timeout := time.NewTimer(5 * time.Second)
	select {
	case <-c.done:
		log.Printf("Received all %d responses\n", c.receivedResponses)
	case <-timeout.C:
		log.Printf("Timeout waiting for responses. Received %d/%d\n", c.receivedResponses, sentMessages)
	}

	// Calculate and display statistics
	c.stats.Calculate()

	// Add information about warm-up phase if enabled
	if c.config.PrewarmCount > 0 {
		log.Printf("Note: First %d messages were skipped for warm-up phase", c.config.PrewarmCount)
	}

	c.stats.PrintResults()

	return nil
}

// GetStats returns the latency statistics
func (c *Client) GetStats() *stats.LatencyStats {
	return c.stats
}
