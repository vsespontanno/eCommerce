package validator

import (
	"context"
	"errors"

	proto "github.com/vsespontanno/eCommerce/proto/sso"
	validatorService "github.com/vsespontanno/eCommerce/services/sso-service/internal/services/validator"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ValidationServer struct {
	proto.UnimplementedValidatorServer
	validator Validator
	log       *zap.SugaredLogger
}

type Validator interface {
	ValidateJWTToken(token string) (valid bool, userID float64, err error)
}

func NewValidationServer(gRPCServer *grpc.Server, validator Validator, log *zap.SugaredLogger) {
	log.Info("Registering ValidationServer")
	proto.RegisterValidatorServer(gRPCServer, &ValidationServer{
		validator: validator,
		log:       log,
	})
}

func (s *ValidationServer) ValidateToken(ctx context.Context, in *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	const op = "ValidationServer.ValidateToken"

	if in.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	valid, userID, err := s.validator.ValidateJWTToken(in.Token)
	if err != nil {
		if errors.Is(err, validatorService.ErrInvalidToken) {
			s.log.Warnw("invalid token", "op", op)
			return &proto.ValidateTokenResponse{
				Valid:  false,
				UserId: 0,
				Email:  "",
			}, nil
		}
		if errors.Is(err, validatorService.ErrTokenExpired) {
			s.log.Warnw("token expired", "op", op)
			return &proto.ValidateTokenResponse{
				Valid:  false,
				UserId: 0,
				Email:  "",
			}, nil
		}

		s.log.Errorw("failed to validate token", "op", op, "error", err)
		return nil, status.Error(codes.Internal, "failed to validate token")
	}

	return &proto.ValidateTokenResponse{
		Valid:  valid,
		UserId: int64(userID),
		Email:  "",
	}, nil
}
