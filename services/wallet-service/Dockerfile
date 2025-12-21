# Simple Dockerfile for Wallet service
FROM golang:1.24

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the application
WORKDIR /app/wallet-service
RUN go build -o /app/wallet-service-bin ./cmd/server/main.go

# Expose ports (gRPC user, gRPC saga, HTTP gateway)
EXPOSE 50050 50054 8080

# Run the application
CMD ["/app/wallet-service-bin"]