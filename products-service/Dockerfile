# Build stage
FROM golang:1.24 AS builder

WORKDIR /app

# Copy go mod files from root
COPY go.mod go.sum ./
RUN go mod download

# Copy entire project
COPY . .

# Build the products service
WORKDIR /app/products-service
RUN go build -o /app/products-service-bin ./cmd/server/main.go

# Copy .env file
COPY products-service/.env /app/.env

# Runtime stage
FROM golang:1.24

WORKDIR /app

# Copy binary and .env from builder
COPY --from=builder /app/products-service-bin /app/products-service-bin
COPY --from=builder /app/.env /app/.env

# Expose ports: HTTP, gRPC Products, gRPC Saga
EXPOSE 8081 50052 50053

CMD ["/app/products-service-bin"]