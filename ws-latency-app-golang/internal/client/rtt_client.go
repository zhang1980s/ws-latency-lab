package client

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"github.com/zhang1980s/ws-latency-lab/ws-latency-app-golang/pkg/config"
	"github.com/zhang1980s/ws-latency-lab/ws-latency-app-golang/pkg/models"
	"github.com/zhang1980s/ws-latency-lab/ws-latency-app-golang/pkg/stats"
	"github.com/zhang1980s/ws-latency-lab/ws-latency-app-golang/pkg/util"
)

// WebSocketRttClient represents a WebSocket client for latency testing in RTT mode
type WebSocketRttClient struct {
	config                  *config.RttClientConfig
	conn                    *websocket.Conn
	done                    chan struct{}
	statisticsLock          sync.Mutex
	statisticsCalc          *stats.StatisticsCalculator // For RTT (client send to client receive)
	oneWayLatencyStatistics *stats.StatisticsCalculator // For one-way latency (server send to client receive)
	lastReportTime          time.Time
	reportIntervalSecs      int
	sequenceCounter         int64
	pendingRequests         sync.Map // Map of sequence number to send time
}

// NewWebSocketRttClient creates a new WebSocket RTT client with the given configuration
func NewWebSocketRttClient(config *config.RttClientConfig) *WebSocketRttClient {
	return &WebSocketRttClient{
		config:                  config,
		done:                    make(chan struct{}),
		statisticsCalc:          stats.NewStatisticsCalculator(config.PrewarmCount),
		oneWayLatencyStatistics: stats.NewStatisticsCalculator(config.PrewarmCount),
		lastReportTime:          time.Now(),
		reportIntervalSecs:      5, // Report statistics every 5 seconds
		sequenceCounter:         0,
	}
}

