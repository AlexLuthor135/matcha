package middleware

import (
	"context"
	"math"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = contextKey("userID")

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			http.Error(w, "Authentication is not configured", http.StatusInternalServerError)
			return
		}

		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (any, error) { return []byte(secret), nil }, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["type"] != "access" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		rawUserID, ok := claims["user_id"].(float64)
		if !ok || rawUserID < 1 || rawUserID != math.Trunc(rawUserID) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userID := uint(rawUserID)
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
