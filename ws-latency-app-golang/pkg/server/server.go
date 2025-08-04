// Package server provides WebSocket server functionality for latency testing
package server

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Config holds the configuration for the WebSocket server
type Config struct {
	Port string
}

// Server represents a WebSocket server for latency testing
type Server struct {
	config   Config
	upgrader websocket.Upgrader
}

// NewServer creates a new WebSocket server with the given configuration
func NewServer(config Config) *Server {
	return &Server{
		config: config,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all connections for testing purposes
			},
		},
	}
}

// Start starts the WebSocket server
func (s *Server) Start() error {
	// Add a health check endpoint
	http.HandleFunc("/health", s.handleHealth)

	// Set up WebSocket handler
	http.HandleFunc("/ws", s.handleConnection)

	// Start server
	log.Printf("WebSocket server starting on port %s...\n", s.config.Port)
	log.Printf("Connect to: ws://localhost:%s/ws\n", s.config.Port)
	log.Printf("Health check available at: http://localhost:%s/health\n", s.config.Port)
	return http.ListenAndServe(":"+s.config.Port, nil)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
}

// processMessage processes a WebSocket message by adding server timestamp
func (s *Server) processMessage(message []byte) ([]byte, error) {
	// Parse message
	var data map[string]interface{}
	if err := json.Unmarshal(message, &data); err != nil {
		return nil, err
	}

	// Add server timestamp
	if test, ok := data["_test"].(map[string]interface{}); ok {
		test["server_ts_us"] = time.Now().UnixNano() / 1000
	} else {
		test := make(map[string]interface{})
		test["server_ts_us"] = time.Now().UnixNano() / 1000
		data["_test"] = test
	}

	// Serialize and return
	return json.Marshal(data)
}

// handleConnection handles WebSocket connections
func (s *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	// Get client IP address
	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = r.Header.Get("X-Real-IP")
	}
	if clientIP == "" {
		clientIP, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	// Get TCP client IP from the underlying connection
	tcpAddr, ok := conn.UnderlyingConn().RemoteAddr().(*net.TCPAddr)
	var tcpIP string
	if ok {
		tcpIP = tcpAddr.IP.String()
	} else {
		tcpIP = "unknown"
	}

	log.Printf("Client connected - HTTP IP: %s, TCP IP: %s", clientIP, tcpIP)

	// Set TCP_NODELAY to disable Nagle's algorithm for lower latency
	if tcpConn, ok := conn.UnderlyingConn().(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
	}

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		// Process the message
		response, err := s.processMessage(message)
		if err != nil {
			log.Println("Process error:", err)
			continue
		}

		// Send the response
		if err := conn.WriteMessage(messageType, response); err != nil {
			log.Println("Write error:", err)
			break
		}
	}

	log.Printf("Client disconnected - HTTP IP: %s, TCP IP: %s", clientIP, tcpIP)
}
