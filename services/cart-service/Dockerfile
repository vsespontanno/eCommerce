# Simple Dockerfile for Cart service
FROM golang:1.24

# Install librdkafka for ARM64
RUN apt-get update && apt-get install -y \
    librdkafka-dev \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the application with CGO enabled
WORKDIR /app/cart-service
RUN CGO_ENABLED=1 go build -tags dynamic -o /app/cart-service-bin ./cmd/server/main.go

# Expose HTTP port
EXPOSE 8082

# Run the application
CMD ["/app/cart-service-bin"]