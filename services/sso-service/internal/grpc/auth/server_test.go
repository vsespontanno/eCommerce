package auth

import (
	"context"
	"errors"
	"testing"

	proto "github.com/vsespontanno/eCommerce/proto/sso"
	"github.com/vsespontanno/eCommerce/services/sso-service/internal/repository"
	"github.com/vsespontanno/eCommerce/services/sso-service/internal/services/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockAuth struct {
	loginToken     string
	loginExpiresAt int64
	loginErr       error
	registerID     int64
	registerErr    error
}

func (m *mockAuth) Login(ctx context.Context, email, password string) (string, int64, error) {
	return m.loginToken, m.loginExpiresAt, m.loginErr
}

func (m *mockAuth) RegisterNewUser(ctx context.Context, email, password, firstName, lastName string) (int64, error) {
	return m.registerID, m.registerErr
}

func newTestLogger() *zap.SugaredLogger {
	logger, _ := zap.NewDevelopment()
	return logger.Sugar()
}

func TestServer_Register_Success(t *testing.T) {
	mockAuth := &mockAuth{registerID: 1}
	server := &Server{auth: mockAuth, log: newTestLogger()}

	resp, err := server.Register(context.Background(), &proto.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	})

	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if resp.UserId != 1 {
		t.Errorf("Register() UserId = %v, want 1", resp.UserId)
	}
}

func TestServer_Register_EmptyEmail(t *testing.T) {
	mockAuth := &mockAuth{}
	server := &Server{auth: mockAuth, log: newTestLogger()}

	_, err := server.Register(context.Background(), &proto.RegisterRequest{
		Email:     "",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	})

	if err == nil {
		t.Fatal("Register() expected error")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("Register() code = %v, want InvalidArgument", st.Code())
	}
}

func TestServer_Register_EmptyPassword(t *testing.T) {
	mockAuth := &mockAuth{}
	server := &Server{auth: mockAuth, log: newTestLogger()}

	_, err := server.Register(context.Background(), &proto.RegisterRequest{
		Email:     "test@example.com",
		Password:  "",
		FirstName: "John",
		LastName:  "Doe",
	})

	if err == nil {
		t.Fatal("Register() expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("Register() code = %v, want InvalidArgument", st.Code())
	}
}

func TestServer_Register_EmptyFirstName(t *testing.T) {
	mockAuth := &mockAuth{}
	server := &Server{auth: mockAuth, log: newTestLogger()}

	_, err := server.Register(context.Background(), &proto.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "",
		LastName:  "Doe",
	})

	if err == nil {
		t.Fatal("Register() expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("Register() code = %v, want InvalidArgument", st.Code())
	}
}

func TestServer_Register_EmptyLastName(t *testing.T) {
	mockAuth := &mockAuth{}
	server := &Server{auth: mockAuth, log: newTestLogger()}

	_, err := server.Register(context.Background(), &proto.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "",
	})

	if err == nil {
		t.Fatal("Register() expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("Register() code = %v, want InvalidArgument", st.Code())
	}
}

func TestServer_Register_UserExists(t *testing.T) {
	mockAuth := &mockAuth{registerErr: repository.ErrUserExists}
	server := &Server{auth: mockAuth, log: newTestLogger()}

	_, err := server.Register(context.Background(), &proto.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	})

	if err == nil {
		t.Fatal("Register() expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.AlreadyExists {
		t.Errorf("Register() code = %v, want AlreadyExists", st.Code())
	}
}

func TestServer_Register_InvalidInput(t *testing.T) {
	mockAuth := &mockAuth{registerErr: auth.ErrInvalidInput}
	server := &Server{auth: mockAuth, log: newTestLogger()}

	_, err := server.Register(context.Background(), &proto.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	})

	if err == nil {
		t.Fatal("Register() expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("Register() code = %v, want InvalidArgument", st.Code())
	}
}

func TestServer_Register_InternalError(t *testing.T) {
	mockAuth := &mockAuth{registerErr: errors.New("database error")}
	server := &Server{auth: mockAuth, log: newTestLogger()}

	_, err := server.Register(context.Background(), &proto.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	})

	if err == nil {
		t.Fatal("Register() expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Errorf("Register() code = %v, want Internal", st.Code())
	}
}

func TestServer_Login_Success(t *testing.T) {
	mockAuth := &mockAuth{loginToken: "test-token", loginExpiresAt: 1234567890}
	server := &Server{auth: mockAuth, log: newTestLogger()}

	resp, err := server.Login(context.Background(), &proto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if resp.Token != "test-token" {
		t.Errorf("Login() Token = %v, want test-token", resp.Token)
	}

	if resp.ExpiresAt != 1234567890 {
		t.Errorf("Login() ExpiresAt = %v, want 1234567890", resp.ExpiresAt)
	}
}

func TestServer_Login_EmptyEmail(t *testing.T) {
	mockAuth := &mockAuth{}
	server := &Server{auth: mockAuth, log: newTestLogger()}

	_, err := server.Login(context.Background(), &proto.LoginRequest{
		Email:    "",
		Password: "password123",
	})

	if err == nil {
		t.Fatal("Login() expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("Login() code = %v, want InvalidArgument", st.Code())
	}
}

func TestServer_Login_EmptyPassword(t *testing.T) {
	mockAuth := &mockAuth{}
	server := &Server{auth: mockAuth, log: newTestLogger()}

	_, err := server.Login(context.Background(), &proto.LoginRequest{
		Email:    "test@example.com",
		Password: "",
	})

	if err == nil {
		t.Fatal("Login() expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("Login() code = %v, want InvalidArgument", st.Code())
	}
}

func TestServer_Login_InvalidCredentials(t *testing.T) {
	mockAuth := &mockAuth{loginErr: auth.ErrInvalidCredentials}
	server := &Server{auth: mockAuth, log: newTestLogger()}

	_, err := server.Login(context.Background(), &proto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})

	if err == nil {
		t.Fatal("Login() expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Errorf("Login() code = %v, want Unauthenticated", st.Code())
	}
}

func TestServer_Login_InternalError(t *testing.T) {
	mockAuth := &mockAuth{loginErr: errors.New("database error")}
	server := &Server{auth: mockAuth, log: newTestLogger()}

	_, err := server.Login(context.Background(), &proto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	if err == nil {
		t.Fatal("Login() expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Errorf("Login() code = %v, want Internal", st.Code())
	}
}
