package main

import (
	"encoding/json"
	"net/http"
    "strings"
)

type UpdateBioRequest struct {
    Gender    string   `json:"gender"`
    Preferences    string   `json:"preferences"`
    Bio       string   `json:"bio"`
    Interests []string `json:"interests"`
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