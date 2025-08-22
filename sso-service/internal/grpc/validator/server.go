package validator

import (
	"context"
	"log"

	proto "github.com/vsespontanno/eCommerce/proto/sso"
	"google.golang.org/grpc"
)

type ValidationServer struct {
	proto.UnimplementedValidatorServer
	validator Validator
}

type Validator interface {
	ValidateJWTToken(token string) (valid bool, userID float64, err error)
}

func NewValidationServer(gRPCServer *grpc.Server, validator Validator) {
	log.Println("Registering ValidationServer")
	proto.RegisterValidatorServer(gRPCServer, &ValidationServer{validator: validator})
}

func (s *ValidationServer) ValidateToken(ctx context.Context, in *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	log.Printf("Validating token: %s", in.Token)

	valid, userID, err := s.validator.ValidateJWTToken(in.Token)

	if err != nil {
		return nil, err
	}

	return &proto.ValidateTokenResponse{Valid: valid, UserId: int64(userID)}, nil
}
