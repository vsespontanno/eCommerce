package auth

import (
	"context"
	"errors"
	"fmt"
	"log"

	proto "github.com/vsespontanno/eCommerce/proto/sso"
	"github.com/vsespontanno/eCommerce/sso-service/internal/client"
	"github.com/vsespontanno/eCommerce/sso-service/internal/repository"
	"github.com/vsespontanno/eCommerce/sso-service/internal/services/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	proto.UnimplementedAuthServer
	auth   Auth
	client *client.WalletClient
}

type Auth interface {
	Login(ctx context.Context, email, password string) (token string, err error)
	RegisterNewUser(ctx context.Context, email, FirstName, LastName, password string) (userID int64, err error)
}

func NewAuthServer(gRPCServer *grpc.Server, auth Auth, client *client.WalletClient) {
	log.Println("Registering AuthServer")
	proto.RegisterAuthServer(gRPCServer, &AuthServer{auth: auth, client: client})
}

func (s *AuthServer) Register(ctx context.Context, in *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	log.Println("req comes here")
	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	uid, err := s.auth.RegisterNewUser(ctx, in.Email, in.Password, in.FirstName, in.LastName)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}
	fmt.Println("REGISTER COMPLETEED, NOW WE DO STUFF WITH WALLET")
	ok, _, err := s.client.CreateWallet(context.Background(), uid)
	if !ok || err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create wallet: ok = %v; err = %s", ok, err))
	}
	fmt.Println("WALLET CREATED SUCCESSFULLY")
	return &proto.RegisterResponse{UserId: uid}, nil

}

func (s *AuthServer) Login(ctx context.Context, in *proto.LoginRequest) (*proto.LoginResponse, error) {
	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	token, err := s.auth.Login(ctx, in.Email, in.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		return nil, status.Error(codes.Internal, "failed to login")
	}
	return &proto.LoginResponse{Token: token}, nil
}
