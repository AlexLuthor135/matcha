package user

import (
	_ "embed"
	"strings"
)

//go:embed common_passwords.txt
var commonPasswordsData string

var commonPasswords = func() map[string]struct{} {
	passwords := make(map[string]struct{})
	for _, password := range strings.Split(commonPasswordsData, "\n") {
		password = strings.ToLower(strings.TrimSpace(password))
		if password != "" {
			passwords[password] = struct{}{}
		}
	}
	return passwords
}()

func isCommonPassword(password string) bool {
	normalizedPassword := strings.ToLower(strings.TrimSpace(password))
	_, exists := commonPasswords[normalizedPassword]
	return exists
}
