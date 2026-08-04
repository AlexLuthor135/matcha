package user

import (
	"strings"
	"unicode"
)

func isValidGenderPreferences(gender string) bool {
	switch gender {
	case "Male", "Female", "Other":
		return true
	default:
		return false
	}
}

func isValidPassword(password string) bool {
	if len(password) < 8 || len([]byte(password)) > 72 {
		return false
	}

	var hasUpper bool
	var hasLower bool
	var hasNumber bool
	var hasSpecial bool

	for _, character := range password {
		switch {
		case unicode.IsSpace(character):
			return false
		case unicode.IsUpper(character):
			hasUpper = true
		case unicode.IsLower(character):
			hasLower = true
		case unicode.IsDigit(character):
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*", character):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasNumber && hasSpecial
}
