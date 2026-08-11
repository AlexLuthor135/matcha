package user

import (
	"backend/api"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type ResendVerificationEmailRequest struct {
	Email string `json:"email"`
}

type ResendVerificationEmailResponse struct {
	Message string `json:"message"`
}

func (h *UserHandler) ResendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	var request ResendVerificationEmailRequest
	if !api.DecodeJSONRequest(w, r, &request) {
		return
	}
	err := h.service.ResendVerificationEmail(r.Context(), request.Email)
	switch {
	case errors.Is(err, UserErrors.EmailBlank):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.EmailDeliveryFailed):
		http.Error(w, UserErrors.EmailDeliveryFailed.Error(), http.StatusBadGateway)
		return
	case err != nil:
		log.Printf("Rsend verification email: %v", err)
		http.Error(w, "Failed to resend verification email", http.StatusInternalServerError)
		return
	}
	response := ResendVerificationEmailResponse{
		Message: "If the account exists, a verification email has been sent",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode resend verification email: %v", err)
	}
}
