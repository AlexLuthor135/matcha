package account

import (
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"
)

func isValidUserName(userName string) bool {
	length := utf8.RuneCountInString(userName)
	if length < 3 || length > 30 {
		return false
	}
	for _, character := range userName {
		if unicode.IsLetter(character) || unicode.IsDigit(character) ||
			character == '_' || character == '-' || character == '.' {
			continue
		}
		return false
	}
	return true
}

func isValidPassword(password string) bool {
	passwordLength := utf8.RuneCountInString(password)
	if passwordLength < 12 || len(password) > 72 || isCommonPassword(password) {
		return false
	}
	var hasUpper, hasLower, hasNumber, hasSpecial bool
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

func isValidEmail(email string) bool {
	if len(email) > 254 || strings.ContainsAny(email, "\r\n") {
		return false
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return false
	}
	atIndex := strings.LastIndexByte(email, '@')
	if atIndex <= 0 || atIndex == len(email)-1 {
		return false
	}
	return strings.Contains(email[atIndex+1:], ".")
}
