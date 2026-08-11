package relationship

import (
	"backend/middleware"
	"errors"
	"log"
	"net/http"
	"strconv"
)

func parseTargetUserID(r *http.Request) (uint, error) {
	rawID := r.PathValue("targetUserID")
	parsedID, err := strconv.ParseUint(rawID, 10, strconv.IntSize)
	if err != nil || parsedID == 0 {
		return 0, RelationshipErrors.InvalidTargetUserID
	}
	return uint(parsedID), nil
}

func writeUserBlockError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, RelationshipErrors.InvalidTargetUserID), errors.Is(err, RelationshipErrors.CannotBlockSelf):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, RelationshipErrors.UserNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
	case errors.Is(err, RelationshipErrors.TargetUserNotFound):
		http.Error(w, "Target user not found", http.StatusNotFound)
	default:
		http.Error(w, "Failed to change user block", http.StatusInternalServerError)
	}
}

func (h *Handler) BlockUser(w http.ResponseWriter, r *http.Request) {
	blockerID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	blockedUserID, err := parseTargetUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.service.BlockUser(r.Context(), blockerID, blockedUserID)
	if err != nil {
		log.Printf("Block user %d by user %d: %v", blockedUserID, blockerID, err)
		writeUserBlockError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	blockerID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	blockedUserID, err := parseTargetUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.service.UnblockUser(r.Context(), blockerID, blockedUserID)
	if err != nil {
		log.Printf("Unblock user %d by user %d: %v", blockedUserID, blockerID, err)
		writeUserBlockError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
