// Command ws-latency-app is a WebSocket latency testing tool.
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"ws-latency-app-golang/pkg/client"
	"ws-latency-app-golang/pkg/server"
)

// Command line flags
var (
	// Common flags
	mode = flag.String("mode", "", "Mode to run: 'server' or 'client' (required)")

	// Server flags
	port = flag.String("port", "8080", "Port for server to listen on")

	// Client flags
	serverAddr         = flag.String("server", "ws://localhost:8080/ws", "WebSocket server address for client")
	messageRate        = flag.Int("rate", 10, "Messages per second for client")
	testDuration       = flag.Int("duration", 30, "Test duration in seconds for client")
	prewarmCount       = flag.Int("prewarm-count", 100, "Skip calculating RTT for first N messages (for warm-up)")
	insecureSkipVerify = flag.Bool("insecure", false, "Skip TLS certificate verification (not recommended for production)")
	continuous         = flag.Bool("continuous", false, "Run in continuous monitoring mode")
)

func init() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
}

func main() {
	// Parse command line flags
	flag.Parse()

	// Validate mode
	if *mode == "" {
		fmt.Println("Error: -mode flag is required")
		printUsage()
		os.Exit(1)
	}

	// Run in specified mode
	switch *mode {
	case "server":
		runServer()
	case "client":
		runClient()
	default:
		fmt.Printf("Error: Invalid mode '%s'. Must be 'server' or 'client'\n", *mode)
		printUsage()
		os.Exit(1)
	}
}

// printUsage prints the usage information.
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  Server mode: ws-latency-app -mode=server [-port=8080]")
	fmt.Println("  Client mode: ws-latency-app -mode=client [-server=ws://localhost:8080/ws] [-rate=10] [-duration=30] [-prewarm-count=100] [-insecure] [-continuous]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -prewarm-count  Skip calculating RTT for first N messages (default: 100)")
	fmt.Println("  -insecure       Skip TLS certificate verification (not recommended for production)")
	fmt.Println("  -continuous     Run in continuous monitoring mode (ignores duration)")
}

// runServer starts the WebSocket server.
func runServer() {
	// Create server configuration
	config := server.Config{
		Port: *port,
	}

	// Create and start server
	srv := server.NewServer(config)
	log.Fatal(srv.Start())
}

// runClient runs the WebSocket client.
func runClient() {
	// Create client configuration
	config := client.Config{
		ServerURL:          *serverAddr,
		MessageRate:        *messageRate,
		TestDuration:       *testDuration,
		PrewarmCount:       *prewarmCount,
		InsecureSkipVerify: *insecureSkipVerify,
		Continuous:         *continuous,
	}

	// Create client
	c := client.NewClient(config)

	// Connect to server
	if err := c.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()

	// Run test
	if err := c.RunTest(); err != nil {
		log.Fatalf("Test failed: %v", err)
	}
}
