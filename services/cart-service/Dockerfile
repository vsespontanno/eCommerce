FROM golang:1.24

# Install librdkafka for ARM64
RUN apt-get update && apt-get install -y \
    librdkafka-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application with CGO enabled
WORKDIR /app
RUN CGO_ENABLED=1 go build -tags dynamic -o /app/cart-service-bin ./services/cart-service/cmd/server/main.go

EXPOSE 8080

CMD ["/app/cart-service-bin"]