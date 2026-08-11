package discovery

import (
	"backend/models"
	"context"
	"math"
	"sort"
	"strings"
)

const (
	defaultProfileFeedLimit = 20
	maxProfileFeedLimit     = 50
)

type ProfileFeedSort string

const (
	ProfileFeedSortRecommended ProfileFeedSort = "recommended"
	ProfileFeedSortDistance    ProfileFeedSort = "distance"
	ProfileFeedSortAge         ProfileFeedSort = "age"
	ProfileFeedSortFame        ProfileFeedSort = "fame"
	ProfileFeedSortInterests   ProfileFeedSort = "interests"
)

type ProfileFeedOptions struct {
	Limit       int
	MinAge      *int
	MaxAge      *int
	MaxDistance *float64
	MinFame     *int64
	MaxFame     *int64
	Interests   []string
	Sort        ProfileFeedSort
}

type rankedProfile struct {
	Profile         models.User
	Age             int
	CommonInterests int
}

func (options *ProfileFeedOptions) Normalize() {
	options.Sort = ProfileFeedSort(strings.ToLower(strings.TrimSpace(string(options.Sort))))
	if options.Sort == "" {
		options.Sort = ProfileFeedSortRecommended
	}
	normalizedInterests := make([]string, 0, len(options.Interests))
	seenInterests := make(map[string]struct{})
	for _, interest := range options.Interests {
		normalized := strings.ToLower(strings.TrimSpace(interest))
		if normalized == "" {
			continue
		}
		if _, exists := seenInterests[normalized]; exists {
			continue
		}
		seenInterests[normalized] = struct{}{}
		normalizedInterests = append(normalizedInterests, normalized)
	}
	options.Interests = normalizedInterests
}

func (options *ProfileFeedOptions) Validate() error {
	if options.Limit < 1 {
		return DiscoveryErrors.InvalidProfileFeedLimit
	}
	if options.MinAge != nil {
		if *options.MinAge < 18 || *options.MinAge > 150 {
			return DiscoveryErrors.InvalidProfileFeedFilter
		}
	}
	if options.MaxAge != nil {
		if *options.MaxAge < 18 || *options.MaxAge > 150 {
			return DiscoveryErrors.InvalidProfileFeedFilter
		}
	}
	if options.MinAge != nil && options.MaxAge != nil && *options.MinAge > *options.MaxAge {
		return DiscoveryErrors.InvalidProfileFeedFilter
	}
	if options.MaxDistance != nil {
		if *options.MaxDistance < 0 || math.IsNaN(*options.MaxDistance) || math.IsInf(*options.MaxDistance, 0) {
			return DiscoveryErrors.InvalidProfileFeedFilter
		}
	}
	if options.MinFame != nil && *options.MinFame < 0 {
		return DiscoveryErrors.InvalidProfileFeedFilter
	}
	if options.MaxFame != nil && *options.MaxFame < 0 {
		return DiscoveryErrors.InvalidProfileFeedFilter
	}
	if options.MinFame != nil && options.MaxFame != nil && *options.MinFame > *options.MaxFame {
		return DiscoveryErrors.InvalidProfileFeedFilter
	}
	switch options.Sort {
	case ProfileFeedSortRecommended,
		ProfileFeedSortDistance,
		ProfileFeedSortAge,
		ProfileFeedSortFame,
		ProfileFeedSortInterests:
		return nil
	default:
		return DiscoveryErrors.InvalidProfileFeedSort
	}
}

func commonInterestCount(firstInterests []string, secondInterests []string) int {
	firstSet := make(map[string]struct{}, len(firstInterests))
	for _, interest := range firstInterests {
		normalized := strings.ToLower(strings.TrimSpace(interest))
		if normalized != "" {
			firstSet[normalized] = struct{}{}
		}
	}
	matched := make(map[string]struct{})
	count := 0

	for _, interest := range secondInterests {
		normalized := strings.ToLower(strings.TrimSpace(interest))
		if _, exists := firstSet[normalized]; !exists {
			continue
		}
		if _, alreadyCounted := matched[normalized]; alreadyCounted {
			continue
		}
		matched[normalized] = struct{}{}
		count++
	}
	return count
}

func containsAllInterests(candidateInterests []string, requiredInterests []string) bool {
	if len(requiredInterests) == 0 {
		return true
	}
	candidateSet := make(map[string]struct{}, len(candidateInterests))
	for _, interest := range candidateInterests {
		normalized := strings.ToLower(strings.TrimSpace(interest))
		if normalized != "" {
			candidateSet[normalized] = struct{}{}
		}
	}
	for _, required := range requiredInterests {
		if _, exists := candidateSet[required]; !exists {
			return false
		}
	}
	return true
}

