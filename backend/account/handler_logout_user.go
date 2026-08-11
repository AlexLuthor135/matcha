package account

import (
	"backend/api"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

func (h *Handler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	refreshCookie, cookieErr := r.Cookie("refresh_token")
	if cookieErr == nil {
		claims, err := api.ParseRefreshToken(refreshCookie.Value)
		if errors.Is(err, api.ErrAuthenticationNotConfigured) {
			http.Error(w, "Authentication is not configured", http.StatusInternalServerError)
			return
		}
		if err == nil {
			if err := h.service.RevokeAuthSession(r.Context(), claims.SessionID, claims.UserID); err != nil {
				log.Printf("Revoke session for user %d: %v", claims.UserID, err)
				http.Error(w, "Failed to log out", http.StatusInternalServerError)
				return
			}
		}
	}
	api.ClearAuthCookies(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}
