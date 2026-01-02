FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app
RUN go build -o /app/products-service-bin ./services/products-service/cmd/server/main.go

FROM golang:1.23

WORKDIR /app

COPY --from=builder /app/products-service-bin /app/products-service-bin

EXPOSE 8080 50051 50052

CMD ["/app/products-service-bin"]