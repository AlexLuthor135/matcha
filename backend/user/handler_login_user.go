package user

import (
	"backend/api"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	user, err := h.service.Login(r.Context(), req.Email, req.Password)
	switch {
	case errors.Is(err, UserErrors.LoginFieldsMissing):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.InvalidCredentials):
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if err != nil {
		log.Printf("Login user: %v", err)
		http.Error(w, "Failed to log in", http.StatusInternalServerError)
		return
	}
	if api.GenerateTokens(w, user) {
		return
	}
}
