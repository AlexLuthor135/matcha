package account

import (
	"backend/api"
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

func (h *Handler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req UpdatePasswordRequest
	if !api.DecodeJSONRequest(w, r, &req) {
		return
	}
	err := h.service.UpdatePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword)
	switch {
	case errors.Is(err, AccountErrors.PasswordFieldsMissing):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, AccountErrors.InvalidPassword):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, AccountErrors.NewPasswordUnchanged):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, AccountErrors.CurrentPasswordIncorrect):
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	case errors.Is(err, AccountErrors.UserNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
		return
	case err != nil:
		log.Printf("Updated password for user %d: %v", userID, err)
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}
	api.ClearAuthCookies(w)
	response := UpdatePasswordResponse{
		Message: "Password updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode password update for %d: %v", userID, err)
	}
}
