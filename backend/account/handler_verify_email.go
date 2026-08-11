package account

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type VerifyEmailResponse struct {
	Message string `json:"message"`
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	rawToken := r.URL.Query().Get("token")
	err := h.service.VerifyEmail(r.Context(), rawToken)
	switch {
	case errors.Is(err, AccountErrors.InvalidVerificationToken):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, AccountErrors.UserAlreadyExists):
		http.Error(w, "Email address is already used by another account", http.StatusConflict)
		return
	case err != nil:
		log.Printf("Verify email: %v", err)
		http.Error(w, "Failed to verify email", http.StatusInternalServerError)
		return
	}
	response := VerifyEmailResponse{
		Message: "Email verified successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode email verification: %v", err)
	}
}
