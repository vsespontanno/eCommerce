package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/vsespontanno/eCommerce/services/sso-service/internal/domain/models"
	"github.com/vsespontanno/eCommerce/services/sso-service/internal/lib/jwt"
	"github.com/vsespontanno/eCommerce/services/sso-service/internal/lib/validator"
	"github.com/vsespontanno/eCommerce/services/sso-service/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidInput       = errors.New("invalid input")
)

type UserStorage interface {
	SaveUser(ctx context.Context, email, firstName, lastName string, passHash []byte) (int64, error)
	User(ctx context.Context, email string) (models.User, error)
}

type Auth struct {
	log         *zap.SugaredLogger
	userStorage UserStorage
	tokenTTL    time.Duration
	jwtSecret   string
}

func NewAuth(log *zap.SugaredLogger, userStorage UserStorage, tokenTTL time.Duration, jwtSecret string) *Auth {
	return &Auth{
		log:         log,
		userStorage: userStorage,
		tokenTTL:    tokenTTL,
		jwtSecret:   jwtSecret,
	}
}

func (a *Auth) RegisterNewUser(ctx context.Context, email, password, firstName, lastName string) (int64, error) {
	const op = "auth.RegisterNewUser"

	a.log.Infow("registering new user", "op", op, "email", email)

	// Validate input
	if !validator.ValidateUser(email, password, firstName, lastName) {
		a.log.Warnw("invalid user input", "op", op, "email", email)
		return 0, fmt.Errorf("%s: %w", op, ErrInvalidInput)
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.log.Errorw("failed to hash password", "op", op, "error", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// Save user in DB
	id, err := a.userStorage.SaveUser(ctx, email, firstName, lastName, hash)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			a.log.Warnw("user already exists", "op", op, "email", email)
			return 0, fmt.Errorf("%s: %w", op, repository.ErrUserExists)
		}
		a.log.Errorw("failed to save user", "op", op, "email", email, "error", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	a.log.Infow("user registered successfully", "op", op, "user_id", id, "email", email)
	return id, nil
}

func (a *Auth) Login(ctx context.Context, email, password string) (string, int64, error) {
	const op = "auth.Login"

	a.log.Infow("attempting login", "op", op, "email", email)

	user, err := a.userStorage.User(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			a.log.Warnw("user not found", "op", op, "email", email)
			return "", 0, ErrInvalidCredentials
		}
		a.log.Errorw("failed to fetch user from storage", "op", op, "email", email, "error", err)
		return "", 0, fmt.Errorf("%s: %w", op, err)
	}

	// Compare password
	if err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Warnw("invalid password", "op", op, "email", email)
		return "", 0, ErrInvalidCredentials
	}

	// Generate JWT with expiry
	tokenPair, err := jwt.NewTokenWithExpiry(user, a.jwtSecret, a.tokenTTL)
	if err != nil {
		a.log.Errorw("failed to generate JWT token", "op", op, "email", email, "error", err)
		return "", 0, fmt.Errorf("%s: %w", op, err)
	}

	a.log.Infow("user logged in successfully", "op", op, "user_id", user.ID, "email", email)
	return tokenPair.Token, tokenPair.ExpiresAt, nil
}
