# WebSocket Latency Testing Application (Go Implementation)

This application is designed to measure WebSocket latency in two different modes:

1. **Server-Push Mode**: The server generates events at a specified rate and pushes them to connected clients. The client measures the one-way latency from server to client.

2. **RTT (Round-Trip Time) Mode**: The client sends requests to the server at a specified rate, and the server responds immediately. The client measures the round-trip time.

## Features

- High-precision latency measurements with nanosecond resolution
- Support for both server-push and RTT measurement modes
- Dual latency metrics in RTT mode (full RTT and server-to-client one-way latency)
- Configurable event/request rates
- Configurable payload sizes
- Warm-up period to exclude initial connection overhead
- Detailed latency statistics (min, max, mean, percentiles)
- Support for secure WebSocket connections (wss://)
- Continuous monitoring mode

## Prerequisites

- Go 1.24 or later

## Installation

```bash
# Clone the repository
git clone https://github.com/zhang1980s/ws-latency-lab.git
cd ws-latency-lab/ws-latency-app-golang

# Install dependencies
go mod download
```

## Building

### Using Go directly

```bash
# Build for current platform
go build -o build/ws-latency-app cmd/ws-latency-app/main.go
```

### Using Makefile

```bash
# Build for current platform
make build

# Build for all supported platforms (Linux, macOS, Windows - both AMD64 and ARM64)
make build-all

# Build for specific platform
make build-linux-amd64
make build-linux-arm64
make build-darwin-amd64
make build-darwin-arm64
make build-windows-amd64

# Clean build directory
make clean

# Show all available make targets
make help
```

### Using Docker

```bash
# Build Docker image
docker build -t ws-latency-app:latest .

# Run server in Docker
docker run -p 10443:10443 ws-latency-app:latest -m server -a push -p 10443 -r 10

# Run client in Docker (connecting to a server)
docker run ws-latency-app:latest -m client -a push -s ws://server-address:10443/ws -d 30
```

## Usage

### Server-Push Mode

#### Start the server:

```bash
go run cmd/ws-latency-app/main.go -m server -a push -p 10443 -r 10
```

Parameters:
- `-m server`: Run in server mode
- `-a push`: Use server-push application type
- `-p 10443`: Listen on port 10443
- `-r 10`: Generate 10 events per second

#### Start the client:

```bash
go run cmd/ws-latency-app/main.go -m client -a push -s ws://localhost:10443/ws -d 30 --prewarm-count 100
```

Parameters:
- `-m client`: Run in client mode
- `-a push`: Use server-push application type
- `-s ws://localhost:10443/ws`: Server WebSocket URL
- `-d 30`: Run test for 30 seconds
- `--prewarm-count 100`: Skip first 100 messages for warm-up

### RTT Mode

#### Start the server:

```bash
go run cmd/ws-latency-app/main.go -m server -a rtt -p 10443 --payload-size 100
```

Parameters:
- `-m server`: Run in server mode
- `-a rtt`: Use RTT application type
- `-p 10443`: Listen on port 10443
- `--payload-size 100`: Use 100 bytes payload size

#### Start the client:

```bash
go run cmd/ws-latency-app/main.go -m client -a rtt -s ws://localhost:10443/ws -d 30 -r 10 --prewarm-count 100 --payload-size 100
```

Parameters:
- `-m client`: Run in client mode
- `-a rtt`: Use RTT application type
- `-s ws://localhost:10443/ws`: Server WebSocket URL
- `-d 30`: Run test for 30 seconds
- `-r 10`: Send 10 requests per second
- `--prewarm-count 100`: Skip first 100 messages for warm-up
- `--payload-size 100`: Use 100 bytes payload size

## Command-Line Arguments

### Common Arguments

- `-m, --mode`: Mode to run: 'server' or 'client' (required)
- `-a, --app-type`: Application type: 'push' for server-push model or 'rtt' for request-response model (default: "push")
- `-p, --port`: Port for server to listen on (default: 10443)
- `--payload-size`: Size of the message payload in bytes (default: 100)
- `--insecure`: Skip TLS certificate verification

### Server Arguments

- `-r, --rate`: Events per second in server-push mode (default: 10)

### Client Arguments

- `-s, --server`: WebSocket server address (default: "ws://localhost:10443/ws")
- `-d, --duration`: Test duration in seconds (default: 30)
- `-r, --rate`: Requests per second in RTT mode (default: 10)
- `--prewarm-count`: Skip calculating latency for first N events (default: 100)
- `--continuous`: Run in continuous monitoring mode

## Latency Calculation

### Server-Push Mode

In server-push mode, the server includes a timestamp in each event message. When the client receives the message, it calculates the difference between the current time and the server's timestamp. This provides a one-way latency measurement from server to client.

Note: This requires clock synchronization between server and client for accurate measurements.

### RTT Mode
In RTT mode, the application now measures two different latency metrics:

1. **RTT (Round-Trip Time)**: This is the full round-trip time from when the client sends a request to when it receives the response. It's calculated as `receiveTime - sendTime` where:
   - `sendTime`: The timestamp when the client sent the request
   - `receiveTime`: The timestamp when the client received the response

2. **One-way Latency**: This is the one-way latency from when the server sends the response to when the client receives it. It's calculated as `receiveTime - serverSendTime` where:
   - `serverSendTime`: The timestamp when the server sent the response (captured right before sending)
   - `receiveTime`: The timestamp when the client received the response

The difference between these metrics:
- RTT measures the complete request-response cycle (client → server → client)
- One-way latency measures only the server-to-client portion of the journey

These two metrics together provide more detailed insights into network performance:
- If RTT is high but one-way latency is low, the bottleneck is likely in the client-to-server direction
- If both RTT and one-way latency are high, the bottleneck is likely in the server-to-client direction
- The difference between RTT and one-way latency can help identify server processing time


## Output

The application outputs latency statistics periodically during the test and a final summary at the end. In RTT mode, it shows statistics for both RTT and message_rtt metrics:

```
--- RTT Statistics (last 5 seconds) ---
Samples: 50
Min: 1234567 ns
P50 (median): 2345678 ns
P90: 3456789 ns
P99: 4567890 ns
Max: 5678901 ns
Mean: 2345678.90 ns
Skipped warm-up samples: 100
----------------------------------------

--- One-way Latency Statistics (last 5 seconds) ---
Samples: 50
Min: 567890 ns
P50 (median): 678901 ns
P90: 789012 ns
P99: 890123 ns
Max: 901234 ns
Mean: 678901.23 ns
----------------------------------------

=== Final RTT Statistics (client send to client receive) ===
Total samples: 300
Min: 1234567 ns
P10: 1456789 ns
P50 (median): 2345678 ns
P90: 3456789 ns
P99: 4567890 ns
Max: 5678901 ns
Mean: 2345678.90 ns
Skipped warm-up samples: 100
===============================

=== Final One-way Latency Statistics (server send to client receive) ===
Total samples: 300
Min: 567890 ns
P10: 623456 ns
P50 (median): 678901 ns
P90: 789012 ns
P99: 890123 ns
Max: 901234 ns
Mean: 678901.23 ns
===============================
```

In server-push mode, it shows only one set of latency statistics for the one-way server-to-client latency.

## License

MIT