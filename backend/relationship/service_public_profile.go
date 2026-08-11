package relationship

import (
	"backend/models"
	"context"
	"errors"
)

type PublicProfileResult struct {
	Profile      models.User
	Relationship models.ProfileRelationship
}

func (s *Service) GetPublicProfile(ctx context.Context, viewerID uint, targetUserID uint) (PublicProfileResult, error) {
	if targetUserID == 0 {
		return PublicProfileResult{}, RelationshipErrors.InvalidTargetUserID
	}
	if viewerID != targetUserID {
		blockExists, err := s.repository.HasBlockBetweenUsers(ctx, viewerID, targetUserID)
		if err != nil {
			return PublicProfileResult{}, err
		}
		if blockExists {
			return PublicProfileResult{}, RelationshipErrors.TargetUserNotFound
		}
	}
	profile, err := s.repository.GetPublicProfile(ctx, targetUserID)
	if errors.Is(err, RelationshipErrors.UserNotFound) {
		return PublicProfileResult{}, RelationshipErrors.TargetUserNotFound
	}
	if err != nil {
		return PublicProfileResult{}, err
	}
	if !profile.IsCompleted {
		return PublicProfileResult{}, RelationshipErrors.TargetUserNotFound
	}
	if viewerID == targetUserID {
		distance := 0.0
		profile.Distance = &distance
		return PublicProfileResult{Profile: profile}, nil
	}
	viewerLatitude, viewerLongitude, err := s.repository.GetUserLocation(ctx, viewerID)
	if err != nil {
		return PublicProfileResult{}, err
	}
	if profile.Latitude == nil || profile.Longitude == nil || !isLocationValid(profile.Latitude, profile.Longitude) {
		return PublicProfileResult{}, RelationshipErrors.InvalidLocation
	}
	distance := distanceInKM(*viewerLatitude, *viewerLongitude, *profile.Latitude, *profile.Longitude)
	profile.Distance = &distance
	relationship, err := s.repository.GetProfileRelationship(ctx, viewerID, targetUserID)
	if err != nil {
		return PublicProfileResult{}, err
	}
	return PublicProfileResult{Profile: profile, Relationship: relationship}, nil
}
