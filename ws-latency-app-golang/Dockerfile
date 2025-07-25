# Multi-stage build for ws-latency-app-golang

# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o ws-latency-app ./cmd/ws-latency-app/main.go

# Final stage
FROM alpine:latest

# Add CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/ws-latency-app .

# Expose the default port
EXPOSE 10443

# Set the entrypoint
ENTRYPOINT ["./ws-latency-app"]

# Default command (can be overridden)
CMD ["-m", "server", "-a", "push", "-p", "10443", "-r", "10"]