# WebSocket Latency Testing Application (Java)

A high-performance WebSocket latency testing tool written in Java. This application measures the round-trip time (RTT) between a WebSocket client and server with precise microsecond timing.

## Features

- Request-response model for RTT measurement
- Netty-based implementation for high performance and low latency
- Configurable request rate (requests per second)
- Detailed latency statistics (min, max, P10, P50, P90, P99, mean)
- Direct console output of latency metrics
- Advanced low-latency optimizations (TCP_NODELAY, optimized buffer sizes, native epoll transport when available)
- Docker containerization

## How It Works

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

### RTT Calculation Process

1. **Client Side (Request)**:
   - The client sends requests at a configurable rate
   - Each request includes a client timestamp (`clientSendTimestampUs`) using microsecond precision
   - The timestamp is generated using a hybrid approach combining wall clock time with `System.nanoTime()`
   - Requests are sent to the server via WebSocket

2. **Server Side**:
   - When the server receives a request, it adds its own timestamp (`serverTimestampUs`)
   - The server immediately sends the message back to the client without modification
   - No processing is done on the server side to minimize latency

3. **Client Side (Response)**:
   - When the client receives a response, it immediately records its own timestamp (`clientReceiveTimestampUs`)
   - RTT is calculated as: `rtt = clientReceiveTimestampUs - clientSendTimestampUs`
   - The first N messages (configurable via `--prewarm-count`) are skipped to allow system stabilization
   - RTT statistics are calculated and displayed directly on the console

4. **Implementation Details**:
   - Timestamps are managed by the `TimeUtils` class which provides microsecond precision
   - The `RttMessage` class handles the RTT calculation logic
   - The `StatisticsCalculator` processes RTT samples to generate percentiles and other metrics
   - No clock synchronization is required since RTT is measured using only the client's clock

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

### Server Side

```bash
# Start the server
java -jar target/ws-latency-app-1.0.0.jar -m=server -p=10443 --payload-size=100
```

Options:
- `-m`, `--mode`: Run in server mode
- `-p`, `--port`: Port for server to listen on (default: 10443)
- `--payload-size`: Size of the message payload in bytes (default: 100)

Using Docker:
```bash
# Start the server
docker run -p 10443:10443 ws-latency-app -m=server -p=10443 --payload-size=100
```

### Client Side

```bash
# Run the client
java -jar target/ws-latency-app-1.0.0.jar -m=client -s=ws://localhost:10443/ws -d=30 -r=10 --payload-size=100 --prewarm-count=100
```

Options:
- `-m`, `--mode`: Run in client mode
- `-s`, `--server`: WebSocket server address (default: ws://localhost:10443/ws)
- `-d`, `--duration`: Test duration in seconds (default: 30)
- `-r`, `--rate`: Requests per second for client to send (default: 10)
- `--payload-size`: Size of the message payload in bytes (default: 100)
- `--prewarm-count`: Skip calculating latency for first N events (default: 100)
- `--insecure`: Skip TLS certificate verification (not recommended for production)
- `--continuous`: Run in continuous monitoring mode

## RTT Statistics Output

The client outputs detailed round-trip time statistics directly to the console:

```
========== RTT LATENCY TEST RESULTS ==========
Latency Statistics (microseconds):
  Samples: 1000
  Min: 470 µs
  P10: 624 µs
  P50: 778 µs (median)
  P90: 1024 µs
  P99: 1578 µs
  Max: 2490 µs
  Mean: 804.70 µs
=============================================
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

## No Clock Synchronization Required

One of the key advantages of the RTT measurement approach is that it doesn't require synchronized clocks between server and client:

1. **Single Clock Source:**
   - All timing is done using only the client's clock
   - The client measures the time from when it sends a request until it receives a response
   - This eliminates issues with clock skew and drift between different machines

2. **Microsecond Precision:**
   - Timestamps use microsecond precision for optimal balance of accuracy and performance
   - The timestamp is generated using a hybrid approach combining wall clock time with `System.nanoTime()`

## Implementation Details

### Netty-based Architecture

The application uses Netty, a high-performance asynchronous event-driven network application framework:

- **Non-blocking I/O**: Efficient handling of many connections with minimal threads
- **Zero-copy capable byte buffers**: Reduces memory copies for better performance
- **Native transports**: Uses epoll on Linux when available for improved performance
- **Channel pipeline**: Flexible processing of WebSocket frames

### Performance Optimizations

The application includes several optimizations for maximum performance:

1. **TCP_NODELAY**: Disables Nagle's algorithm to reduce latency
2. **Optimized Buffer Sizes**: Configures send/receive buffer sizes for optimal performance
3. **Write Buffer Water Marks**: Sets appropriate high/low water marks for write buffers
4. **Native Transport**: Uses epoll on Linux when available for improved performance
5. **Connection Timeouts**: Sets appropriate connection timeouts to avoid hanging connections

### RTT Calculation

RTT is calculated with microsecond precision:

1. Client adds timestamp to outgoing message using `TimeUtils.getCurrentTimeMicros()`
2. Server receives message, adds its own timestamp, and sends it back
3. Client receives response and records receive timestamp
4. The `RttMessage.calculateRttUs()` method computes RTT as `clientReceiveTimestampUs - clientSendTimestampUs`
5. Statistics are calculated including min/max, percentiles, and mean
6. Results are displayed directly on the console

## Benefits of RTT Measurement

1. **No Clock Synchronization Required**: RTT measurement doesn't depend on synchronized clocks between server and client
2. **Accurate Measurement**: Eliminates issues with clock skew and drift
3. **Configurable Payload Size**: Allows testing with different message sizes
4. **Consistent Results**: Provides reliable measurements across different environments
5. **Microsecond Precision**: Provides sufficient precision for network latency measurement
6. **Performance Optimized**: Includes multiple optimizations for minimal overhead