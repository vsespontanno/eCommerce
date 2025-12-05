package auth

import (
	"context"
	"errors"

	proto "github.com/vsespontanno/eCommerce/proto/sso"
	"github.com/vsespontanno/eCommerce/sso-service/internal/repository"
	"github.com/vsespontanno/eCommerce/sso-service/internal/services/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	proto.UnimplementedAuthServer
	auth Auth
	log  *zap.SugaredLogger
}

type Auth interface {
	Login(ctx context.Context, email, password string) (token string, expiresAt int64, err error)
	RegisterNewUser(ctx context.Context, email, password, firstName, lastName string) (userID int64, err error)
}

func NewAuthServer(gRPCServer *grpc.Server, auth Auth, log *zap.SugaredLogger) {
	log.Info("Registering AuthServer")
	proto.RegisterAuthServer(gRPCServer, &AuthServer{
		auth: auth,
		log:  log,
	})
}

func (s *AuthServer) Register(ctx context.Context, in *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	const op = "AuthServer.Register"

	s.log.Infow("register request received", "op", op, "email", in.Email)

	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	if in.FirstName == "" {
		return nil, status.Error(codes.InvalidArgument, "first_name is required")
	}

	if in.LastName == "" {
		return nil, status.Error(codes.InvalidArgument, "last_name is required")
	}

	uid, err := s.auth.RegisterNewUser(ctx, in.Email, in.Password, in.FirstName, in.LastName)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			s.log.Warnw("user already exists", "op", op, "email", in.Email)
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		if errors.Is(err, auth.ErrInvalidInput) {
			s.log.Warnw("invalid input", "op", op, "email", in.Email)
			return nil, status.Error(codes.InvalidArgument, "invalid input data")
		}

		s.log.Errorw("failed to register user", "op", op, "error", err)
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	s.log.Infow("user registered successfully", "op", op, "user_id", uid)
	return &proto.RegisterResponse{UserId: uid}, nil
}

func (s *AuthServer) Login(ctx context.Context, in *proto.LoginRequest) (*proto.LoginResponse, error) {
	const op = "AuthServer.Login"

	s.log.Infow("login request received", "op", op, "email", in.Email)

	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	token, expiresAt, err := s.auth.Login(ctx, in.Email, in.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			s.log.Warnw("invalid credentials", "op", op, "email", in.Email)
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}

		s.log.Errorw("failed to login", "op", op, "error", err)
		return nil, status.Error(codes.Internal, "failed to login")
	}

	s.log.Infow("user logged in successfully", "op", op, "email", in.Email)
	return &proto.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}
