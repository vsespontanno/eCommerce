package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/vsespontanno/eCommerce/sso-service/internal/domain/models"
	"github.com/vsespontanno/eCommerce/sso-service/internal/lib/jwt"
	"github.com/vsespontanno/eCommerce/sso-service/internal/lib/logger/sl"
	"github.com/vsespontanno/eCommerce/sso-service/internal/lib/validator"
	"github.com/vsespontanno/eCommerce/sso-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserStorage interface {
	SaveUser(ctx context.Context, email, FirstName, LastName string, passHash []byte) (uid int64, err error)
	User(ctx context.Context, email string) (models.User, error)
}

type Auth struct {
	log         *slog.Logger
	userStorage UserStorage
	tokenTTL    time.Duration
	jwtSecret   string
}

func NewAuth(log *slog.Logger, userStorage UserStorage, tokenTTL time.Duration, jwtSecret string) *Auth {
	return &Auth{
		log:         log,
		userStorage: userStorage,
		tokenTTL:    tokenTTL,
		jwtSecret:   jwtSecret,
	}
}

func (a *Auth) RegisterNewUser(ctx context.Context, email, password, FirstName, LastName string) (userID int64, err error) {
	// op (operation) - имя текущей функции и пакета. Такую метку удобно
	// добавлять в логи и в текст ошибок, чтобы легче было искать хвосты
	// в случае поломок.
	const op = "Auth.RegisterNewUser"
	log := a.log.With(slog.String("op", op), slog.String("email", email))
	log.Info("registering user")

	// Validate user input
	if !validator.ValidateUser(email, password, FirstName, LastName) {
		log.Info("invalid user input")
		return 0, ErrInvalidCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	id, err := a.userStorage.SaveUser(ctx, email, FirstName, LastName, hash)
	if err != nil {
		log.Error("failed to save user", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (a *Auth) Login(ctx context.Context, email, password string) (token string, err error) {
	const op = "Auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("username", email),
		// password либо не логируем, либо логируем в замаскированном виде
	)

	log.Info("attempting to login user")

	user, err := a.userStorage.User(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			log.Info("user not found")
			return "", ErrInvalidCredentials
		}
		log.Error("failed to get user", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid creds", sl.Err(err))
		return "", ErrInvalidCredentials
	}

	log.Info("user logged in successfully")

	// Generate a new token for the user
	token, err = jwt.NewToken(user, a.jwtSecret, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate token", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}
