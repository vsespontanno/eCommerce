package auth

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	proto "github.com/vsespontanno/eCommerce/proto/user"
	"github.com/vsespontanno/eCommerce/user-service/internal/models"
	"github.com/vsespontanno/eCommerce/user-service/internal/repository/pg"
)

type AuthService struct {
	proto.UnimplementedAuthServer
	pg         *pg.UserStore
	jwtService *JwtService
	tokenTTL   time.Duration
}

func NewAuthService(pg *pg.UserStore, jwtService *JwtService, tokenTTL time.Duration) *AuthService {
	return &AuthService{pg: pg, jwtService: jwtService, tokenTTL: tokenTTL}
}

func (s *AuthService) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	var params models.User
	params.Email = req.GetEmail()
	params.Password = req.GetPassword()
	params.FirstName = req.GetFirstName()
	params.LastName = req.GetLastName()
	params.ID = int64(uuid.New().ID())
	err := s.pg.SaveUser(ctx, &params)
	if err != nil {
		return nil, err
	}
	return &proto.RegisterResponse{UserId: params.ID}, nil
}

func (s *AuthService) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	var authParams models.AuthParams
	authParams.Email = req.GetEmail()
	authParams.Password = req.GetPassword()
	user, err := s.pg.GetUserByEmail(ctx, authParams.Email)
	if err != nil {
		return nil, err
	}
	if user.Password != authParams.Password {
		return nil, err
	}
	token, err := s.jwtService.GenerateToken(user, s.tokenTTL)
	if err != nil {
		log.Fatalf("failed to generate token: %v", err)
	}
	return &proto.LoginResponse{Token: token, UserId: user.ID}, nil
}
