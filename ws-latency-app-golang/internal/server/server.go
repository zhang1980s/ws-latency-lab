package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"github.com/zhang1980s/ws-latency-lab/ws-latency-app-golang/pkg/config"
	"github.com/zhang1980s/ws-latency-lab/ws-latency-app-golang/pkg/models"
	"github.com/zhang1980s/ws-latency-lab/ws-latency-app-golang/pkg/util"
)

// WebSocketServer represents a WebSocket server for latency testing
type WebSocketServer struct {
	config         *config.ServerConfig
	upgrader       websocket.Upgrader
	clients        map[*websocket.Conn]bool
	clientsMutex   sync.RWMutex
	done           chan struct{}
	messageCounter int64
	running        bool
}

// NewWebSocketServer creates a new WebSocket server with the given configuration
func NewWebSocketServer(config *config.ServerConfig) *WebSocketServer {
	// Set default mode if not specified
	if config.Mode == "" {
		config.Mode = "push"
	}

	// Calculate event interval in milliseconds from events per second
	if config.EventIntervalMs == 0 && config.EventsPerSecond > 0 {
		config.EventIntervalMs = 1000 / config.EventsPerSecond
	}

	// Set default payload size if not specified
	if config.PayloadSizeBytes == 0 {
		config.PayloadSizeBytes = 100 // Default 100 bytes
	}

	return &WebSocketServer{
		config: config,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins
			},
		},
		clients:      make(map[*websocket.Conn]bool),
		clientsMutex: sync.RWMutex{},
		done:         make(chan struct{}),
	}
}

// Start starts the WebSocket server
func (s *WebSocketServer) Start() error {
	if s.running {
		return fmt.Errorf("server is already running")
	}

	// Set up HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleHealthCheck)
	mux.HandleFunc("/ws", s.handleWebSocket)

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: mux,
	}

	// Start event generator if in server-push mode
	if s.config.Mode == "push" {
		go s.startEventGenerator()
	}

	s.running = true

	// Start server in the main goroutine to keep the process alive
	log.Printf("Starting WebSocket server on port %d", s.config.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Error starting server: %v", err)
	}

	return nil
}

// Stop stops the WebSocket server
func (s *WebSocketServer) Stop() {
	if !s.running {
		return
	}

	// Signal event generator to stop
	close(s.done)

	// Close all client connections
	s.clientsMutex.Lock()
	for client := range s.clients {
		client.Close()
	}
	s.clients = make(map[*websocket.Conn]bool)
	s.clientsMutex.Unlock()

	s.running = false
	log.Println("WebSocket server stopped")
}

// handleHealthCheck handles health check requests
func (s *WebSocketServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// handleWebSocket handles WebSocket connection requests
func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	// Register client
	s.clientsMutex.Lock()
	s.clients[conn] = true
	s.clientsMutex.Unlock()

	log.Printf("Client connected: %s", conn.RemoteAddr())

	// Handle client messages
	go s.handleClient(conn)
}

// handleClient handles messages from a connected client
func (s *WebSocketServer) handleClient(conn *websocket.Conn) {
	defer func() {
		// Unregister client on disconnect
		s.clientsMutex.Lock()
		delete(s.clients, conn)
		s.clientsMutex.Unlock()
		conn.Close()
		log.Printf("Client disconnected: %s", conn.RemoteAddr())
	}()

	// If in RTT mode, handle client requests
	if s.config.Mode == "rtt" {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Error reading message: %v", err)
				}
				break
			}

			// Process RTT request
			go s.handleRttRequest(conn, message)
		}
	} else {
		// In server-push mode, just keep the connection open
		// Events are sent from the event generator
		<-s.done
	}
}

// handleRttRequest processes an RTT request and sends a response
func (s *WebSocketServer) handleRttRequest(conn *websocket.Conn, message []byte) {
	// Parse the request message
	msg, err := models.ParseMessage(message)
	if err != nil {
		log.Printf("Error parsing message: %v", err)
		return
	}

	// Check if it's an RTT request
	request, ok := msg.(*models.RttRequestMessage)
	if !ok {
		log.Printf("Received non-RTT request message")
		return
	}

	// Get server processing timestamp
	serverTimestamp := util.GetCurrentTimeNanos()

	// Create response message
	response := models.NewRttResponseMessage(request, serverTimestamp, serverTimestamp)

	// Serialize response
	responseData, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error serializing response: %v", err)
		return
	}

	// Send response
	if err := conn.WriteMessage(websocket.TextMessage, responseData); err != nil {
		log.Printf("Error sending response: %v", err)
		return
	}
}

// startEventGenerator starts generating and sending events to clients
func (s *WebSocketServer) startEventGenerator() {
	ticker := time.NewTicker(time.Duration(s.config.EventIntervalMs) * time.Millisecond)
	defer ticker.Stop()

	log.Printf("Starting event generator with interval %d ms", s.config.EventIntervalMs)

	for {
		select {
		case <-ticker.C:
			s.sendEventToAllClients()
		case <-s.done:
			return
		}
	}
}

// sendEventToAllClients sends an event to all connected clients
func (s *WebSocketServer) sendEventToAllClients() {
	// Generate payload of specified size
	payload := generatePayload(s.config.PayloadSizeBytes)

	// Create event message
	timestamp := util.GetCurrentTimeNanos()
	sequence := atomic.AddInt64(&s.messageCounter, 1)
	event := models.NewServerPushMessage(timestamp, sequence, payload)

	// Serialize event
	eventData, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error serializing event: %v", err)
		return
	}

	// Send to all clients
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	for client := range s.clients {
		// Send in a non-blocking way to avoid one slow client affecting others
		go func(c *websocket.Conn) {
			if err := c.WriteMessage(websocket.TextMessage, eventData); err != nil {
				log.Printf("Error sending event to client %s: %v", c.RemoteAddr(), err)
			}
		}(client)
	}
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

// GetClientCount returns the number of connected clients
func (s *WebSocketServer) GetClientCount() int {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	return len(s.clients)
}
