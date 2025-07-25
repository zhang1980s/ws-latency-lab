package client

import (
	"crypto/tls"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/zhang1980s/ws-latency-lab/ws-latency-app-golang/pkg/config"
	"github.com/zhang1980s/ws-latency-lab/ws-latency-app-golang/pkg/models"
	"github.com/zhang1980s/ws-latency-lab/ws-latency-app-golang/pkg/stats"
	"github.com/zhang1980s/ws-latency-lab/ws-latency-app-golang/pkg/util"
)

// WebSocketClient represents a WebSocket client for latency testing in server-push mode
type WebSocketClient struct {
	config             *config.ClientConfig
	conn               *websocket.Conn
	done               chan struct{}
	statisticsLock     sync.Mutex
	statisticsCalc     *stats.StatisticsCalculator
	lastReportTime     time.Time
	reportIntervalSecs int
}

// NewWebSocketClient creates a new WebSocket client with the given configuration
func NewWebSocketClient(config *config.ClientConfig) *WebSocketClient {
	return &WebSocketClient{
		config:             config,
		done:               make(chan struct{}),
		statisticsCalc:     stats.NewStatisticsCalculator(config.PrewarmCount),
		lastReportTime:     time.Now(),
		reportIntervalSecs: 5, // Report statistics every 5 seconds
	}
}

// Run runs the WebSocket client
func (c *WebSocketClient) Run() {
	// Parse server URL
	u, err := url.Parse(c.config.ServerURL)
	if err != nil {
		log.Fatalf("Error parsing server URL: %v", err)
	}

	// Set up WebSocket dialer with TLS configuration
	dialer := websocket.DefaultDialer
	if c.config.InsecureSkipVerify {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Connect to server
	log.Printf("Connecting to %s", c.config.ServerURL)
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}
	c.conn = conn
	defer c.conn.Close()

	// Set up interrupt handler
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Set up test duration timer
	var durationTimer *time.Timer
	if !c.config.Continuous {
		durationTimer = time.NewTimer(time.Duration(c.config.TestDuration) * time.Second)
	}

	// Set up statistics reporting ticker
	reportTicker := time.NewTicker(time.Duration(c.reportIntervalSecs) * time.Second)
	defer reportTicker.Stop()

	// Start message handler
	go c.handleMessages()

	// Wait for test to complete or interrupt
	for {
		select {
		case <-durationTimer.C:
			if !c.config.Continuous {
				log.Printf("Test duration of %d seconds completed", c.config.TestDuration)
				c.printFinalStatistics()
				return
			}
		case <-reportTicker.C:
			c.printStatistics()
		case <-interrupt:
			log.Println("Interrupt received, closing connection")
			// Cleanly close the connection by sending a close message
			err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Printf("Error sending close message: %v", err)
			}
			// Wait for the server to close the connection
			select {
			case <-c.done:
			case <-time.After(time.Second):
			}
			c.printFinalStatistics()
			return
		case <-c.done:
			log.Println("Connection closed by server")
			c.printFinalStatistics()
			return
		}
	}
}

// handleMessages handles incoming WebSocket messages
func (c *WebSocketClient) handleMessages() {
	defer close(c.done)
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			return
		}

		// Process the message
		c.processMessage(message)
	}
}

// processMessage processes an incoming WebSocket message
func (c *WebSocketClient) processMessage(message []byte) {
	// Parse the message
	msg, err := models.ParseMessage(message)
	if err != nil {
		log.Printf("Error parsing message: %v", err)
		return
	}

	// Check if it's a server push event
	event, ok := msg.(*models.ServerPushMessage)
	if !ok {
		log.Printf("Received unexpected message type")
		return
	}

	// Calculate latency
	receiveTime := util.GetCurrentTimeNanos()
	latency := receiveTime - event.Timestamp

	// Add latency sample to statistics calculator
	c.statisticsLock.Lock()
	added := c.statisticsCalc.AddSample(latency)
	c.statisticsLock.Unlock()

	if added {
		// Log every 1000th message for visibility
		if c.statisticsCalc.GetSampleCount()%1000 == 0 {
			log.Printf("Received %d messages (after warm-up)", c.statisticsCalc.GetSampleCount())
		}
	}
}

// printStatistics prints the current latency statistics
func (c *WebSocketClient) printStatistics() {
	c.statisticsLock.Lock()
	defer c.statisticsLock.Unlock()

	stats := c.statisticsCalc.Calculate()
	log.Printf("--- Latency Statistics (last %d seconds) ---", c.reportIntervalSecs)
	log.Printf("Samples: %d", stats.Count)
	log.Printf("Min: %d ns", stats.Min)
	log.Printf("P50 (median): %d ns", stats.P50)
	log.Printf("P90: %d ns", stats.P90)
	log.Printf("P99: %d ns", stats.P99)
	log.Printf("Max: %d ns", stats.Max)
	log.Printf("Mean: %.2f ns", stats.Mean)
	log.Printf("Skipped warm-up samples: %d", c.statisticsCalc.GetSkippedCount())
	log.Printf("----------------------------------------")
}

// printFinalStatistics prints the final latency statistics
func (c *WebSocketClient) printFinalStatistics() {
	c.statisticsLock.Lock()
	defer c.statisticsLock.Unlock()

	stats := c.statisticsCalc.Calculate()
	log.Printf("\n=== Final Latency Statistics ===")
	log.Printf("Total samples: %d", stats.Count)
	log.Printf("Min: %d ns", stats.Min)
	log.Printf("P10: %d ns", stats.P10)
	log.Printf("P50 (median): %d ns", stats.P50)
	log.Printf("P90: %d ns", stats.P90)
	log.Printf("P99: %d ns", stats.P99)
	log.Printf("Max: %d ns", stats.Max)
	log.Printf("Mean: %.2f ns", stats.Mean)
	log.Printf("Skipped warm-up samples: %d", c.statisticsCalc.GetSkippedCount())
	log.Printf("===============================")
}
