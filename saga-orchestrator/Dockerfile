# Simple Dockerfile for Saga Orchestrator service
FROM golang:1.23

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the application
WORKDIR /app/saga-orchestrator
RUN go build -o /app/saga-orchestrator-bin ./cmd/server/main.go

# Expose gRPC port
EXPOSE 50054

# Run the application
CMD ["/app/saga-orchestrator-bin"]