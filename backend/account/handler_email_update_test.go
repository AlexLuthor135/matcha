package account

import (
	"backend/models"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUpdateUserHandlerReturnsPendingEmail(t *testing.T) {
	const userID uint = 17
	const pendingEmail = "new@example.com"
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
			if input.Email == nil || *input.Email != pendingEmail {
				t.Fatalf("UpdateUser() email = %v, want %q", input.Email, pendingEmail)
			}
			if verificationToken == nil {
				t.Fatal("UpdateUser() verification token is nil")
			}
			return UpdateUserResult{
				EmailChanged: true,
				PendingEmail: pendingEmail,
			}, nil
		},
	}
	emailSender := &fakeEmailSender{
		sendVerificationEmailFn: func(context.Context, string, string) error {
			return nil
		},
	}
	service := NewService(repository, nil)
	service.SetEmailSender(emailSender)
	handler := NewHandler(service)
	request := authenticatedUserRequest(
		httptest.NewRequest(
			http.MethodPatch,
			"/users/me",
			strings.NewReader(`{"email":"new@example.com"}`),
		),
		userID,
	)
	response := httptest.NewRecorder()

	handler.UpdateUser(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", response.Code, http.StatusOK, response.Body.String())
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}
	var body UpdateUserResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.VerificationRequired {
		t.Fatal("verification_required = false, want true")
	}
	if body.PendingEmail != pendingEmail {
		t.Fatalf("pending_email = %q, want %q", body.PendingEmail, pendingEmail)
	}
	if body.Message != "User updated; verify the new email address" {
		t.Fatalf("message = %q, want email verification message", body.Message)
	}
}

func TestVerifyEmailHandlerMapsAddressConflict(t *testing.T) {
	repository := &fakeUserRepository{
		verifyEmailFn: func(context.Context, string) error {
			return AccountErrors.UserAlreadyExists
		},
	}
	handler := NewHandler(NewService(repository, nil))
	request := httptest.NewRequest(
		http.MethodGet,
		"/auth/verify-email?token=verification-token",
		nil,
	)
	response := httptest.NewRecorder()

	handler.VerifyEmail(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusConflict)
	}
	wantBody := "Email address is already used by another account\n"
	if response.Body.String() != wantBody {
		t.Fatalf("body = %q, want %q", response.Body.String(), wantBody)
	}
}
