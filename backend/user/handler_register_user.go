package user

import (
	"backend/api"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type RegisterUserRequest struct {
	UserName  string `json:"user_name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type RegisteredUser struct {
	ID        uint   `json:"id"`
	UserName  string `json:"user_name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type RegisterUserResponse struct {
	Message string         `json:"message"`
	User    RegisteredUser `json:"user"`
}

func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterUserRequest
	if !api.DecodeJSONRequest(w, r, &req) {
		return
	}
	newUser, err := h.service.Register(r.Context(), RegisterInput{
		UserName:  req.UserName,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
	})
	switch {
	case errors.Is(err, UserErrors.RegistrationFieldMissing):
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.InvalidUserName):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.InvalidPassword) || errors.Is(err, UserErrors.InvalidEmail):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.UserAlreadyExists):
		http.Error(w, "Username or email already exists", http.StatusConflict)
		return
	case errors.Is(err, UserErrors.EmailDeliveryFailed):
		log.Printf("Send verification email during registration: %v", err)
		http.Error(w, UserErrors.EmailDeliveryFailed.Error(), http.StatusBadGateway)
		return
	case err != nil:
		log.Printf("Register user: %v", err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	response := RegisterUserResponse{
		Message: "User registered successfully",
		User: RegisteredUser{
			ID:        newUser.ID,
			UserName:  newUser.UserName,
			FirstName: newUser.FirstName,
			LastName:  newUser.LastName,
			Email:     newUser.Email,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode registration for user %d: %v", newUser.ID, err)
	}
}
