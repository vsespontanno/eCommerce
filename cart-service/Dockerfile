# Simple Dockerfile for Cart service
FROM golang:1.23

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the application
WORKDIR /app/cart-service
RUN go build -o /app/cart-service-bin ./cmd/server/main.go

# Expose HTTP port
EXPOSE 8082

# Run the application
CMD ["/app/cart-service-bin"]