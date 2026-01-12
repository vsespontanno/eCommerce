FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app
RUN go build -o /app/wallet-service-bin ./services/wallet-service/cmd/server/main.go

EXPOSE 50051 50054 8080

CMD ["/app/wallet-service-bin"]