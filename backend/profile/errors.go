package profile

import "errors"

var ProfileErrors = struct {
	UserNotFound              error
	ProfileFieldsMissing      error
	InvalidGenderPreference   error
	NoProfileFields           error
	ProfileBioBlank           error
	ProfileInterestsMissing   error
	AvatarEmpty               error
	AvatarTooLarge            error
	AvatarTypeUnsupported     error
	PhotoLimitExceeded        error
	PhotosMissing             error
	PhotoEmpty                error
	PhotoTooLarge             error
	PhotoTypeUnsupported      error
	PhotoNotFound             error
	InvalidPhotoID            error
	ProfilePictureRequired    error
	UserUnderage              error
	InvalidLocation           error
	InvalidLocationSource     error
	LocationConsentRequired   error
	ManualLocationNameMissing error
	InvalidInterestTag        error
}{
	UserNotFound:              errors.New("user not found"),
	ProfileFieldsMissing:      errors.New("all profile fields are required"),
	InvalidGenderPreference:   errors.New("invalid gender or preference"),
	NoProfileFields:           errors.New("no profile fields provided"),
	ProfileBioBlank:           errors.New("bio cannot be blank"),
	ProfileInterestsMissing:   errors.New("at least one interest is required"),
	AvatarEmpty:               errors.New("avatar cannot be empty"),
	AvatarTooLarge:            errors.New("avatar is too large"),
	AvatarTypeUnsupported:     errors.New("avatar type is unsupported"),
	PhotoLimitExceeded:        errors.New("profile photo limit exceeded"),
	PhotosMissing:             errors.New("at least one photo is required"),
	PhotoEmpty:                errors.New("photo cannot be empty"),
	PhotoTooLarge:             errors.New("photo is too large"),
	PhotoTypeUnsupported:      errors.New("photo type is unsupported"),
	PhotoNotFound:             errors.New("photo not found"),
	InvalidPhotoID:            errors.New("invalid photo ID"),
	ProfilePictureRequired:    errors.New("profile picture is required"),
	UserUnderage:              errors.New("user must be at least 18 years old"),
	InvalidLocation:           errors.New("user location is invalid"),
	InvalidLocationSource:     errors.New("location source must be gps or manual"),
	LocationConsentRequired:   errors.New("explicit consent is required for GPS location"),
	ManualLocationNameMissing: errors.New("city or neighborhood is required for manual location"),
	InvalidInterestTag:        errors.New("one or more interest tags are not supported"),
}
