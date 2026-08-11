package profile

import (
	"backend/middleware"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

const maxAvatarBodySize = 6 << 20

type UploadAvatarResponse struct {
	AvatarURL string `json:"avatar_url"`
}

func (h *Handler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarBodySize)
	file, _, err := r.FormFile("avatar")
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			http.Error(w, "Avatar is too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Avatar file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()
	fileData, err := io.ReadAll(io.LimitReader(file, maxAvatarFileSize+1))
	if err != nil {
		log.Printf("Read avatar for user %d: %v", userID, err)
		http.Error(w, "Failed to read avatar", http.StatusInternalServerError)
		return
	}
	avatarURL, err := h.service.UploadAvatar(r.Context(), userID, fileData)
	switch {
	case errors.Is(err, ProfileErrors.AvatarEmpty):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, ProfileErrors.AvatarTooLarge):
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		return
	case errors.Is(err, ProfileErrors.AvatarTypeUnsupported):
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	case errors.Is(err, ProfileErrors.UserNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
		return
	case err != nil:
		log.Printf("Upload avatar for user %d: %v", userID, err)
		http.Error(w, "Failed to upload avatar", http.StatusInternalServerError)
		return
	}
	response := UploadAvatarResponse{
		AvatarURL: avatarURL,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode updated avatar for user %d: %v", userID, err)
	}
}
