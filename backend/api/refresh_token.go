package api

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	ErrInvalidRefreshToken         = errors.New("invalid refresh token")
	ErrAuthenticationNotConfigured = errors.New("authentication is not configured")
)

type RefreshTokenClaims struct {
	UserID    uint   `json:"user_id"`
	SessionID string `json:"session_id"`
	TokenType string `json:"type"`
	jwt.RegisteredClaims
}

func ParseRefreshToken(rawToken string) (RefreshTokenClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return RefreshTokenClaims{}, ErrAuthenticationNotConfigured
	}
	var claims RefreshTokenClaims
	token, err := jwt.ParseWithClaims(rawToken, &claims, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid || claims.TokenType != "refresh" || claims.UserID == 0 || strings.TrimSpace(claims.SessionID) == "" {
		return RefreshTokenClaims{}, ErrInvalidRefreshToken
	}
	return claims, nil
}

type RefreshSessionValidator interface {
	ValidateAuthSession(ctx context.Context, sessionID string, userID uint) (bool, error)
}

func RefreshToken(validator RefreshSessionValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			http.Error(w, "Refresh token missing", http.StatusUnauthorized)
			return
		}
		claims, err := ParseRefreshToken(cookie.Value)
		if errors.Is(err, ErrAuthenticationNotConfigured) {
			http.Error(w, "Authentication is not configured", http.StatusInternalServerError)
			return
		}
		if err != nil {
			http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
			return
		}
		userID := claims.UserID
		sessionID := claims.SessionID
		secret := os.Getenv("JWT_SECRET")
		validSession, err := validator.ValidateAuthSession(r.Context(), sessionID, userID)
		if err != nil {
			log.Printf("Validate refresh session for user %d: %v", userID, err)
			http.Error(w, "Failed to refresh token", http.StatusInternalServerError)
			return
		}
		if !validSession {
			http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
			return
		}
		expiresAt := time.Now().Add(15 * time.Minute)
		accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": userID,
			"type":    "access",
			"exp":     expiresAt.Unix(),
		})
		accessTokenString, err := accessToken.SignedString([]byte(secret))
		if err != nil {
			http.Error(w, "Failed to refresh access token", http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    accessTokenString,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			Expires:  expiresAt,
			MaxAge:   int((15 * time.Minute).Seconds()),
		})
		w.WriteHeader(http.StatusNoContent)
	}
}
