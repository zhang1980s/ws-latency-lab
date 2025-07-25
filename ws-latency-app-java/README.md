# WebSocket Latency Testing Application (Java)

A high-performance WebSocket latency testing tool written in Java. This application measures the one-way latency from a WebSocket server to clients with precise nanosecond timing.

## Features

- Server-push model: server generates events and sends to clients
- Netty-based implementation for high performance and low latency
- Configurable event rate (events per second)
- Detailed latency statistics (min, max, P10, P50, P90, P99, mean)
- Direct console output of latency metrics
- Low-latency optimizations (TCP_NODELAY, native epoll transport when available, etc.)
- Docker containerization

## How It Works

```
┌─────────────────┐                                  ┌─────────────────┐
│                 │                                  │                 │
│  WebSocket      │                                  │  WebSocket      │
│  Server         │                                  │  Client         │
│                 │                                  │                 │
│  ┌───────────┐  │                                  │  ┌───────────┐  │
│  │           │  │  1. Generate event with          │  │           │  │
│  │ Event     │──┼─────timestamp────────────────────┼─▶│ WebSocket │  │
│  │ Generator │  │                                  │  │ Handler   │  │
│  │           │  │                                  │  │           │  │
│  └───────────┘  │                                  │  └─────┬─────┘  │
│                 │                                  │        │        │
│                 │                                  │        │        │
│                 │                                  │        ▼        │
│                 │                                  │  ┌───────────┐  │
│                 │                                  │  │ Record    │  │
│                 │                                  │  │ Receive   │  │
│                 │                                  │  │ Timestamp │  │
│                 │                                  │  └─────┬─────┘  │
│                 │                                  │        │        │
│                 │                                  │        │        │
│                 │                                  │        ▼        │
│                 │                                  │  ┌───────────┐  │
│                 │                                  │  │ Calculate │  │
│                 │                                  │  │ Latency   │  │
│                 │                                  │  └─────┬─────┘  │
│                 │                                  │        │        │
│                 │                                  │        │        │
│                 │                                  │        ▼        │
│                 │                                  │  ┌───────────┐  │
│                 │                                  │  │ Statistics│  │
│                 │                                  │  │ Calculator│  │
│                 │                                  │  └─────┬─────┘  │
│                 │                                  │        │        │
│                 │                                  │        │        │
│                 │                                  │        ▼        │
│                 │                                  │  ┌───────────┐  │
│                 │                                  │  │ Output    │  │
│                 │                                  │  │ Results   │  │
│                 │                                  │  └───────────┘  │
│                 │                                  │                 │
└─────────────────┘                                  └─────────────────┘
```

### Latency Calculation Process

1. **Server Side**:
   - The server generates events at a configurable rate
   - Each event includes a high-precision timestamp (`serverSendTimestampNs`) using wall clock time
   - The timestamp is generated using `Instant.now().toEpochMilli() * 1_000_000 + (System.nanoTime() % 1_000_000)` for nanosecond precision
   - Events are sent to all connected clients via WebSocket

2. **Client Side**:
   - When a client receives an event, it immediately records its own timestamp (`clientReceiveTimestampNs`)
   - The client timestamp also uses wall clock time with the same method as the server
   - One-way latency is calculated as: `latency = clientReceiveTimestampNs - serverSendTimestampNs`
   - The first N messages (configurable via `--prewarm-count`) are skipped to allow system stabilization
   - Latency statistics are calculated and displayed directly on the console

3. **Clock Synchronization**:
   - The application uses wall clock time (epoch time) on both server and client for consistent timestamps

4. **Implementation Details**:
   - Timestamps are managed by the `TimeUtils` class which provides nanosecond precision
   - The `TestMetadata` class handles the latency calculation logic
   - The `StatisticsCalculator` processes latency samples to generate percentiles and other metrics
   - All timing operations prioritize consistency between server and client over absolute precision

## Requirements

