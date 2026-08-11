package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type decodeJSONRequestBody struct {
	Name string `json:"name"`
}

func TestDecodeJSONRequest(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantOK     bool
		wantStatus int
		wantName   string
	}{
		{
			name:     "valid JSON",
			body:     `{"name":"Alex"}`,
			wantOK:   true,
			wantName: "Alex",
		},
		{
			name:       "empty body",
			body:       "",
			wantOK:     false,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown field",
			body:       `{"name":"Alex","unknown":true}`,
			wantOK:     false,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "multiple JSON values",
			body:       `{"name":"Alex"}{"name":"Maria"}`,
			wantOK:     false,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(
				http.MethodPost,
				"/",
				strings.NewReader(test.body),
			)
			response := httptest.NewRecorder()
			var destination decodeJSONRequestBody

			ok := DecodeJSONRequest(response, request, &destination)
			if ok != test.wantOK {
				t.Fatalf("DecodeJSONRequest() = %v, want %v", ok, test.wantOK)
			}
			if test.wantStatus != 0 && response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if test.wantOK && destination.Name != test.wantName {
				t.Fatalf("decoded name = %q, want %q", destination.Name, test.wantName)
			}
		})
	}
}

func TestDecodeJSONRequestRejectsLargeBody(t *testing.T) {
	body := `{"name":"` + strings.Repeat("a", int(maxJSONRequestSize)) + `"}`
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	response := httptest.NewRecorder()
	var destination decodeJSONRequestBody

	if DecodeJSONRequest(response, request, &destination) {
		t.Fatal("DecodeJSONRequest() = true, want false")
	}
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf(
			"status = %d, want %d",
			response.Code,
			http.StatusRequestEntityTooLarge,
		)
	}
}
