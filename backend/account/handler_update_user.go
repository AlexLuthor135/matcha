package account

import (
	"backend/api"
	"backend/middleware"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type UpdateUserRequest struct {
	UserName  *string `json:"user_name"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     *string `json:"email"`
}

type UpdateUserResponse struct {
	Message              string `json:"message"`
	VerificationRequired bool   `json:"verification_required"`
	PendingEmail         string `json:"pending_email,omitempty"`
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req UpdateUserRequest
	if !api.DecodeJSONRequest(w, r, &req) {
		return
	}
	result, err := h.service.UpdateUser(
		r.Context(),
		userID,
		UserUpdateInput{
			UserName:  req.UserName,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
		})
	switch {
	case errors.Is(err, AccountErrors.NoUserFields):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, AccountErrors.UserNameBlank),
		errors.Is(err, AccountErrors.InvalidUserName),
		errors.Is(err, AccountErrors.FirstNameBlank),
		errors.Is(err, AccountErrors.LastNameBlank),
		errors.Is(err, AccountErrors.EmailBlank),
		errors.Is(err, AccountErrors.InvalidEmail):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, AccountErrors.UserNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
		return
	case errors.Is(err, AccountErrors.UserAlreadyExists):
		http.Error(w, "Username or email already exists", http.StatusConflict)
		return
	case errors.Is(err, AccountErrors.EmailDeliveryFailed):
		log.Printf("Send updated-email verification for user %d: %v", userID, err)
		http.Error(w, AccountErrors.EmailDeliveryFailed.Error(), http.StatusBadGateway)
		return
	case err != nil:
		log.Printf("Update user %d: %v", userID, err)
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := UpdateUserResponse{
		Message:              "User updated successfully",
		VerificationRequired: result.EmailChanged,
		PendingEmail:         result.PendingEmail,
	}
	if result.EmailChanged {
		response.Message = "User updated; verify the new email address"
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode update for user %d: %v", userID, err)
	}
}
