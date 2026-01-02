FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app
RUN go build -o /app/order-service-bin ./services/order-service/cmd/server/main.go

EXPOSE 50051

CMD ["/app/order-service-bin"]