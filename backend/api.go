package main

import (
    "errors"
	"os"
	"fmt"
	"time"
	"github.com/golang-jwt/jwt/v5"
    "github.com/jackc/pgx/v5/pgconn"
	"encoding/json"
	"net/http"
    "strings"
	"golang.org/x/crypto/bcrypt"
)

type LoginResponse struct {
    Message     string `json:"message"`
    IsCompleted bool `gorm:"default:false"`
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
    if strings.TrimSpace(req.UserName) == "" || strings.TrimSpace(req.FirstName) == "" || strings.TrimSpace(req.LastName) == "" || strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
        http.Error(w, "All fields are required", http.StatusBadRequest)
        return
    }
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	newUser := User{
		UserName: req.UserName,
		FirstName: req.FirstName,
		LastName: req.LastName,
		Email: req.Email,
		Password: string(hashedPassword),
	}

	result := DB.Create(&newUser)
	if result.Error != nil {
        var pgErr *pgconn.PgError
        if errors.As(result.Error, &pgErr) && pgErr.Code == "23505" {
            http.Error(w, "Username or email already exists", http.StatusConflict)
            return
        }
		http.Error(w, "Error creating user: "+result.Error.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"user": map[string]interface{}{
			"id": newUser.ID,
			"username": newUser.UserName,
			"first_name": newUser.FirstName,
			"last_name": newUser.LastName,
			"email": newUser.Email,
		},
	})
}

func loginUser(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    var user User
    if result := DB.Where("email = ?", req.Email).First(&user); result.Error != nil {
        http.Error(w, "Invalid email or password", http.StatusUnauthorized)
        return
    }
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        http.Error(w, "Invalid email or password", http.StatusUnauthorized)
        return
    }

    
    secret := os.Getenv("JWT_SECRET")
    
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": user.ID,
        "type": "access",
        "exp":  time.Now().Add(time.Minute * 15).Unix(),
    })
    accessTokenString, err := accessToken.SignedString([]byte(secret))
    if err != nil {
        http.Error(w, "Error generating access token", http.StatusInternalServerError)
        return
    }

    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": user.ID,
        "type": "refresh",
        "exp":  time.Now().Add(time.Hour * 24 * 7).Unix(),
    })
    refreshTokenString, err := refreshToken.SignedString([]byte(secret))
    if err != nil {
        http.Error(w, "Error generating refresh token", http.StatusInternalServerError)
        return
    }
    
    http.SetCookie(w, &http.Cookie{
        Name: "access_token",
        Value: accessTokenString,
        HttpOnly: true,
        Secure: true, 
        SameSite: http.SameSiteLaxMode,
        Path: "/",
        Expires: time.Now().Add(time.Minute * 15),
    })
    
    http.SetCookie(w, &http.Cookie{
        Name: "refresh_token",
        Value: refreshTokenString,
        HttpOnly: true,
        Secure: true, 
        SameSite: http.SameSiteLaxMode, 
        Path: "/api/refresh",
        Expires: time.Now().Add(time.Hour * 24 * 7),
    })
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    
    response := LoginResponse{
        Message:    "Login successful",
        IsCompleted: user.IsCompleted,
    }
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }
}

func verifyUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user User
	if result := DB.First(&user, userID); result.Error != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"is_staff": user.IsStaff,
        "is_completed": user.IsCompleted,
    })
}

func logoutUser(w http.ResponseWriter, r *http.Request) {
    http.SetCookie(w, &http.Cookie{
        Name:     "access_token",
        Value:    "",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
        Path:     "/",
        MaxAge:   -1,
    })

    http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    "",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
        Path:     "/api/accounts/token/refresh/",
        MaxAge:   -1,
    })
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Logged out successfully",
    })
}

func refreshToken(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("refresh_token")
    if err != nil {
        http.Error(w, "Refresh token missing", http.StatusUnauthorized)
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
        http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
        return
    }
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || claims["type"] != "refresh" {
        http.Error(w, "Invalid token type", http.StatusUnauthorized)
        return
    }
    userID, ok := claims["user_id"].(float64)
    if !ok {
        http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
        return
    }
    newAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": uint(userID),
        "type":    "access",
        "exp":     time.Now().Add(time.Minute * 15).Unix(),
    })
    
    newAccessTokenString, err := newAccessToken.SignedString([]byte(secret))
    if err != nil {
        http.Error(w, "Failed to generate new access token", http.StatusInternalServerError)
        return
    }
    http.SetCookie(w, &http.Cookie{
        Name:     "access_token",
        Value:    newAccessTokenString,
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteLaxMode,
        Path:     "/",
        Expires:  time.Now().Add(time.Minute * 15),
    })
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Token refreshed successfully",
    })
}