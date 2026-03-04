package main

import (
	"encoding/json"
	"net/http"
)

type UpdateBioRequest struct {
    Gender    string   `json:"gender"`
    Bio       string   `json:"bio"`
    Interests []string `json:"interests"`
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