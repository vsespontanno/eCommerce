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
RUN CGO_ENABLED=1 go build -tags dynamic -o /app/saga-orchestrator-bin ./services/saga-orchestrator/cmd/server/main.go

EXPOSE 50051

CMD ["/app/saga-orchestrator-bin"]