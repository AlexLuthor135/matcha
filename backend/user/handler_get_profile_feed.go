package user

import (
	"backend/middleware"
	"backend/models"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PublicProfileResponse struct {
	ID          uint                   `json:"id"`
	UserName    string                 `json:"user_name"`
	FirstName   string                 `json:"first_name"`
	LastName    string                 `json:"last_name"`
	Gender      string                 `json:"gender"`
	Preferences string                 `json:"preferences"`
	Bio         string                 `json:"bio"`
	Interests   []string               `json:"interests"`
	Avatar      string                 `json:"avatar"`
	Photos      []ProfilePhotoResponse `json:"photos"`
	FameRating  int64                  `json:"fame_rating"`
	Age         *int                   `json:"age"`
	Distance    *float64               `json:"distance"`
}

type ProfileFeedResponse struct {
	Profiles []PublicProfileResponse `json:"profiles"`
	Count    int                     `json:"count"`
}

func parseProfileFeedLimit(rawLimit string) (int, error) {
	if rawLimit == "" {
		return defaultProfileFeedLimit, nil
	}
	limit, err := strconv.Atoi(rawLimit)
	if err != nil {
		return 0, UserErrors.InvalidProfileFeedLimit
	}
	return limit, nil
}

func parseProfileFeedOptions(request *http.Request) (ProfileFeedOptions, error) {
	query := request.URL.Query()
	limit, err := parseProfileFeedLimit(query.Get("limit"))
	if err != nil {
		return ProfileFeedOptions{}, err
	}
	minAge, err := parseOptionalFeed[int](query.Get("min_age"))
	if err != nil {
		return ProfileFeedOptions{}, err
	}
	maxAge, err := parseOptionalFeed[int](query.Get("max_age"))
	if err != nil {
		return ProfileFeedOptions{}, err
	}
	maxDistance, err := parseOptionalFeed[float64](query.Get("max_distance"))
	if err != nil {
		return ProfileFeedOptions{}, err
	}
	minFame, err := parseOptionalFeed[int64](query.Get("min_fame"))
	if err != nil {
		return ProfileFeedOptions{}, err
	}
	maxFame, err := parseOptionalFeed[int64](query.Get("max_fame"))
	if err != nil {
		return ProfileFeedOptions{}, err
	}
	var interests []string
	rawInterests := strings.TrimSpace(query.Get("interests"))
	if rawInterests != "" {
		interests = strings.Split(rawInterests, ",")
	}
	return ProfileFeedOptions{
		Limit:       limit,
		MinAge:      minAge,
		MaxAge:      maxAge,
		MaxDistance: maxDistance,
		MinFame:     minFame,
		MaxFame:     maxFame,
		Interests:   interests,
		Sort:        ProfileFeedSort(query.Get("sort")),
	}, nil
}

func parseOptionalFeed[T int | int64 | float64](rawValue string) (*T, error) {
	rawValue = strings.TrimSpace(rawValue)
	if rawValue == "" {
		return nil, nil
	}
	var t T
	var value any
	var err error
	switch any(t).(type) {
	case int:
		value, err = strconv.Atoi(rawValue)
	case int64:
		value, err = strconv.ParseInt(rawValue, 10, 64)
	case float64:
		value, err = strconv.ParseFloat(rawValue, 64)
	}
	if err != nil {
		return nil, UserErrors.InvalidProfileFeedFilter
	}
	result := value.(T)
	return &result, nil
}

func publicProfileResponse(profile models.User) PublicProfileResponse {
	return PublicProfileResponse{
		ID:          profile.ID,
		UserName:    profile.UserName,
		FirstName:   profile.FirstName,
		LastName:    profile.LastName,
		Gender:      profile.Gender,
		Preferences: profile.Preferences,
		Bio:         profile.Bio,
		Interests:   profile.Interests,
		Avatar:      profile.Avatar,
		Photos:      profilePhotoResponses(profile.Photos),
		FameRating:  profile.FameRating,
		Age:         profileAge(profile.BirthDate),
		Distance:    profile.Distance,
	}
}

func publicProfileResponses(profiles []models.User) []PublicProfileResponse {
	response := make([]PublicProfileResponse, 0, len(profiles))
	for _, profile := range profiles {
		response = append(response, publicProfileResponse(profile))
	}
	return response
}

func profileAge(birthDate *time.Time) *int {
	if birthDate == nil {
		return nil
	}
	age := ageAt(*birthDate)
	return &age
}

func (h *UserHandler) GetProfileFeed(w http.ResponseWriter, r *http.Request) {
	h.serveProfileList(w, r, false)
}

func (h *UserHandler) SearchProfiles(w http.ResponseWriter, r *http.Request) {
	h.serveProfileList(w, r, true)
}

func (h *UserHandler) serveProfileList(w http.ResponseWriter, r *http.Request, search bool) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	options, err := parseProfileFeedOptions(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var profiles []models.User
	if search {
		profiles, err = h.service.SearchProfiles(r.Context(), userID, options)
	} else {
		profiles, err = h.service.GetProfileFeed(r.Context(), userID, options)
	}
	switch {
	case errors.Is(err, UserErrors.InvalidProfileFeedLimit) ||
		errors.Is(err, UserErrors.InvalidProfileFeedFilter) ||
		errors.Is(err, UserErrors.InvalidProfileFeedSort) ||
		errors.Is(err, UserErrors.InvalidLocation):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, UserErrors.UserNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
		return
	case err != nil:
		log.Printf("serve profile list for user %d: %v", userID, err)
		http.Error(w, "Failed to get profiles", http.StatusInternalServerError)
		return
	}
	response := ProfileFeedResponse{
		Profiles: publicProfileResponses(profiles),
		Count:    len(profiles),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Encode profile feed for %d: %v", userID, err)
	}
}
