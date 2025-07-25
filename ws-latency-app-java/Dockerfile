# Build stage
FROM --platform=${BUILDPLATFORM:-linux/amd64} maven:3.8-openjdk-17 AS build
WORKDIR /app
COPY pom.xml .
RUN mvn dependency:go-offline
COPY src ./src
RUN mvn package -DskipTests

# Runtime stage
FROM --platform=${TARGETPLATFORM:-linux/amd64} amazoncorretto:17-alpine
WORKDIR /app
COPY --from=build /app/target/ws-latency-app-*.jar ./ws-latency-app.jar

# Expose ports
# - 10443: WebSocket server port (also serves health check endpoint)
EXPOSE 10443

# Set environment variables with default values
ENV MODE=server
ENV PORT=10443
ENV EVENTS_PER_SECOND=10

# Create a startup script with explicit argument handling
RUN echo '#!/bin/sh' > /app/start.sh && \
    echo 'java -jar /app/ws-latency-app.jar -m $MODE -p $PORT -r $EVENTS_PER_SECOND "$@"' >> /app/start.sh && \
    chmod +x /app/start.sh

# Set entry point to the startup script
ENTRYPOINT ["/app/start.sh"]