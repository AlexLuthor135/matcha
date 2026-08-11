package account

import (
	"backend/models"
	"context"
	"errors"
	"testing"
	"time"
)

func TestServiceVerifyEmailHashesRawToken(t *testing.T) {
	const rawToken = "verification-token"
	repository := &fakeUserRepository{
		verifyEmailFn: func(_ context.Context, tokenHash string) error {
			if tokenHash != hashAccountToken(rawToken) {
				t.Fatalf("VerifyEmail() token hash = %q, want hash of raw token", tokenHash)
			}
			return nil
		},
	}

	err := NewService(repository, nil).VerifyEmail(
		context.Background(),
		rawToken,
	)
	if err != nil {
		t.Fatalf("VerifyEmail() unexpected error: %v", err)
	}
}

func TestServiceVerifyEmailRejectsBlankToken(t *testing.T) {
	err := NewService(&fakeUserRepository{}, nil).VerifyEmail(
		context.Background(),
		"   ",
	)
	if !errors.Is(err, AccountErrors.InvalidVerificationToken) {
		t.Fatalf("VerifyEmail() error = %v, want %v", err, AccountErrors.InvalidVerificationToken)
	}
}

func TestServiceResendVerificationEmailSendsReplacementToken(t *testing.T) {
	const userID uint = 17
	const normalizedEmail = "user@example.com"
	var savedToken models.AccountToken

	repository := &fakeUserRepository{
		getUserByEmailFn: func(_ context.Context, email string) (models.User, error) {
			if email != normalizedEmail {
				t.Fatalf("GetUserByEmail() email = %q, want %q", email, normalizedEmail)
			}
			return models.User{ID: userID}, nil
		},
		replaceAccountTokenFn: func(_ context.Context, token models.AccountToken) error {
			savedToken = token
			return nil
		},
	}
	emailSender := &fakeEmailSender{
		sendVerificationEmailFn: func(_ context.Context, recipientEmail string, rawToken string) error {
			if recipientEmail != normalizedEmail {
				t.Fatalf("SendVerificationEmail() recipient = %q, want %q", recipientEmail, normalizedEmail)
			}
			if rawToken == "" {
				t.Fatal("SendVerificationEmail() raw token is blank")
			}
			if savedToken.Hash != hashAccountToken(rawToken) {
				t.Fatalf("saved token hash does not match the emailed raw token")
			}
			return nil
		},
	}
	service := NewService(repository, nil)
	service.SetEmailSender(emailSender)

	err := service.ResendVerificationEmail(
		context.Background(),
		"  USER@Example.COM  ",
	)
	if err != nil {
		t.Fatalf("ResendVerificationEmail() unexpected error: %v", err)
	}
	if savedToken.UserID != userID {
		t.Fatalf("replacement token user ID = %d, want %d", savedToken.UserID, userID)
	}
	if savedToken.Purpose != models.AccountTokenPurposeEmailVerification {
		t.Fatalf("replacement token purpose = %q, want email verification", savedToken.Purpose)
	}
	minimumExpiry := time.Now().UTC().Add(23*time.Hour + 59*time.Minute)
	maximumExpiry := time.Now().UTC().Add(24*time.Hour + time.Minute)
	if savedToken.ExpiresAt.Before(minimumExpiry) || savedToken.ExpiresAt.After(maximumExpiry) {
		t.Fatalf("replacement token expiration = %v, want approximately 24 hours", savedToken.ExpiresAt)
	}
}

func TestServiceResendVerificationEmailDoesNotRevealAccountState(t *testing.T) {
	tests := []struct {
		name       string
		lookupUser func(context.Context, string) (models.User, error)
	}{
		{
			name: "unknown account",
			lookupUser: func(context.Context, string) (models.User, error) {
				return models.User{}, AccountErrors.UserNotFound
			},
		},
		{
			name: "already verified account",
			lookupUser: func(context.Context, string) (models.User, error) {
				return models.User{ID: 17, IsVerified: true}, nil
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeUserRepository{getUserByEmailFn: test.lookupUser}
			service := NewService(repository, nil)
			service.SetEmailSender(&fakeEmailSender{})

			if err := service.ResendVerificationEmail(context.Background(), "user@example.com"); err != nil {
				t.Fatalf("ResendVerificationEmail() unexpected error: %v", err)
			}
		})
	}
}

func TestServiceResendVerificationEmailWrapsDeliveryError(t *testing.T) {
	deliveryError := errors.New("SMTP unavailable")
	repository := &fakeUserRepository{
		getUserByEmailFn: func(context.Context, string) (models.User, error) {
			return models.User{ID: 17, Email: "user@example.com"}, nil
		},
		replaceAccountTokenFn: func(context.Context, models.AccountToken) error {
			return nil
		},
	}
	emailSender := &fakeEmailSender{
		sendVerificationEmailFn: func(context.Context, string, string) error {
			return deliveryError
		},
	}
	service := NewService(repository, nil)
	service.SetEmailSender(emailSender)

	err := service.ResendVerificationEmail(context.Background(), "user@example.com")
	if !errors.Is(err, AccountErrors.EmailDeliveryFailed) {
		t.Fatalf("ResendVerificationEmail() error = %v, want %v", err, AccountErrors.EmailDeliveryFailed)
	}
}
