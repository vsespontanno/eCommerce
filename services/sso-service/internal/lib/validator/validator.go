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
	return len(password) >= 8
}

func ValidateName(name string) bool {
	if len(name) < 2 || len(name) > 100 {
		return false
	} else if !nameRegex.MatchString(name) {
		return false
	}
	return true
}

func ValidateUser(email, password, firstName, lastName string) bool {
	return ValidateEmail(email) && ValidatePassword(password) && ValidateName(firstName) && ValidateName(lastName)
}
