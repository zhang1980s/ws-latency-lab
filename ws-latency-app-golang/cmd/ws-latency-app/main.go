package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/zhang1980s/ws-latency-lab/ws-latency-app-golang/internal/client"
	"github.com/zhang1980s/ws-latency-lab/ws-latency-app-golang/internal/server"
	"github.com/zhang1980s/ws-latency-lab/ws-latency-app-golang/pkg/config"
)

func main() {
	// Define command line flags
	appType := flag.String("a", "push", "Application type: 'push' for server-push model or 'rtt' for request-response model")
	mode := flag.String("m", "", "Mode to run: 'server' or 'client'")
	port := flag.Int("p", 10443, "Port for server to listen on")
	rate := flag.Int("r", 10, "Events/requests per second")
	serverURL := flag.String("s", "ws://localhost:10443/ws", "WebSocket server address")
	duration := flag.Int("d", 30, "Test duration in seconds")
	payloadSize := flag.Int("payload-size", 100, "Size of the message payload in bytes")
	prewarmCount := flag.Int("prewarm-count", 100, "Skip calculating latency for first N events")
	insecureSkipVerify := flag.Bool("insecure", false, "Skip TLS certificate verification")
	continuous := flag.Bool("continuous", false, "Run in continuous monitoring mode")

	// Parse command line flags
	flag.Parse()

	// Validate required flags
	if *mode == "" {
		fmt.Println("Error: -m/--mode flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Convert to lowercase for case-insensitive comparison
	modeStr := strings.ToLower(*mode)
	appTypeStr := strings.ToLower(*appType)

	// Validate mode
	if modeStr != "server" && modeStr != "client" {
		fmt.Printf("Error: Invalid mode: %s. Must be 'server' or 'client'\n", *mode)
		flag.Usage()
		os.Exit(1)
	}

	// Validate app type
	if appTypeStr != "push" && appTypeStr != "rtt" {
		fmt.Printf("Error: Invalid application type: %s. Must be 'push' or 'rtt'\n", *appType)
		flag.Usage()
		os.Exit(1)
	}

	// Run in server or client mode based on options
	if modeStr == "server" {
		if appTypeStr == "push" {
			// Create server configuration
			serverConfig := &config.ServerConfig{
				Port:             *port,
				EventsPerSecond:  *rate,
				Mode:             "push",
				EventIntervalMs:  1000 / *rate,
				PayloadSizeBytes: *payloadSize,
			}

			// Create and start server
			wsServer := server.NewWebSocketServer(serverConfig)
			log.Printf("Starting server-push WebSocket server on port %d with event rate %d events/sec\n", *port, *rate)
			wsServer.Start()
		} else { // rtt mode
			// Create RTT server configuration
			rttServerConfig := &config.RttServerConfig{
				Port:        *port,
				PayloadSize: *payloadSize,
				Mode:        "rtt",
			}

			// Create and start RTT server
			rttServer := server.NewWebSocketRttServer(rttServerConfig)
			rttServer.Start()
		}
	} else { // client mode
		if appTypeStr == "push" {
			// Create client configuration
			clientConfig := &config.ClientConfig{
				ServerURL:          *serverURL,
				TestDuration:       *duration,
				PrewarmCount:       *prewarmCount,
				InsecureSkipVerify: *insecureSkipVerify,
				Continuous:         *continuous,
			}

			// Create and run client
			wsClient := client.NewWebSocketClient(clientConfig)
			log.Printf("Starting server-push WebSocket client connecting to %s for %s\n",
				*serverURL, getContinuousOrDuration(*continuous, *duration))
			wsClient.Run()
		} else { // rtt mode
			// Create RTT client configuration
			rttClientConfig := &config.RttClientConfig{
				ServerURL:          *serverURL,
				TestDuration:       *duration,
				RequestsPerSecond:  *rate,
				PayloadSize:        *payloadSize,
				PrewarmCount:       *prewarmCount,
				InsecureSkipVerify: *insecureSkipVerify,
				Continuous:         *continuous,
			}

			// Create and run RTT client
			rttClient := client.NewWebSocketRttClient(rttClientConfig)
			log.Printf("Starting RTT WebSocket client connecting to %s for %s with rate %d req/sec\n",
				*serverURL, getContinuousOrDuration(*continuous, *duration), *rate)
			rttClient.Run()
		}
	}
}

// Helper function to get duration string
func getContinuousOrDuration(continuous bool, duration int) string {
	if continuous {
		return "continuous monitoring"
	}
	return fmt.Sprintf("%d seconds", duration)
}
