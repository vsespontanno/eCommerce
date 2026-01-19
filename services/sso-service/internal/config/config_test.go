package config

import (
	"os"
	"testing"
)

func TestMustLoad_Success(t *testing.T) {
	os.Setenv("GRPC_PORT", "50051")
	os.Setenv("HTTP_PORT", "8080")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("PG_USER", "testuser")
	os.Setenv("PG_PASSWORD", "testpass")
	os.Setenv("PG_NAME", "testdb")
	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_PORT", "5432")
	defer func() {
		os.Unsetenv("GRPC_PORT")
		os.Unsetenv("HTTP_PORT")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("PG_USER")
		os.Unsetenv("PG_PASSWORD")
		os.Unsetenv("PG_NAME")
		os.Unsetenv("PG_HOST")
		os.Unsetenv("PG_PORT")
	}()

	cfg, err := MustLoad()
	if err != nil {
		t.Fatalf("MustLoad() error = %v", err)
	}

	if cfg.GRPCPort != 50051 {
		t.Errorf("GRPCPort = %v, want 50051", cfg.GRPCPort)
	}
	if cfg.HTTPPort != 8080 {
		t.Errorf("HTTPPort = %v, want 8080", cfg.HTTPPort)
	}
	if cfg.JWTSecret != "test-secret" {
		t.Errorf("JWTSecret = %v, want test-secret", cfg.JWTSecret)
	}
}

func TestMustLoad_MissingGRPCPort(t *testing.T) {
	os.Unsetenv("GRPC_PORT")

	_, err := MustLoad()
	if err == nil {
		t.Error("MustLoad() expected error for missing GRPC_PORT")
	}
}

func TestMustLoad_InvalidGRPCPort(t *testing.T) {
	os.Setenv("GRPC_PORT", "invalid")
	defer os.Unsetenv("GRPC_PORT")

	_, err := MustLoad()
	if err == nil {
		t.Error("MustLoad() expected error for invalid GRPC_PORT")
	}
}

func TestMustLoad_MissingJWTSecret(t *testing.T) {
	os.Setenv("GRPC_PORT", "50051")
	os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("GRPC_PORT")

	_, err := MustLoad()
	if err == nil {
		t.Error("MustLoad() expected error for missing JWT_SECRET")
	}
}

func TestMustLoad_MissingPGUser(t *testing.T) {
	os.Setenv("GRPC_PORT", "50051")
	os.Setenv("JWT_SECRET", "secret")
	os.Unsetenv("PG_USER")
	defer func() {
		os.Unsetenv("GRPC_PORT")
		os.Unsetenv("JWT_SECRET")
	}()

	_, err := MustLoad()
	if err == nil {
		t.Error("MustLoad() expected error for missing PG_USER")
	}
}

func TestMustLoad_DefaultHTTPPort(t *testing.T) {
	os.Setenv("GRPC_PORT", "50051")
	os.Unsetenv("HTTP_PORT")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("PG_USER", "testuser")
	os.Setenv("PG_PASSWORD", "testpass")
	os.Setenv("PG_NAME", "testdb")
	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_PORT", "5432")
	defer func() {
		os.Unsetenv("GRPC_PORT")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("PG_USER")
		os.Unsetenv("PG_PASSWORD")
		os.Unsetenv("PG_NAME")
		os.Unsetenv("PG_HOST")
		os.Unsetenv("PG_PORT")
	}()

	cfg, err := MustLoad()
	if err != nil {
		t.Fatalf("MustLoad() error = %v", err)
	}

	if cfg.HTTPPort != 8081 {
		t.Errorf("HTTPPort = %v, want 8081 (default)", cfg.HTTPPort)
	}
}
