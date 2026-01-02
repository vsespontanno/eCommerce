FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app
RUN go build -o /app/sso-service-bin ./services/sso-service/cmd/server/main.go

EXPOSE 50051 8080

CMD ["/app/sso-service-bin"]