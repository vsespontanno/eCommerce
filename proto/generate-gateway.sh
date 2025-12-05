#!/bin/bash

set -e

# Add Go bin to PATH
export PATH=$PATH:$(go env GOPATH)/bin

echo "Generating gRPC-Gateway code..."

# Проверка protoc
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc not found. Install it first."
    exit 1
fi

# Создать директорию если не существует
mkdir -p ../internal/proto/sso

# Генерация Go кода и gRPC-Gateway
protoc -I . \
    --go_out=.. \
    --go_opt=paths=source_relative \
    --go-grpc_out=.. \
    --go-grpc_opt=paths=source_relative \
    --grpc-gateway_out=.. \
    --grpc-gateway_opt=paths=source_relative \
    --grpc-gateway_opt=generate_unbound_methods=true \
    sso/sso.proto

echo "✓ Code generated successfully!"
echo "Generated files:"
ls -la ../proto/sso/