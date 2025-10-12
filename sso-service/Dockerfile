FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY proto/ ./proto/
COPY sso-service/ ./sso-service/

WORKDIR /app/sso-service
RUN go build -o sso-service ./cmd/sso

EXPOSE 50051
CMD ["./sso-service"]