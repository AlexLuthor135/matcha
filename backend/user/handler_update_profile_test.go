package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUpdateProfileHandlerValidatesBirthDateAndLocation(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		wantResponse string
	}{
		{
			name:         "invalid date format",
			body:         `{"birth_date":"10-05-2000"}`,
			wantResponse: "Birth date must use YYYY-MM-DD format\n",
		},
		{
			name:         "underage",
			body:         `{"birth_date":"2020-05-10"}`,
			wantResponse: UserErrors.UserUnderage.Error() + "\n",
		},
		{
			name:         "partial location",
			body:         `{"location":{"source":"gps","latitude":52.52,"consent":true}}`,
			wantResponse: UserErrors.InvalidLocation.Error() + "\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeUserRepository{
				updateProfileFn: func(context.Context, uint, UpdateProfileInput) error {
					t.Fatal("UpdateProfile() repository called for invalid request")
					return nil
				},
			}
			handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
			request := authenticatedUserRequest(
				httptest.NewRequest(http.MethodPatch, "/profile", strings.NewReader(test.body)),
				10,
			)
			response := httptest.NewRecorder()

			handler.UpdateProfile(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body = %q", response.Code, http.StatusBadRequest, response.Body.String())
			}
			if response.Body.String() != test.wantResponse {
				t.Fatalf("body = %q, want %q", response.Body.String(), test.wantResponse)
			}
		})
	}
}
