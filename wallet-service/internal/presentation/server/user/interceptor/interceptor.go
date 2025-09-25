package interceptor

import (
	"context"
	"strings"

	"github.com/vsespontanno/eCommerce/wallet-service/internal/domain/wallet/entity/keys"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if info.FullMethod == "/proto_wallet.Creator/CreateWallet" {
			return handler(ctx, req)
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		authHeaders := md["authorization"]
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}
		authHeader := authHeaders[0]
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}
		token := parts[1]

		ctx = context.WithValue(ctx, keys.JwtKey, token)
		return handler(ctx, req)
	}
}
