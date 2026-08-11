package user

import (
	"encoding/json"
	"log"
	"net/http"
)

type ListInterestTagsResponse struct {
	Interests []string `json:"interests"`
}

func (h *UserHandler) ListInterestTags(w http.ResponseWriter, r *http.Request) {
	response := ListInterestTagsResponse{
		Interests: h.service.ListInterestTags(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode interest tags: %v", err)
	}
}
