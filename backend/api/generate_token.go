package api

import (
	"backend/models"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"os"
	"time"
)

type LoginResponse struct {
	Message     string `json:"message"`
	IsCompleted bool   `json:"is_completed"`
}

func GenerateTokens(w http.ResponseWriter, user models.User, sessionID string, refreshExpiresAt time.Time) bool {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		http.Error(w, "Authentication is not configured", http.StatusInternalServerError)
		return true
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"type":    "access",
		"exp":     time.Now().Add(time.Minute * 15).Unix(),
	})
	accessTokenString, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		http.Error(w, "Error generating access token", http.StatusInternalServerError)
		return true
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    user.ID,
		"session_id": sessionID,
		"type":       "refresh",
		"exp":        refreshExpiresAt.Unix(),
	})
	refreshTokenString, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		http.Error(w, "Error generating refresh token", http.StatusInternalServerError)
		return true
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessTokenString,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		Expires:  time.Now().Add(time.Minute * 15),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshTokenString,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		Expires:  refreshExpiresAt,
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := LoginResponse{
		Message:     "Login successful",
		IsCompleted: user.IsCompleted,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode login response for user %d: %v", user.ID, err)
	}
	return false

}

func ClearAuthCookies(w http.ResponseWriter) {
	for _, name := range []string{
		"access_token", "refresh_token",
	} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
		})
	}
}