// Run runs the WebSocket RTT client
func (c *WebSocketRttClient) Run() {
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

	// Set up request ticker
	requestInterval := time.Duration(1000/c.config.RequestsPerSecond) * time.Millisecond
	requestTicker := time.NewTicker(requestInterval)
	defer requestTicker.Stop()

	// Start message handler
	go c.handleMessages()

	// Wait for test to complete or interrupt
	for {
		select {
		case <-requestTicker.C:
			c.sendRequest()
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

// sendRequest sends an RTT request to the server
func (c *WebSocketRttClient) sendRequest() {
	// Generate payload of specified size
	payload := generatePayload(c.config.PayloadSize)

	// Create request message
	timestamp := util.GetCurrentTimeNanos()
	sequence := atomic.AddInt64(&c.sequenceCounter, 1)
	request := models.NewRttRequestMessage(timestamp, sequence, payload)

	// Store send time for this sequence
	c.pendingRequests.Store(sequence, timestamp)

	// Serialize request
	requestData, err := json.Marshal(request)
	if err != nil {
		log.Printf("Error serializing request: %v", err)
		return
	}

	// Send request
	if err := c.conn.WriteMessage(websocket.TextMessage, requestData); err != nil {
		log.Printf("Error sending request: %v", err)
		return
	}
}

// handleMessages handles incoming WebSocket messages
func (c *WebSocketRttClient) handleMessages() {
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
func (c *WebSocketRttClient) processMessage(message []byte) {
	// Parse the message
	msg, err := models.ParseMessage(message)
	if err != nil {
		log.Printf("Error parsing message: %v", err)
		return
	}

	// Check if it's an RTT response
	response, ok := msg.(*models.RttResponseMessage)
	if !ok {
		log.Printf("Received unexpected message type")
		return
	}

	// Get the send time for this sequence
	sendTimeVal, ok := c.pendingRequests.LoadAndDelete(response.Sequence)
	if !ok {
		log.Printf("Received response for unknown sequence: %d", response.Sequence)
		return
	}
	sendTime := sendTimeVal.(int64)

	// Calculate RTT (client send to client receive)
	receiveTime := util.GetCurrentTimeNanos()
	rtt := receiveTime - sendTime

	// Calculate one-way latency (server send to client receive)
	oneWayLatency := receiveTime - response.ServerSendTime

	// Add samples to statistics calculators
	c.statisticsLock.Lock()
	added := c.statisticsCalc.AddSample(rtt)
	c.oneWayLatencyStatistics.AddSample(oneWayLatency)
	c.statisticsLock.Unlock()

	if added {
		// Log every 1000th message for visibility
		if c.statisticsCalc.GetSampleCount()%1000 == 0 {
			log.Printf("Received %d responses (after warm-up)", c.statisticsCalc.GetSampleCount())
		}
	}
}

// printStatistics prints the current latency statistics
func (c *WebSocketRttClient) printStatistics() {
	c.statisticsLock.Lock()
	defer c.statisticsLock.Unlock()

	rttStats := c.statisticsCalc.Calculate()
	oneWayLatencyStats := c.oneWayLatencyStatistics.Calculate()

	log.Printf("--- RTT Statistics (last %d seconds) ---", c.reportIntervalSecs)
	log.Printf("Samples: %d", rttStats.Count)
	log.Printf("Min: %d ns", rttStats.Min)
	log.Printf("P50 (median): %d ns", rttStats.P50)
	log.Printf("P90: %d ns", rttStats.P90)
	log.Printf("P99: %d ns", rttStats.P99)
	log.Printf("Max: %d ns", rttStats.Max)
	log.Printf("Mean: %.2f ns", rttStats.Mean)
	log.Printf("----------------------------------------")

	log.Printf("--- One-way Latency Statistics (last %d seconds) ---", c.reportIntervalSecs)
	log.Printf("Samples: %d", oneWayLatencyStats.Count)
	log.Printf("Min: %d ns", oneWayLatencyStats.Min)
	log.Printf("P50 (median): %d ns", oneWayLatencyStats.P50)
	log.Printf("P90: %d ns", oneWayLatencyStats.P90)
	log.Printf("P99: %d ns", oneWayLatencyStats.P99)
	log.Printf("Max: %d ns", oneWayLatencyStats.Max)
	log.Printf("Mean: %.2f ns", oneWayLatencyStats.Mean)
	log.Printf("----------------------------------------")
}

// printFinalStatistics prints the final latency statistics
func (c *WebSocketRttClient) printFinalStatistics() {
	c.statisticsLock.Lock()
	defer c.statisticsLock.Unlock()

	rttStats := c.statisticsCalc.Calculate()
	oneWayLatencyStats := c.oneWayLatencyStatistics.Calculate()

	log.Printf("\n=== Final RTT Statistics (client send to client receive) ===")
	log.Printf("Total samples: %d", rttStats.Count)
	log.Printf("Min: %d ns", rttStats.Min)
	log.Printf("P10: %d ns", rttStats.P10)
	log.Printf("P50 (median): %d ns", rttStats.P50)
	log.Printf("P90: %d ns", rttStats.P90)
	log.Printf("P99: %d ns", rttStats.P99)
	log.Printf("Max: %d ns", rttStats.Max)
	log.Printf("Mean: %.2f ns", rttStats.Mean)
	log.Printf("Skipped warm-up samples: %d", c.statisticsCalc.GetSkippedCount())
	log.Printf("===============================")

	log.Printf("\n=== Final One-way Latency Statistics (server send to client receive) ===")
	log.Printf("Total samples: %d", oneWayLatencyStats.Count)
	log.Printf("Min: %d ns", oneWayLatencyStats.Min)
	log.Printf("P10: %d ns", oneWayLatencyStats.P10)
	log.Printf("P50 (median): %d ns", oneWayLatencyStats.P50)
	log.Printf("P90: %d ns", oneWayLatencyStats.P90)
	log.Printf("P99: %d ns", oneWayLatencyStats.P99)
	log.Printf("Max: %d ns", oneWayLatencyStats.Max)
	log.Printf("Mean: %.2f ns", oneWayLatencyStats.Mean)
	log.Printf("===============================")
}

// generatePayload generates a string payload of the specified size
func generatePayload(size int) string {
	if size <= 0 {
		return ""
	}

	// Create a payload of the specified size
	payload := make([]byte, size)
	for i := 0; i < size; i++ {
		payload[i] = 'A' + byte(i%26)
	}

	return string(payload)
}
