package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCompleteProfileHandlerRequiresProfilePicture(t *testing.T) {
	repository := &fakeUserRepository{
		getAvatarURLFn: func(context.Context, uint) (string, error) {
			return "", nil
		},
		completeProfileFn: func(context.Context, uint, CompleteProfileInput) (bool, error) {
			t.Fatal("CompleteProfile() repository called without a profile picture")
			return false, nil
		},
	}
	handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
	request := authenticatedUserRequest(
		httptest.NewRequest(
			http.MethodPost,
			"/profile/complete",
			strings.NewReader(`{
				"gender":"Male",
				"preferences":"Female",
				"bio":"Hello",
				"interests":["music"],
				"birth_date":"2000-05-10",
				"location":{
					"source":"manual",
					"name":"Berlin",
					"latitude":52.52,
					"longitude":13.405,
					"consent":false
				}
			}`),
		),
		10,
	)
	response := httptest.NewRecorder()

	handler.CompleteProfile(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %q", response.Code, http.StatusBadRequest, response.Body.String())
	}
	wantBody := UserErrors.ProfilePictureRequired.Error() + "\n"
	if response.Body.String() != wantBody {
		t.Fatalf("body = %q, want %q", response.Body.String(), wantBody)
	}
}
