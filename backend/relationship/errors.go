package relationship

import "errors"

var RelationshipErrors = struct {
	UserNotFound           error
	TargetUserNotFound     error
	InvalidTargetUserID    error
	CannotDecideOwnProfile error
	InvalidProfileDecision error
	CannotViewOwnProfile   error
	ProfilePictureRequired error
	CannotBlockSelf        error
	CannotReportSelf       error
	InvalidLocation        error
}{
	UserNotFound:           errors.New("user not found"),
	TargetUserNotFound:     errors.New("target user not found"),
	InvalidTargetUserID:    errors.New("invalid target user id"),
	CannotDecideOwnProfile: errors.New("cannot like or dislike your own profile"),
	InvalidProfileDecision: errors.New("decision must be like or dislike"),
	CannotViewOwnProfile:   errors.New("cannot view your own profile"),
	ProfilePictureRequired: errors.New("profile picture is required"),
	CannotBlockSelf:        errors.New("cannot block yourself"),
	CannotReportSelf:       errors.New("cannot report yourself"),
	InvalidLocation:        errors.New("user location is invalid"),
}
