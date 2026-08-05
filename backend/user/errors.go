package user

import "errors"

var UserErrors = struct {
	UserNotFound             error
	LoginFieldsMissing       error
	InvalidCredentials       error
	RegistrationFieldMissing error
	InvalidPassword          error
	UserAlreadyExists        error
	PasswordFieldsMissing    error
	CurrentPasswordIncorrect error
	NewPasswordUnchanged     error
	InvalidProfileFeedLimit  error
	ProfileFieldsMissing     error
	InvalidGenderPreference  error
	NoProfileFields          error
	ProfileBioBlank          error
	ProfileInterestsMissing  error
	NoUserFields             error
	UserNameBlank            error
	FirstNameBlank           error
	LastNameBlank            error
	EmailBlank               error
	AvatarEmpty              error
	AvatarTooLarge           error
	AvatarTypeUnsupported    error
	PhotoLimitExceeded       error
	PhotosMissing            error
	PhotoEmpty               error
	PhotoTooLarge            error
	PhotoTypeUnsupported     error
	PhotoNotFound            error
	InvalidPhotoID           error
	InvalidTargetUserID      error
	CannotDecideOwnProfile   error
	InvalidProfileDecision   error
	TargetUserNotFound       error
}{
	UserNotFound:             errors.New("user not found"),
	LoginFieldsMissing:       errors.New("email and password are required"),
	InvalidCredentials:       errors.New("invalid email or password"),
	RegistrationFieldMissing: errors.New("all registration fields are required"),
	InvalidPassword:          errors.New("invalid password"),
	UserAlreadyExists:        errors.New("username or email already exists"),
	PasswordFieldsMissing:    errors.New("current and new password are required"),
	CurrentPasswordIncorrect: errors.New("current password is incorrect"),
	NewPasswordUnchanged:     errors.New("new password must be different"),
	InvalidProfileFeedLimit:  errors.New("profile feed limit must be a positive integer"),
	ProfileFieldsMissing:     errors.New("all profile fields are required"),
	InvalidGenderPreference:  errors.New("invalid gender or preference"),
	NoProfileFields:          errors.New("no profile fields provided"),
	ProfileBioBlank:          errors.New("bio cannot be blank"),
	ProfileInterestsMissing:  errors.New("at least one interest is required"),
	NoUserFields:             errors.New("no user fields provided"),
	UserNameBlank:            errors.New("username cannot be blank"),
	FirstNameBlank:           errors.New("first name cannot be blank"),
	LastNameBlank:            errors.New("last name cannot be blank"),
	EmailBlank:               errors.New("email cannot be blank"),
	AvatarEmpty:              errors.New("avatar cannot be empty"),
	AvatarTooLarge:           errors.New("avatar is too large"),
	AvatarTypeUnsupported:    errors.New("avatar type is unsupported"),
	PhotoLimitExceeded:       errors.New("profile photo limit exceeded"),
	PhotosMissing:            errors.New("at least one photo is required"),
	PhotoEmpty:               errors.New("photo cannot be empty"),
	PhotoTooLarge:            errors.New("photo is too large"),
	PhotoTypeUnsupported:     errors.New("photo type is unsupported"),
	PhotoNotFound:            errors.New("photo not found"),
	InvalidPhotoID:           errors.New("invalid photo ID"),
	InvalidTargetUserID:      errors.New("invalid target user id"),
	CannotDecideOwnProfile:   errors.New("cannot like or dislike your own profile"),
	InvalidProfileDecision:   errors.New("decision must be like or dislike"),
	TargetUserNotFound:       errors.New("target user not found"),
}
