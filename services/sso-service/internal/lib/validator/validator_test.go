package validator

import "testing"

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"valid email", "test@example.com", true},
		{"valid email with subdomain", "test@mail.example.com", true},
		{"valid email with plus", "test+tag@example.com", true},
		{"empty email", "", false},
		{"no at sign", "testexample.com", false},
		{"no domain", "test@", false},
		{"no local part", "@example.com", false},
		{"spaces", "test @example.com", false},
		{"invalid tld", "test@example.c", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateEmail(tt.email); got != tt.want {
				t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"valid password 8 chars", "password", true},
		{"valid password long", "verylongpassword123", true},
		{"too short 7 chars", "passwor", false},
		{"empty password", "", false},
		{"exactly 8 chars", "12345678", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidatePassword(tt.password); got != tt.want {
				t.Errorf("ValidatePassword(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid name", "John", true},
		{"valid name with space", "John Doe", true},
		{"too short", "J", false},
		{"empty", "", false},
		{"with numbers", "John123", false},
		{"with special chars", "John@Doe", false},
		{"exactly 2 chars", "Jo", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateName(tt.input); got != tt.want {
				t.Errorf("ValidateName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		password  string
		firstName string
		lastName  string
		want      bool
	}{
		{"all valid", "test@example.com", "password123", "John", "Doe", true},
		{"invalid email", "invalid", "password123", "John", "Doe", false},
		{"invalid password", "test@example.com", "short", "John", "Doe", false},
		{"invalid first name", "test@example.com", "password123", "J", "Doe", false},
		{"invalid last name", "test@example.com", "password123", "John", "D", false},
		{"all invalid", "", "", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateUser(tt.email, tt.password, tt.firstName, tt.lastName); got != tt.want {
				t.Errorf("ValidateUser() = %v, want %v", got, tt.want)
			}
		})
	}
}
