package user

import (
	"backend/middleware"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

const maxPhotoRequestSize = 26 << 20

type UploadPhotoResponse struct {
	Photos []ProfilePhotoResponse `json:"photos"`
}

func (h *UserHandler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxPhotoRequestSize)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		var maxBytesError *http.MaxBytesError

		if errors.As(err, &maxBytesError) {
			http.Error(w, "Photo upload is too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Invalid photo upload", http.StatusBadRequest)
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}
	fileHeaders := r.MultipartForm.File["photos"]
	filesData := make([][]byte, 0, len(fileHeaders))
	for _, fileHeader := range fileHeaders {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Failed to open photo", http.StatusBadRequest)
			return
		}
		fileData, readErr := io.ReadAll(io.LimitReader(file, maxPhotoFileSize+1))
		closeErr := file.Close()
		if readErr != nil {
			log.Printf("Read photo for user %d: %v", userID, readErr)
			http.Error(w, "Failed to read photo", http.StatusInternalServerError)
			return
		}
		if closeErr != nil {
			log.Printf("Close uploaded photo for user %d: %v", userID, closeErr)
		}
		filesData = append(filesData, fileData)
	}
	photos, err := h.service.UploadPhotos(r.Context(), userID, filesData)
	switch {
	case errors.Is(err, UserErrors.PhotosMissing):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.PhotoEmpty):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.PhotoTooLarge):
		http.Error(w, "Each photo must not be larger than 5 MB", http.StatusRequestEntityTooLarge)
		return
	case errors.Is(err, UserErrors.PhotoTypeUnsupported):
		http.Error(w, "Photos must be JPEG, PNG or WEBP", http.StatusUnsupportedMediaType)
		return
	case errors.Is(err, UserErrors.PhotoLimitExceeded):
		http.Error(w, "You can have no more than 5 pictures including your profile picture", http.StatusConflict)
		return
	case errors.Is(err, UserErrors.UserNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
		return
	case err != nil:
		log.Printf("Upload photos for user %d: %v", userID, err)
		http.Error(w, "Failed to upload photos", http.StatusInternalServerError)
		return
	}

	response := UploadPhotoResponse{
		Photos: profilePhotoResponses(photos),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode uploaded photos for user %d: %v", userID, err)
	}
}
