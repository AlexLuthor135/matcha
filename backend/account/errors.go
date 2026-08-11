package account

import "errors"

var AccountErrors = struct {
	UserNotFound                     error
	LoginFieldsMissing               error
	InvalidCredentials               error
	RegistrationFieldMissing         error
	InvalidPassword                  error
	UserAlreadyExists                error
	PasswordFieldsMissing            error
	CurrentPasswordIncorrect         error
	NewPasswordUnchanged             error
	NoUserFields                     error
	UserNameBlank                    error
	FirstNameBlank                   error
	LastNameBlank                    error
	EmailBlank                       error
	EmailNotVerified                 error
	InvalidVerificationToken         error
	EmailDeliveryFailed              error
	InvalidAccountTokenPurpose       error
	SMTPConfigInvalid                error
	SMTPAddressInvalid               error
	PasswordResetFieldsMissing       error
	InvalidPasswordResetToken        error
	PasswordResetEmailDeliveryFailed error
	InvalidUserName                  error
	InvalidEmail                     error
	InvalidAuthSession               error
}{
	UserNotFound:                     errors.New("user not found"),
	LoginFieldsMissing:               errors.New("username or email and password are required"),
	InvalidCredentials:               errors.New("invalid email or password"),
	RegistrationFieldMissing:         errors.New("all registration fields are required"),
	InvalidPassword:                  errors.New("password does not meet security requirements or is commonly used"),
	UserAlreadyExists:                errors.New("username or email already exists"),
	PasswordFieldsMissing:            errors.New("current and new password are required"),
	CurrentPasswordIncorrect:         errors.New("current password is incorrect"),
	NewPasswordUnchanged:             errors.New("new password must be different"),
	NoUserFields:                     errors.New("no user fields provided"),
	UserNameBlank:                    errors.New("username cannot be blank"),
	FirstNameBlank:                   errors.New("first name cannot be blank"),
	LastNameBlank:                    errors.New("last name cannot be blank"),
	EmailBlank:                       errors.New("email cannot be blank"),
	EmailNotVerified:                 errors.New("email is not verified"),
	InvalidVerificationToken:         errors.New("verification link is invalid or expired"),
	EmailDeliveryFailed:              errors.New("failed to send verification email"),
	InvalidAccountTokenPurpose:       errors.New("invalid account token purpose"),
	SMTPConfigInvalid:                errors.New("SMTP configuration is incomplete"),
	SMTPAddressInvalid:               errors.New("invalid SMTP sender address"),
	PasswordResetFieldsMissing:       errors.New("reset token and new password are required"),
	InvalidPasswordResetToken:        errors.New("password reset link is invalid or expired"),
	PasswordResetEmailDeliveryFailed: errors.New("failed to send password reset email"),
	InvalidUserName:                  errors.New("username must contain 3 to 30 letters, numbers, dots, hyphens or underscores"),
	InvalidEmail:                     errors.New("invalid email address"),
	InvalidAuthSession:               errors.New("authentication session is invalid or expired"),
}
