package user

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

func generateAccountToken() (string, string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", err
	}
	rawToken := base64.RawURLEncoding.EncodeToString(randomBytes)
	tokenHash := hashAccountToken(rawToken)
	return rawToken, tokenHash, nil
}

func hashAccountToken(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hash[:])
}
