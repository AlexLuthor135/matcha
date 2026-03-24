package main

import (
	"encoding/json"
	"net/http"
    "strings"
    "errors"
    "gorm.io/gorm"
    "golang.org/x/crypto/bcrypt"
)

type UpdateBioRequest struct {
    Gender    string   `json:"gender"`
    Preferences    string   `json:"preferences"`
    Bio       string   `json:"bio"`
    Interests []string `json:"interests"`
}

type BioResponse struct {
	Gender      string   `json:"gender"`
	Preferences string   `json:"preferences"`
	Bio         string   `json:"bio"`
	Interests   []string `json:"interests"`
}

type UpdateUserRequest struct {
    Username string `gorm:"unique;not null"`
	FirstName string `gorm:"not null"`
	LastName string `gorm:"not null"`
	Email string `gorm:"unique;not null"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func createBio(w http.ResponseWriter, r *http.Request) {
    userID, ok := r.Context().Value(userIDKey).(uint)
    if !ok {
        http.Error(w, "Unauthorized: Invalid context", http.StatusUnauthorized)
        return
    }
    var req UpdateBioRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }
    isCompleted := strings.TrimSpace(req.Gender) != "" &&
        strings.TrimSpace(req.Preferences) != "" &&
        strings.TrimSpace(req.Bio) != "" &&
        len(req.Interests) > 0

    result := DB.Model(&User{}).Where("id = ?", userID).Updates(User{
        Gender:    req.Gender,
        Bio:       req.Bio,
        Interests: req.Interests,
        Preferences:    req.Preferences,
        IsCompleted: isCompleted,
    })

    if result.Error != nil {
        http.Error(w, "Failed to update profile", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Profile updated successfully",
    })
}

func updateBio(w http.ResponseWriter, r *http.Request) {
    userID, ok := r.Context().Value(userIDKey).(uint)
    if !ok {
        http.Error(w, "Unauthorized: Invalid context", http.StatusUnauthorized)
        return
    }
    var req UpdateBioRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }
    result := DB.Model(&User{}).Where("id = ?", userID).Updates(User{
        Gender:    req.Gender,
        Bio:       req.Bio,
        Interests: req.Interests,
        Preferences:    req.Preferences,
    })

    if result.Error != nil {
        http.Error(w, "Failed to update profile", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Profile updated successfully",
    })
}

func getBio(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized: Invalid context", http.StatusUnauthorized)
		return
	}

	var user User
	result := DB.Select("gender", "preferences", "bio", "interests").First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve profile", http.StatusInternalServerError)
		return
	}
	response := BioResponse{
		Gender:      user.Gender,
		Preferences: user.Preferences,
		Bio:         user.Bio,
		Interests:   user.Interests,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func updateUser(w http.ResponseWriter, r *http.Request) {
    userID, ok := r.Context().Value(userIDKey).(uint)
    if !ok {
        http.Error(w, "Unauthorized: Invalid context", http.StatusUnauthorized)
        return
    }
    var req UpdateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }
    result := DB.Model(&User{}).Where("id = ?", userID).Updates(User{
        Username: req.Username,
	    FirstName: req.FirstName,
	    LastName: req.LastName,
	    Email: req.Email,
    })

    if result.Error != nil {
        http.Error(w, "Failed to update profile", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Profile updated successfully",
    })
}

func updatePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized: Invalid context", http.StatusUnauthorized)
		return
	}
	var req UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if req.OldPassword == "" || req.NewPassword == "" {
		http.Error(w, "Both old and new passwords are required", http.StatusBadRequest)
		return
	}
	var user User
	if err := DB.Select("password").First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword))
	if err != nil {
		http.Error(w, "Incorrect current password", http.StatusUnauthorized)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error processing new password", http.StatusInternalServerError)
		return
	}
	result := DB.Model(&User{}).Where("id = ?", userID).Update("password", string(hashedPassword))
	if result.Error != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password updated successfully",
	})
}