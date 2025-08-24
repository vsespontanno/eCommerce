package interceptor

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (userID int64, err error)
}

type contextKey string

const UserIDKey contextKey = "user_id"

// UnaryServerInterceptor returns a grpc.UnaryServerInterceptor that validates Bearer token
func AuthInterceptor(validator TokenValidator) grpc.UnaryServerInterceptor {
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
		userID, err := validator.ValidateToken(ctx, token)
		if err != nil || userID == 0 {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		// inject user id into context
		ctx = context.WithValue(ctx, UserIDKey, userID)
		return handler(ctx, req)
	}
}
