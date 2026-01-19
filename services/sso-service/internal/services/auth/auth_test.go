package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vsespontanno/eCommerce/services/sso-service/internal/domain/models"
	"github.com/vsespontanno/eCommerce/services/sso-service/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type mockUserStorage struct {
	users     map[string]models.User
	saveError error
	getError  error
}

func newMockUserStorage() *mockUserStorage {
	return &mockUserStorage{
		users: make(map[string]models.User),
	}
}

func (m *mockUserStorage) SaveUser(ctx context.Context, email, firstName, lastName string, passHash []byte) (int64, error) {
	if m.saveError != nil {
		return 0, m.saveError
	}
	if _, exists := m.users[email]; exists {
		return 0, repository.ErrUserExists
	}
	id := int64(len(m.users) + 1)
	m.users[email] = models.User{
		ID:        id,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		PassHash:  passHash,
	}
	return id, nil
}

func (m *mockUserStorage) User(ctx context.Context, email string) (models.User, error) {
	if m.getError != nil {
		return models.User{}, m.getError
	}
	user, exists := m.users[email]
	if !exists {
		return models.User{}, repository.ErrUserNotFound
	}
	return user, nil
}

func newTestLogger() *zap.SugaredLogger {
	logger, _ := zap.NewDevelopment()
	return logger.Sugar()
}

func TestAuth_RegisterNewUser_Success(t *testing.T) {
	storage := newMockUserStorage()
	auth := NewAuth(newTestLogger(), storage, time.Hour, "secret")

	id, err := auth.RegisterNewUser(context.Background(), "test@example.com", "password123", "John", "Doe")
	if err != nil {
		t.Fatalf("RegisterNewUser() error = %v", err)
	}

	if id != 1 {
		t.Errorf("RegisterNewUser() id = %v, want 1", id)
	}
}

func TestAuth_RegisterNewUser_InvalidInput(t *testing.T) {
	storage := newMockUserStorage()
	auth := NewAuth(newTestLogger(), storage, time.Hour, "secret")

	tests := []struct {
		name      string
		email     string
		password  string
		firstName string
		lastName  string
	}{
		{"invalid email", "invalid", "password123", "John", "Doe"},
		{"short password", "test@example.com", "short", "John", "Doe"},
		{"short first name", "test@example.com", "password123", "J", "Doe"},
		{"short last name", "test@example.com", "password123", "John", "D"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := auth.RegisterNewUser(context.Background(), tt.email, tt.password, tt.firstName, tt.lastName)
			if err == nil {
				t.Error("RegisterNewUser() expected error")
			}
			if !errors.Is(err, ErrInvalidInput) {
				t.Errorf("RegisterNewUser() error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestAuth_RegisterNewUser_UserExists(t *testing.T) {
	storage := newMockUserStorage()
	auth := NewAuth(newTestLogger(), storage, time.Hour, "secret")

	_, err := auth.RegisterNewUser(context.Background(), "test@example.com", "password123", "John", "Doe")
	if err != nil {
		t.Fatalf("First RegisterNewUser() error = %v", err)
	}

	_, err = auth.RegisterNewUser(context.Background(), "test@example.com", "password123", "Jane", "Doe")
	if err == nil {
		t.Error("Second RegisterNewUser() expected error")
	}
	if !errors.Is(err, repository.ErrUserExists) {
		t.Errorf("RegisterNewUser() error = %v, want ErrUserExists", err)
	}
}

func TestAuth_Login_Success(t *testing.T) {
	storage := newMockUserStorage()
	auth := NewAuth(newTestLogger(), storage, time.Hour, "secret")

	_, err := auth.RegisterNewUser(context.Background(), "test@example.com", "password123", "John", "Doe")
	if err != nil {
		t.Fatalf("RegisterNewUser() error = %v", err)
	}

	token, expiresAt, err := auth.Login(context.Background(), "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if token == "" {
		t.Error("Login() returned empty token")
	}

	if expiresAt == 0 {
		t.Error("Login() returned zero expiresAt")
	}
}

func TestAuth_Login_UserNotFound(t *testing.T) {
	storage := newMockUserStorage()
	auth := NewAuth(newTestLogger(), storage, time.Hour, "secret")

	_, _, err := auth.Login(context.Background(), "nonexistent@example.com", "password123")
	if err == nil {
		t.Error("Login() expected error")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuth_Login_WrongPassword(t *testing.T) {
	storage := newMockUserStorage()
	auth := NewAuth(newTestLogger(), storage, time.Hour, "secret")

	_, err := auth.RegisterNewUser(context.Background(), "test@example.com", "password123", "John", "Doe")
	if err != nil {
		t.Fatalf("RegisterNewUser() error = %v", err)
	}

	_, _, err = auth.Login(context.Background(), "test@example.com", "wrongpassword")
	if err == nil {
		t.Error("Login() expected error")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuth_Login_StorageError(t *testing.T) {
	storage := newMockUserStorage()
	storage.getError = errors.New("database error")
	auth := NewAuth(newTestLogger(), storage, time.Hour, "secret")

	_, _, err := auth.Login(context.Background(), "test@example.com", "password123")
	if err == nil {
		t.Error("Login() expected error")
	}
}

func TestAuth_RegisterNewUser_StorageError(t *testing.T) {
	storage := newMockUserStorage()
	storage.saveError = errors.New("database error")
	auth := NewAuth(newTestLogger(), storage, time.Hour, "secret")

	_, err := auth.RegisterNewUser(context.Background(), "test@example.com", "password123", "John", "Doe")
	if err == nil {
		t.Error("RegisterNewUser() expected error")
	}
}

func TestAuth_Login_WithPrehashedPassword(t *testing.T) {
	storage := newMockUserStorage()
	auth := NewAuth(newTestLogger(), storage, time.Hour, "secret")

	password := "password123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	storage.users["test@example.com"] = models.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		PassHash:  hash,
	}

	token, _, err := auth.Login(context.Background(), "test@example.com", password)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if token == "" {
		t.Error("Login() returned empty token")
	}
}
