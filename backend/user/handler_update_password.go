package user

import (
	"backend/middleware"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type UpdatePasswordResponse struct {
	Message string `json:"message"`
}

func (h *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	err := h.service.UpdatePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword)
	switch {
	case errors.Is(err, UserErrors.PasswordFieldsMissing):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.InvalidPassword):
		http.Error(w, "New password must contain at least 8 characters, an uppercase letter, a lowercase letter, a number and a special character", http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.NewPasswordUnchanged):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.CurrentPasswordIncorrect):
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	case errors.Is(err, UserErrors.UserNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
		return
	case err != nil:
		log.Printf("Updated password for user %d: %v", userID, err)
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}
	response := UpdatePasswordResponse{
		Message: "Password updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode password update for %d: %v", userID, err)
	}
}
