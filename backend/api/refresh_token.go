package api

import (
	"math"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Refresh token missing", http.StatusUnauthorized)
		return
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		http.Error(w, "Authentication is not configured", http.StatusInternalServerError)
		return
	}
	token, err := jwt.Parse(cookie.Value, (func(token *jwt.Token) (any, error) { return []byte(secret), nil }), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["type"] != "refresh" {
		http.Error(w, "Invalid token type", http.StatusUnauthorized)
		return
	}
	rawUserID, ok := claims["user_id"].(float64)
	if !ok || rawUserID < 1 || rawUserID != math.Trunc(rawUserID) {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}
	userID := uint(rawUserID)
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
