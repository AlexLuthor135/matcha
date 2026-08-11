package profile

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListInterestTagsHandler(t *testing.T) {
	service := NewService(&fakeUserRepository{}, &fakeImageStorage{})
	handler := NewHandler(service)
	request := httptest.NewRequest(http.MethodGet, "/interests", nil)
	response := httptest.NewRecorder()

	handler.ListInterestTags(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}

	var body ListInterestTagsResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := service.ListInterestTags()
	if len(body.Interests) != len(want) {
		t.Fatalf("interest count = %d, want %d", len(body.Interests), len(want))
	}
	for index := range want {
		if body.Interests[index] != want[index] {
			t.Fatalf("interest[%d] = %q, want %q", index, body.Interests[index], want[index])
		}
	}
}

func TestProfileHandlersRejectUnsupportedInterestTag(t *testing.T) {
	tests := []struct {
		name    string
		request *http.Request
		handle  func(*Handler, http.ResponseWriter, *http.Request)
	}{
		{
			name: "complete profile",
			request: authenticatedUserRequest(
				httptest.NewRequest(
					http.MethodPost,
					"/profile/complete",
					strings.NewReader(`{
						"gender":"Male",
						"preferences":"Female",
						"bio":"Hello",
						"interests":["not-supported"],
						"birth_date":"2000-05-10",
						"location":{
							"source":"manual",
							"name":"Berlin",
							"latitude":52.52,
							"longitude":13.405
						}
					}`),
				),
				17,
			),
			handle: func(handler *Handler, response http.ResponseWriter, request *http.Request) {
				handler.CompleteProfile(response, request)
			},
		},
		{
			name: "update profile",
			request: authenticatedUserRequest(
				httptest.NewRequest(
					http.MethodPatch,
					"/profile",
					strings.NewReader(`{"interests":["not-supported"]}`),
				),
				17,
			),
			handle: func(handler *Handler, response http.ResponseWriter, request *http.Request) {
				handler.UpdateProfile(response, request)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewHandler(NewService(&fakeUserRepository{}, &fakeImageStorage{}))
			response := httptest.NewRecorder()

			test.handle(handler, response, test.request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body = %q", response.Code, http.StatusBadRequest, response.Body.String())
			}
			wantBody := ProfileErrors.InvalidInterestTag.Error() + "\n"
			if response.Body.String() != wantBody {
				t.Fatalf("body = %q, want %q", response.Body.String(), wantBody)
			}
		})
	}
}
