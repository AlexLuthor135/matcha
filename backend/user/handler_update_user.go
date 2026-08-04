package user

import (
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
	Message string `json:"message"`
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	err := h.service.UpdateUser(
		r.Context(),
		userID,
		UserUpdateInput{
			UserName:  req.UserName,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
		})
	switch {
	case errors.Is(err, UserErrors.NoUserFields):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.UserNameBlank), errors.Is(err, UserErrors.FirstNameBlank), errors.Is(err, UserErrors.LastNameBlank), errors.Is(err, UserErrors.EmailBlank):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.UserNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
		return
	case errors.Is(err, UserErrors.UserAlreadyExists):
		http.Error(w, "Username or email already exists", http.StatusConflict)
		return
	case err != nil:
		log.Printf("Update user %d: %v", userID, err)
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := UpdateUserResponse{
		Message: "User updated successfully",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode update for user %d: %v", userID, err)
	}
}
