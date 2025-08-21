package validator

import (
	"regexp"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	nameRegex  = regexp.MustCompile(`^[a-zA-Z\s]+$`)
)

func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	if password == "password" {
		return false
	}
	if password == "12345678" {
		return false
	}
	if password == "QWERTY123" {
		return false
	}
	return true
}

func ValidateName(name string) bool {
	if len(name) < 2 || len(name) > 100 {
		return false
	} else if !nameRegex.MatchString(name) {
		return false
	}
	return true
}

func ValidateUser(email, password, FirstName, LastName string) bool {
	return ValidateEmail(email) && ValidatePassword(password) && ValidateName(FirstName) && ValidateName(LastName)
}
