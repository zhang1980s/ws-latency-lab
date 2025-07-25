package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/zhang1980s/ws-latency-lab/ws-latency-app-golang/pkg/config"
	"github.com/zhang1980s/ws-latency-lab/ws-latency-app-golang/pkg/models"
	"github.com/zhang1980s/ws-latency-lab/ws-latency-app-golang/pkg/util"
)

// WebSocketRttServer represents a WebSocket server for RTT latency testing
type WebSocketRttServer struct {
	config       *config.RttServerConfig
	upgrader     websocket.Upgrader
	clients      map[*websocket.Conn]bool
	clientsMutex sync.RWMutex
	done         chan struct{}
	running      bool
}

// NewWebSocketRttServer creates a new WebSocket RTT server with the given configuration
func NewWebSocketRttServer(config *config.RttServerConfig) *WebSocketRttServer {
	// Set default mode
	if config.Mode == "" {
		config.Mode = "rtt"
	}

	// Set default payload size if not specified
	if config.PayloadSize == 0 {
		config.PayloadSize = 100 // Default 100 bytes
	}

	return &WebSocketRttServer{
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

// Start starts the WebSocket RTT server
func (s *WebSocketRttServer) Start() error {
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

	s.running = true

	// Start server in the main goroutine to keep the process alive
	log.Printf("Starting WebSocket RTT server on port %d", s.config.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Error starting server: %v", err)
	}

	return nil
}

// Stop stops the WebSocket RTT server
func (s *WebSocketRttServer) Stop() {
	if !s.running {
		return
	}

	// Signal to stop
	close(s.done)

	// Close all client connections
	s.clientsMutex.Lock()
	for client := range s.clients {
		client.Close()
	}
	s.clients = make(map[*websocket.Conn]bool)
	s.clientsMutex.Unlock()

	s.running = false
	log.Println("WebSocket RTT server stopped")
}

// handleHealthCheck handles health check requests
func (s *WebSocketRttServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// handleWebSocket handles WebSocket connection requests
func (s *WebSocketRttServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
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
func (s *WebSocketRttServer) handleClient(conn *websocket.Conn) {
	defer func() {
		// Unregister client on disconnect
		s.clientsMutex.Lock()
		delete(s.clients, conn)
		s.clientsMutex.Unlock()
		conn.Close()
		log.Printf("Client disconnected: %s", conn.RemoteAddr())
	}()

	for {
		// Read message from client
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
}

// handleRttRequest processes an RTT request and sends a response
func (s *WebSocketRttServer) handleRttRequest(conn *websocket.Conn, message []byte) {
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
	response := models.NewRttResponseMessage(request, serverTimestamp, util.GetCurrentTimeNanos())

	// Update the ServerSendTime right before sending
	serverSendTime := util.GetCurrentTimeNanos()
	response.ServerSendTime = serverSendTime

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

// GetClientCount returns the number of connected clients
func (s *WebSocketRttServer) GetClientCount() int {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	return len(s.clients)
}