- Java 17 or higher
- Maven 3.6 or higher

## Building

```bash
# Build the application
mvn clean package

# Build the Docker image
docker build -t ws-latency-app .
```

## Running the Application

The application supports two different modes:

1. **Push Mode**: Server-push model where the server initiates messages to clients
2. **RTT Mode**: Request-response model where clients send requests and server responds

### Running in Push Mode (One-way Latency Measurement)

#### Server Side (Push Mode)

```bash
# Start the server in push mode
java -jar target/ws-latency-app-1.0.0.jar -mode=server -port=10443 -rate=10
```

Options:
- `-mode=server`: Run in server mode
- `-port`: Port for server to listen on (default: 10443)
- `-rate`: Events per second for server to send (default: 10)

Using Docker:
```bash
# Start the server in push mode
docker run -p 10443:10443 ws-latency-app -mode=server -port=10443 -rate=10
```

#### Client Side (Push Mode)

```bash
# Run the client in push mode
java -jar target/ws-latency-app-1.0.0.jar -mode=client -server=ws://localhost:10443/ws -duration=30 -prewarm-count=100
```

Options:
- `-mode=client`: Run in client mode
- `-server`: WebSocket server address (default: ws://localhost:10443/ws)
- `-duration`: Test duration in seconds (default: 30)
- `-prewarm-count`: Skip calculating latency for first N events (default: 100)
- `-insecure`: Skip TLS certificate verification (not recommended for production)
- `-continuous`: Run in continuous monitoring mode

### Running in RTT Mode (Round-Trip Time Measurement)

#### Server Side (RTT Mode)

```bash
# Start the server in RTT mode
java -jar target/ws-latency-app-1.0.0.jar -a=rtt -m=server -p=10443 --payload-size=100
```

Options:
- `-a=rtt`: Use RTT application
- `-m=server`: Run in server mode
- `-p`, `--port`: Port for server to listen on (default: 10443)
- `--payload-size`: Size of the message payload in bytes (default: 100)

Using Docker:
```bash
# Start the server in RTT mode
docker run -p 10443:10443 ws-latency-app -a=rtt -m=server -p=10443 --payload-size=100
```

#### Client Side (RTT Mode)

```bash
# Run the client in RTT mode
java -jar target/ws-latency-app-1.0.0.jar -a=rtt -m=client -s=ws://localhost:10443/ws -d=30 --rate=10 --payload-size=100 --prewarm-count=100
```

Options:
- `-a=rtt`: Use RTT application
- `-m=client`: Run in client mode
- `-s`, `--server`: WebSocket server address (default: ws://localhost:10443/ws)
- `-d`, `--duration`: Test duration in seconds (default: 30)
- `-r`, `--rate`: Requests per second for client to send (default: 10)
- `--payload-size`: Size of the message payload in bytes (default: 100)
- `--prewarm-count`: Skip calculating latency for first N events (default: 100)
- `--insecure`: Skip TLS certificate verification (not recommended for production)
- `--continuous`: Run in continuous monitoring mode

## Latency Statistics Output

The client outputs detailed latency statistics directly to the console:

```
========== LATENCY TEST RESULTS ==========
Latency Statistics (nanoseconds):
  Samples: 1000
  Min: 235000 ns
  P10: 312000 ns
  P50: 389000 ns (median)
  P90: 512000 ns
  P99: 789000 ns
  Max: 1245000 ns
  Mean: 402350.00 ns
==========================================
```

## Health Check

The server provides a health check endpoint at `http://localhost:10443/health` that returns:

```json
{
  "status": "healthy",
  "timestamp": "2025-05-23T08:01:55Z",
  "version": "1.0.0",
  "metrics": {
    "connections": 5,
    "events_sent": 1000
  }
}
```

## Docker Deployment

The application is containerized and can be deployed using Docker:

```bash
# Build the Docker image
docker build -t ws-latency-app .

# Run the server
docker run -p 10443:10443 ws-latency-app

# Run the client
docker run ws-latency-app -mode=client -server=ws://server-host:10443/ws -duration=30
```

## AWS Deployment

For AWS deployment, see the Pulumi infrastructure code in the `infrastructure` directory.

## Performance Considerations

For best results:

1. Run server and client on the same machine for minimal network overhead during testing
2. Increase process priority
3. Pin processes to specific CPU cores
4. Disable CPU frequency scaling
5. Consider using a real-time kernel

## Clock Synchronization

One-way latency measurement requires synchronized clocks between server and client. This application addresses clock synchronization challenges through:

1. **Consistent Time Source:**
   - Both server and client use the same time source method (`Instant.now().toEpochMilli() * 1_000_000 + (System.nanoTime() % 1_000_000)`)
   - Wall clock time (epoch time) is used with nanosecond precision from `System.nanoTime()`

2. **Recommended Best Practices:**
   - Ensure both server and client use NTP for clock synchronization
   - For production environments, consider implementing additional clock synchronization protocols
   - In distributed environments, monitor and account for clock drift over time

## Implementation Details

### Netty-based Architecture

The application uses Netty, a high-performance asynchronous event-driven network application framework:

- **Non-blocking I/O**: Efficient handling of many connections with minimal threads
- **Zero-copy capable byte buffers**: Reduces memory copies for better performance
- **Native transports**: Uses epoll on Linux when available for improved performance
- **Channel pipeline**: Flexible processing of WebSocket frames

### Latency Calculation

Latency is calculated with nanosecond precision:

1. Server adds timestamp to each message in `TestMetadata` using `TimeUtils.getCurrentTimeNanos()`
2. Client records receive timestamp using the same method and calculates difference
3. The `TestMetadata.calculateLatencyNs()` method computes the latency as `clientReceiveTimestampNs - serverSendTimestampNs`
4. Statistics are calculated including min/max, percentiles, and mean
5. Results are displayed directly on the console

### Recent Improvements

The application includes several recent improvements to enhance accuracy and reliability:

1. **Consistent Timestamp Generation:**
   - Both server and client use wall clock time with nanosecond precision (`Instant.now().toEpochMilli() * 1_000_000 + (System.nanoTime() % 1_000_000)`)
   - This ensures timestamps are comparable across different machines while providing nanosecond precision


3. **WebSocket Protocol Handling:**
   - Improved handling of secure WebSocket connections (wss://)
   - Added proper handling of X-Forwarded-Proto header for connections through load balancers

4. **Infrastructure Improvements:**
   - Updated ALB configuration to use HTTP for backend connections while maintaining HTTPS for client connections
   - Optimized load balancer settings for WebSocket traffic with proper stickiness configuration

5. **Nanosecond Precision:**
   - Enhanced timestamp precision from microseconds to nanoseconds
   - Implemented hybrid approach combining wall clock time with nanosecond precision from `System.nanoTime()`
   - Updated all relevant classes to handle nanosecond precision timestamps

## RTT Measurement Application

In addition to the server-push model, this project includes a separate RTT (Round-Trip Time) measurement application that implements a request-response model similar to the original Go implementation.

### RTT vs One-Way Latency

The main application measures one-way latency from server to client, which requires synchronized clocks. The RTT application measures round-trip time, which doesn't require clock synchronization:

1. **One-Way Latency (Main Application)**:
   - Server sends message with server timestamp
   - Client receives message and calculates: `clientReceiveTime - serverSendTime`
   - Requires synchronized clocks between server and client

2. **Round-Trip Time (RTT Application)**:
   - Client sends message with client timestamp
   - Server receives message, adds server timestamp, and sends it back
   - Client receives response and calculates: `clientReceiveTime - clientSendTime`
   - No clock synchronization required

### How RTT Measurement Works

```
┌─────────────────┐                                  ┌─────────────────┐
│                 │                                  │                 │
│  WebSocket      │                                  │  WebSocket      │
│  Client         │                                  │  Server         │
│                 │                                  │                 │
│  ┌───────────┐  │  1. Send request with            │  ┌───────────┐  │
│  │           │  │     client timestamp             │  │           │  │
│  │ Request   │──┼─────────────────────────────────▶│  │ WebSocket │  │
│  │ Generator │  │                                  │  │ Handler   │  │
│  │           │  │                                  │  │           │  │
│  └─────┬─────┘  │                                  │  └─────┬─────┘  │
│        │        │                                  │        │        │
│        │        │                                  │        │        │
│        ▼        │                                  │        ▼        │
│  ┌───────────┐  │                                  │  ┌───────────┐  │
│  │ WebSocket │  │  2. Receive response with        │  │ Add server│  │
│  │ Handler   │◀─┼─────server timestamp─────────────┼──│ timestamp │  │
│  │           │  │                                  │  │           │  │
│  └─────┬─────┘  │                                  │  └───────────┘  │
│        │        │                                  │                 │
│        │        │                                  │                 │
│        ▼        │                                  │                 │
│  ┌───────────┐  │                                  │                 │
│  │ Calculate │  │                                  │                 │
│  │ RTT       │  │                                  │                 │
│  └─────┬─────┘  │                                  │                 │
│        │        │                                  │                 │
│        │        │                                  │                 │
│        ▼        │                                  │                 │
│  ┌───────────┐  │                                  │                 │
│  │ Statistics│  │                                  │                 │
│  │ Calculator│  │                                  │                 │
│  └─────┬─────┘  │                                  │                 │
│        │        │                                  │                 │
│        │        │                                  │                 │
│        ▼        │                                  │                 │
│  ┌───────────┐  │                                  │                 │
│  │ Output    │  │                                  │                 │
│  │ Results   │  │                                  │                 │
│  └───────────┘  │                                  │                 │
│                 │                                  │                 │
└─────────────────┘                                  └─────────────────┘
```

### Running the RTT Server

```bash
# Start the RTT server
java -jar target/ws-latency-app-1.0.0.jar -a=rtt -m=server -p=10443 --payload-size=100
```

Options:
- `-p`, `--port`: Port for server to listen on (default: 10443)
- `--payload-size`: Size of the message payload in bytes (default: 100)

### Running the RTT Client

```bash
# Run the RTT client
java -jar target/ws-latency-app-1.0.0.jar -a=rtt -m=client -s=ws://localhost:10443/ws -d=30 --rate=10 --payload-size=100 --prewarm-count=100
```

Options:
- `-s`, `--server`: WebSocket server address (default: ws://localhost:10443/ws)
- `-d`, `--duration`: Test duration in seconds (default: 30)
- `-r`, `--rate`: Requests per second for client to send (default: 10)
- `--payload-size`: Size of the message payload in bytes (default: 100)
- `--prewarm-count`: Skip calculating latency for first N events (default: 100)
- `--insecure`: Skip TLS certificate verification (not recommended for production)
- `--continuous`: Run in continuous monitoring mode

### RTT Statistics Output

The RTT client outputs detailed round-trip time statistics directly to the console:

```
========== RTT LATENCY TEST RESULTS ==========
Latency Statistics (nanoseconds):
  Samples: 1000
  Min: 470000 ns
  P10: 624000 ns
  P50: 778000 ns (median)
  P90: 1024000 ns
  P99: 1578000 ns
  Max: 2490000 ns
  Mean: 804700.00 ns
=============================================
```

### Benefits of RTT Measurement

1. **No Clock Synchronization Required**: RTT measurement doesn't depend on synchronized clocks between server and client
2. **Accurate Measurement**: Eliminates issues with clock skew and drift
3. **Configurable Payload Size**: Allows testing with different message sizes
4. **Consistent Results**: Provides reliable measurements across different environments