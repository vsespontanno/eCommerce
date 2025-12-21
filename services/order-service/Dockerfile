# Simple Dockerfile for Order service
FROM golang:1.24

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the application
WORKDIR /app
RUN go build -o /app/order-service-bin ./services/order-service/cmd/server/main.go

# Expose gRPC port
EXPOSE 50051

# Run the application
CMD ["/app/order-service-bin"]