func rateProfiles(first rankedProfile, second rankedProfile) bool {
	firstDistance := *first.Profile.Distance
	secondDistance := *second.Profile.Distance
	if firstDistance != secondDistance {
		return firstDistance < secondDistance
	}
	if first.CommonInterests != second.CommonInterests {
		return first.CommonInterests > second.CommonInterests
	}
	if first.Profile.FameRating != second.Profile.FameRating {
		return first.Profile.FameRating > second.Profile.FameRating
	}
	return first.Profile.ID < second.Profile.ID
}

func rankProfiles(first rankedProfile, second rankedProfile, sortBy ProfileFeedSort) bool {
	switch sortBy {
	case ProfileFeedSortAge:
		if first.Age != second.Age {
			return first.Age < second.Age
		}
	case ProfileFeedSortFame:
		if first.Profile.FameRating != second.Profile.FameRating {
			return first.Profile.FameRating > second.Profile.FameRating
		}
	case ProfileFeedSortInterests:
		if first.CommonInterests != second.CommonInterests {
			return first.CommonInterests > second.CommonInterests
		}
	case ProfileFeedSortDistance, ProfileFeedSortRecommended:
	}
	return rateProfiles(first, second)
}

func (options ProfileFeedOptions) sortCandidates(age int, distance float64, candidate models.User) bool {
	if options.MinAge != nil && age < *options.MinAge {
		return false
	}
	if options.MaxAge != nil && age > *options.MaxAge {
		return false
	}
	if options.MaxDistance != nil && distance > *options.MaxDistance {
		return false
	}
	if options.MinFame != nil && candidate.FameRating < *options.MinFame {
		return false
	}
	if options.MaxFame != nil && candidate.FameRating > *options.MaxFame {
		return false
	}
	if !containsAllInterests(candidate.Interests, options.Interests) {
		return false
	}
	return true
}

func (s *Service) GetProfileFeed(ctx context.Context, userID uint, options ProfileFeedOptions) ([]models.User, error) {
	return s.getProfiles(ctx, userID, options, true)
}

func (s *Service) SearchProfiles(ctx context.Context, userID uint, options ProfileFeedOptions) ([]models.User, error) {
	return s.getProfiles(ctx, userID, options, false)
}

func (s *Service) getProfiles(ctx context.Context, userID uint, options ProfileFeedOptions, excludeDecided bool) ([]models.User, error) {
	options.Normalize()
	if err := options.Validate(); err != nil {
		return nil, err
	}
	if options.Limit > maxProfileFeedLimit {
		options.Limit = maxProfileFeedLimit
	}
	currentProfile, err := s.repository.GetDiscoveryProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	if currentProfile.Latitude == nil || currentProfile.Longitude == nil || !isLocationValid(currentProfile.Latitude, currentProfile.Longitude) {
		return nil, DiscoveryErrors.InvalidLocation
	}
	preferredGender := strings.TrimSpace(currentProfile.Preferences)
	ownGender := strings.TrimSpace(currentProfile.Gender)
	candidates, err := s.repository.ListProfileCandidates(ctx, userID, preferredGender, ownGender, excludeDecided)
	if err != nil {
		return nil, err
	}
	rankedProfiles := make([]rankedProfile, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.BirthDate == nil || candidate.Latitude == nil || candidate.Longitude == nil || !isLocationValid(candidate.Latitude, candidate.Longitude) {
			continue
		}
		age := ageAt(*candidate.BirthDate)
		distance := distanceInKM(*currentProfile.Latitude, *currentProfile.Longitude, *candidate.Latitude, *candidate.Longitude)
		if !options.sortCandidates(age, distance, candidate) {
			continue
		}
		candidate.Distance = &distance
		rankedProfiles = append(rankedProfiles, rankedProfile{
			Profile:         candidate,
			Age:             age,
			CommonInterests: commonInterestCount(currentProfile.Interests, candidate.Interests),
		})
	}
	sort.SliceStable(rankedProfiles, func(firstIndex int, secondIndex int) bool {
		return rankProfiles(rankedProfiles[firstIndex], rankedProfiles[secondIndex], options.Sort)
	})
	if options.Limit > len(rankedProfiles) {
		options.Limit = len(rankedProfiles)
	}
	profiles := make([]models.User, 0, options.Limit)
	for _, ranked := range rankedProfiles[:options.Limit] {
		profiles = append(profiles, ranked.Profile)
	}
	return profiles, nil
}
