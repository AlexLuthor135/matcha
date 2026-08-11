package user

import (
	"backend/api"
	"errors"
	"log"
	"net/http"
)

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if !api.DecodeJSONRequest(w, r, &req) {
		return
	}
	user, err := h.service.Login(r.Context(), req.Login, req.Password)
	switch {
	case errors.Is(err, UserErrors.LoginFieldsMissing):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.InvalidCredentials):
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	case errors.Is(err, UserErrors.EmailNotVerified):
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	case err != nil:
		log.Printf("Login user: %v", err)
		http.Error(w, "Failed to log in", http.StatusInternalServerError)
		return
	}
	session, err := h.service.CreateAuthSession(r.Context(), user.ID)
	if err != nil {
		log.Printf("Create authentication session for user %d: %v", user.ID, err)
		http.Error(w, "Failed to create authentication session", http.StatusInternalServerError)
		return
	}
	if api.GenerateTokens(w, user, session.ID, session.ExpiresAt) {
		if err := h.service.RevokeAuthSession(r.Context(), session.ID, user.ID); err != nil {
			log.Printf("Revoke authentication failed session %q: %v", session.ID, err)
		}
		return
	}
}
