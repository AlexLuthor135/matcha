package user

import (
	"backend/models"
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestServiceRequestPasswordResetCreatesAndEmailsToken(t *testing.T) {
	const userID uint = 21
	const email = "user@example.com"
	var savedToken models.AccountToken

	repository := &fakeUserRepository{
		getUserByEmailFn: func(_ context.Context, gotEmail string) (models.User, error) {
			if gotEmail != email {
				t.Fatalf("GetUserByEmail() email = %q, want %q", gotEmail, email)
			}
			return models.User{ID: userID, IsVerified: true}, nil
		},
		replaceAccountTokenFn: func(_ context.Context, token models.AccountToken) error {
			savedToken = token
			return nil
		},
	}
	emailSender := &fakeEmailSender{
		sendPasswordResetEmailFn: func(_ context.Context, recipientEmail string, rawToken string) error {
			if recipientEmail != email {
				t.Fatalf("SendPasswordResetEmail() recipient = %q, want %q", recipientEmail, email)
			}
			if rawToken == "" {
				t.Fatal("SendPasswordResetEmail() raw token is blank")
			}
			if savedToken.Hash != hashAccountToken(rawToken) {
				t.Fatal("saved token hash does not match emailed raw token")
			}
			return nil
		},
	}
	service := NewService(repository, &fakeImageStorage{})
	service.SetEmailSender(emailSender)

	if err := service.RequestPasswordReset(context.Background(), "  USER@Example.COM "); err != nil {
		t.Fatalf("RequestPasswordReset() unexpected error: %v", err)
	}
	if savedToken.UserID != userID {
		t.Fatalf("password reset token user ID = %d, want %d", savedToken.UserID, userID)
	}
	if savedToken.Purpose != models.AccountTokenPurposePasswordReset {
		t.Fatalf("password reset token purpose = %q, want password reset", savedToken.Purpose)
	}
	minimumExpiry := time.Now().UTC().Add(29 * time.Minute)
	maximumExpiry := time.Now().UTC().Add(31 * time.Minute)
	if savedToken.ExpiresAt.Before(minimumExpiry) || savedToken.ExpiresAt.After(maximumExpiry) {
		t.Fatalf("password reset token expiration = %v, want approximately 30 minutes", savedToken.ExpiresAt)
	}
}

func TestServiceRequestPasswordResetDoesNotRevealAccountState(t *testing.T) {
	tests := []struct {
		name       string
		lookupUser func(context.Context, string) (models.User, error)
	}{
		{
			name: "unknown account",
			lookupUser: func(context.Context, string) (models.User, error) {
				return models.User{}, UserErrors.UserNotFound
			},
		},
		{
			name: "unverified account",
			lookupUser: func(context.Context, string) (models.User, error) {
				return models.User{ID: 21, IsVerified: false}, nil
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeUserRepository{getUserByEmailFn: test.lookupUser}
			service := NewService(repository, &fakeImageStorage{})
			service.SetEmailSender(&fakeEmailSender{})

			if err := service.RequestPasswordReset(context.Background(), "user@example.com"); err != nil {
				t.Fatalf("RequestPasswordReset() unexpected error: %v", err)
			}
		})
	}
}

func TestServiceRequestPasswordResetWrapsDeliveryError(t *testing.T) {
	repository := &fakeUserRepository{
		getUserByEmailFn: func(context.Context, string) (models.User, error) {
			return models.User{ID: 21, IsVerified: true}, nil
		},
		replaceAccountTokenFn: func(context.Context, models.AccountToken) error {
			return nil
		},
	}
	emailSender := &fakeEmailSender{
		sendPasswordResetEmailFn: func(context.Context, string, string) error {
			return errors.New("SMTP unavailable")
		},
	}
	service := NewService(repository, &fakeImageStorage{})
	service.SetEmailSender(emailSender)

	err := service.RequestPasswordReset(context.Background(), "user@example.com")
	if !errors.Is(err, UserErrors.PasswordResetEmailDeliveryFailed) {
		t.Fatalf("RequestPasswordReset() error = %v, want %v", err, UserErrors.PasswordResetEmailDeliveryFailed)
	}
}

func TestServiceResetPasswordHashesTokenAndPassword(t *testing.T) {
	const rawToken = "password-reset-token"
	const newPassword = "NewValid1!Password"
	repository := &fakeUserRepository{
		resetPasswordWithTokenFn: func(_ context.Context, tokenHash string, passwordHash string) error {
			if tokenHash != hashAccountToken(rawToken) {
				t.Fatalf("ResetPasswordWithToken() token hash = %q, want hash of raw token", tokenHash)
			}
			if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(newPassword)); err != nil {
				t.Fatalf("ResetPasswordWithToken() password hash does not match new password: %v", err)
			}
			return nil
		},
	}

	err := NewService(repository, &fakeImageStorage{}).ResetPassword(
		context.Background(),
		ResetPasswordInput{Token: rawToken, NewPassword: newPassword},
	)
	if err != nil {
		t.Fatalf("ResetPassword() unexpected error: %v", err)
	}
}

func TestServiceResetPasswordRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input ResetPasswordInput
		want  error
	}{
		{
			name:  "blank token",
			input: ResetPasswordInput{NewPassword: "NewValid1!Password"},
			want:  UserErrors.PasswordResetFieldsMissing,
		},
		{
			name:  "blank password",
			input: ResetPasswordInput{Token: "token"},
			want:  UserErrors.PasswordResetFieldsMissing,
		},
		{
			name:  "weak password",
			input: ResetPasswordInput{Token: "token", NewPassword: "password"},
			want:  UserErrors.InvalidPassword,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewService(&fakeUserRepository{}, &fakeImageStorage{}).ResetPassword(
				context.Background(),
				test.input,
			)
			if !errors.Is(err, test.want) {
				t.Fatalf("ResetPassword() error = %v, want %v", err, test.want)
			}
		})
	}
}
