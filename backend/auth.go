package main

import (
	"fmt"
	"os"
	"context"
	"net/http"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string
const userIDKey contextKey = contextKey("userID")

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		secret := os.Getenv("JWT_SECRET")
		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["type"] != "access" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userID := uint(claims["user_id"].(float64))
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

