package user

import (
	"backend/models"
	"context"
	"errors"
	"testing"
	"time"
)

func TestServiceUpdateUserCreatesAndEmailsVerificationToken(t *testing.T) {
	const userID uint = 17
	const normalizedEmail = "new@example.com"
	var savedToken *models.AccountToken

	repository := &fakeUserRepository{
		updateUserFn: func(
			_ context.Context,
			gotUserID uint,
			input UserUpdateInput,
			verificationToken *models.AccountToken,
		) (UpdateUserResult, error) {
			if gotUserID != userID {
				t.Fatalf("UpdateUser() user ID = %d, want %d", gotUserID, userID)
			}
			if input.Email == nil || *input.Email != normalizedEmail {
				t.Fatalf("UpdateUser() email = %v, want %q", input.Email, normalizedEmail)
			}
			if verificationToken == nil {
				t.Fatal("UpdateUser() verification token is nil")
			}
			tokenCopy := *verificationToken
			savedToken = &tokenCopy
			return UpdateUserResult{
				EmailChanged: true,
				PendingEmail: normalizedEmail,
			}, nil
		},
	}
	emailSender := &fakeEmailSender{
		sendVerificationEmailFn: func(
			_ context.Context,
			recipientEmail string,
			rawToken string,
		) error {
			if recipientEmail != normalizedEmail {
				t.Fatalf("SendVerificationEmail() recipient = %q, want %q", recipientEmail, normalizedEmail)
			}
			if rawToken == "" {
				t.Fatal("SendVerificationEmail() raw token is blank")
			}
			if savedToken == nil {
				t.Fatal("SendVerificationEmail() called before repository UpdateUser()")
			}
			if savedToken.Hash != hashAccountToken(rawToken) {
				t.Fatal("saved token hash does not match emailed raw token")
			}
			return nil
		},
	}
	service := NewService(repository, &fakeImageStorage{})
	service.SetEmailSender(emailSender)

	result, err := service.UpdateUser(
		context.Background(),
		userID,
		UserUpdateInput{Email: stringPointer("  NEW@Example.COM  ")},
	)
	if err != nil {
		t.Fatalf("UpdateUser() unexpected error: %v", err)
	}
	if !result.EmailChanged || result.PendingEmail != normalizedEmail {
		t.Fatalf("UpdateUser() result = %+v, want changed pending email %q", result, normalizedEmail)
	}
	if savedToken.UserID != userID {
		t.Fatalf("verification token user ID = %d, want %d", savedToken.UserID, userID)
	}
	if savedToken.Purpose != models.AccountTokenPurposeEmailVerification {
		t.Fatalf("verification token purpose = %q, want email verification", savedToken.Purpose)
	}
	minimumExpiry := time.Now().UTC().Add(23*time.Hour + 59*time.Minute)
	maximumExpiry := time.Now().UTC().Add(24*time.Hour + time.Minute)
	if savedToken.ExpiresAt.Before(minimumExpiry) || savedToken.ExpiresAt.After(maximumExpiry) {
		t.Fatalf("verification token expiration = %v, want approximately 24 hours", savedToken.ExpiresAt)
	}
}

func TestServiceUpdateUserDoesNotEmailWhenAddressIsUnchanged(t *testing.T) {
	repository := &fakeUserRepository{
		updateUserFn: func(
			context.Context,
			uint,
			UserUpdateInput,
			*models.AccountToken,
		) (UpdateUserResult, error) {
			return UpdateUserResult{EmailChanged: false}, nil
		},
	}
	emailSender := &fakeEmailSender{
		sendVerificationEmailFn: func(context.Context, string, string) error {
			t.Fatal("SendVerificationEmail() called for unchanged email")
			return nil
		},
	}
	service := NewService(repository, &fakeImageStorage{})
	service.SetEmailSender(emailSender)

	result, err := service.UpdateUser(
		context.Background(),
		17,
		UserUpdateInput{Email: stringPointer("user@example.com")},
	)
	if err != nil {
		t.Fatalf("UpdateUser() unexpected error: %v", err)
	}
	if result.EmailChanged || result.PendingEmail != "" {
		t.Fatalf("UpdateUser() result = %+v, want unchanged email", result)
	}
}

func TestServiceUpdateUserWithoutEmailDoesNotRequireEmailSender(t *testing.T) {
	repository := &fakeUserRepository{
		updateUserFn: func(
			_ context.Context,
			_ uint,
			input UserUpdateInput,
			verificationToken *models.AccountToken,
		) (UpdateUserResult, error) {
			if input.FirstName == nil || *input.FirstName != "Alex" {
				t.Fatalf("UpdateUser() first name = %v, want normalized Alex", input.FirstName)
			}
			if verificationToken != nil {
				t.Fatal("UpdateUser() received verification token without email")
			}
			return UpdateUserResult{}, nil
		},
	}

	_, err := NewService(repository, &fakeImageStorage{}).UpdateUser(
		context.Background(),
		17,
		UserUpdateInput{FirstName: stringPointer("  Alex  ")},
	)
	if err != nil {
		t.Fatalf("UpdateUser() unexpected error: %v", err)
	}
}

func TestServiceUpdateUserRejectsInvalidEmail(t *testing.T) {
	_, err := NewService(&fakeUserRepository{}, &fakeImageStorage{}).UpdateUser(
		context.Background(),
		17,
		UserUpdateInput{Email: stringPointer("not-an-email")},
	)
	if !errors.Is(err, UserErrors.InvalidEmail) {
		t.Fatalf("UpdateUser() error = %v, want %v", err, UserErrors.InvalidEmail)
	}
}

func TestServiceUpdateUserWrapsEmailDeliveryError(t *testing.T) {
	repository := &fakeUserRepository{
		updateUserFn: func(
			context.Context,
			uint,
			UserUpdateInput,
			*models.AccountToken,
		) (UpdateUserResult, error) {
			return UpdateUserResult{
				EmailChanged: true,
				PendingEmail: "new@example.com",
			}, nil
		},
	}
	emailSender := &fakeEmailSender{
		sendVerificationEmailFn: func(context.Context, string, string) error {
			return errors.New("SMTP unavailable")
		},
	}
	service := NewService(repository, &fakeImageStorage{})
	service.SetEmailSender(emailSender)

	_, err := service.UpdateUser(
		context.Background(),
		17,
		UserUpdateInput{Email: stringPointer("new@example.com")},
	)
	if !errors.Is(err, UserErrors.EmailDeliveryFailed) {
		t.Fatalf("UpdateUser() error = %v, want %v", err, UserErrors.EmailDeliveryFailed)
	}
}

func stringPointer(value string) *string {
	return &value
}
