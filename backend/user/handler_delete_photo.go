package user

import (
	"backend/middleware"
	"errors"
	"log"
	"net/http"
	"strconv"
)

func (h *UserHandler) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	rawPhotoID := r.PathValue("photoID")
	parsedPhotoID, err := strconv.ParseInt(rawPhotoID, 10, strconv.IntSize)
	if err != nil {
		http.Error(w, "Invalid photo ID", http.StatusBadRequest)
		return
	}
	photoID := uint(parsedPhotoID)
	err = h.service.DeletePhoto(r.Context(), userID, photoID)
	switch {
	case errors.Is(err, UserErrors.InvalidPhotoID):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.UserNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
		return
	case errors.Is(err, UserErrors.PhotoNotFound):
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	case err != nil:
		log.Printf("Delete photo %d for user %d: %v", photoID, userID, err)
		http.Error(w, "Failed to delete photo", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
