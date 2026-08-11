package user

import (
	"backend/api"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}

func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var request ResetPasswordRequest
	if !api.DecodeJSONRequest(w, r, &request) {
		return
	}
	err := h.service.ResetPassword(r.Context(), ResetPasswordInput{
		Token:       request.Token,
		NewPassword: request.NewPassword,
	})
	switch {
	case errors.Is(err, UserErrors.PasswordResetFieldsMissing),
		errors.Is(err, UserErrors.InvalidPassword),
		errors.Is(err, UserErrors.InvalidPasswordResetToken):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err != nil:
		log.Printf("Reset password: %v", err)
		http.Error(w, "Failed to reset password", http.StatusInternalServerError)
		return
	}
	api.ClearAuthCookies(w)
	response := ResetPasswordResponse{
		Message: "Password reset successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode reset password: %v", err)
	}
}
