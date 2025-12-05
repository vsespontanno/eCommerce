# Simple Dockerfile for SSO service
FROM golang:1.24

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the application
WORKDIR /app/sso-service
RUN go build -o /app/sso-service-bin ./cmd/sso/main.go

# Expose ports
EXPOSE 50051 8080

# Run the application
CMD ["/app/sso-service-bin"]