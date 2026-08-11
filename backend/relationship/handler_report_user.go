package relationship

import (
	"backend/middleware"
	"errors"
	"log"
	"net/http"
)

func (h *Handler) ReportUser(w http.ResponseWriter, r *http.Request) {
	reporterID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	reportedUserID, err := parseTargetUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.service.ReportUser(r.Context(), reporterID, reportedUserID)
	switch {
	case errors.Is(err, RelationshipErrors.InvalidTargetUserID), errors.Is(err, RelationshipErrors.CannotReportSelf):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, RelationshipErrors.UserNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
		return
	case errors.Is(err, RelationshipErrors.TargetUserNotFound):
		http.Error(w, "Target user not found", http.StatusNotFound)
		return
	case err != nil:
		log.Printf("Report user %d by user %d: %v", reportedUserID, reporterID, err)
		http.Error(w, "Failed to report user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
