package user

import (
	"backend/api"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type RequestPasswordResetRequest struct {
	Email string `json:"email"`
}

type RequestPasswordResetResponse struct {
	Message string `json:"message"`
}

func (h *UserHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var request RequestPasswordResetRequest
	if !api.DecodeJSONRequest(w, r, &request) {
		return
	}
	err := h.service.RequestPasswordReset(r.Context(), request.Email)
	switch {
	case errors.Is(err, UserErrors.EmailBlank):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.PasswordResetEmailDeliveryFailed):
		log.Printf("Send password reset email: %v", err)
		http.Error(w, UserErrors.PasswordResetEmailDeliveryFailed.Error(), http.StatusBadGateway)
		return
	case err != nil:
		log.Printf("Request password reset: %v", err)
		http.Error(w, "Failed to request password reset", http.StatusInternalServerError)
		return
	}
	response := RequestPasswordResetResponse{
		Message: "If the account exists, a password reset email has been sent",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode password reset request: %v", err)
	}
}
