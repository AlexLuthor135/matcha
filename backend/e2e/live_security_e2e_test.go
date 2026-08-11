package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func securityLiveRawRequest(
	t *testing.T,
	client *http.Client,
	method string,
	requestURL string,
	contentType string,
	body []byte,
) *http.Response {
	t.Helper()
	request, err := http.NewRequest(method, requestURL, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create security live request: %v", err)
	}
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("send security live request: %v", err)
	}
	return response
}

func securityLiveCookieValue(
	t *testing.T,
	client *http.Client,
	baseURL string,
	name string,
) string {
	t.Helper()
	parsedURL, err := http.NewRequest(http.MethodGet, baseURL, nil)
	if err != nil {
		t.Fatalf("parse cookie URL: %v", err)
	}
	for _, cookie := range client.Jar.Cookies(parsedURL.URL) {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	t.Fatalf("cookie %q is missing", name)
	return ""
}

func TestLiveSecurityAndAuthorization(t *testing.T) {
	baseURL := strings.TrimRight(os.Getenv("SECURITY_E2E_BASE_URL"), "/")
	seedPassword := os.Getenv("SECURITY_E2E_SEED_PASSWORD")
	if baseURL == "" || seedPassword == "" {
		t.Skip("set SECURITY_E2E_BASE_URL and SECURITY_E2E_SEED_PASSWORD to run live security test")
	}

	response := profileLiveJSONRequest(t, profileLiveClient(t), http.MethodGet, baseURL+"/api/accounts/profile", nil)
	profileLiveRequireStatus(t, response, http.StatusUnauthorized)
	response = profileLiveJSONRequest(t, profileLiveClient(t), http.MethodPost, baseURL+"/api/accounts/profiles/feed", nil)
	profileLiveRequireStatus(t, response, http.StatusUnauthorized)

	_, injectionStatus, injectionBody := accountLiveLoginStatus(t, baseURL, "' OR 1=1 --", "Anything9!Password")
	_, wrongStatus, wrongBody := accountLiveLoginStatus(t, baseURL, "seed_0201", "Wrong9!Password")
	if injectionStatus != http.StatusUnauthorized || wrongStatus != http.StatusUnauthorized {
		t.Fatalf("injection/wrong login statuses = %d/%d", injectionStatus, wrongStatus)
	}
	if string(injectionBody) != string(wrongBody) {
		t.Fatalf("credential errors differ: injection %q, wrong password %q", injectionBody, wrongBody)
	}

	firstClient, firstStatus, _ := accountLiveLoginStatus(t, baseURL, "seed_0201", seedPassword)
	secondClient, secondStatus, _ := accountLiveLoginStatus(t, baseURL, "seed_0202", seedPassword)
	if firstStatus != http.StatusOK || secondStatus != http.StatusOK {
		t.Fatalf("seed login statuses = %d/%d", firstStatus, secondStatus)
	}

	forgedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 1,
		"type":    "access",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	forgedTokenString, err := forgedToken.SignedString([]byte("wrong-secret"))
	if err != nil {
		t.Fatalf("sign forged token: %v", err)
	}
	forgedClient := profileLiveClient(t)
	forgedRequest, err := http.NewRequest(http.MethodGet, baseURL+"/api/accounts/profile", nil)
	if err != nil {
		t.Fatalf("create forged-token request: %v", err)
	}
	forgedRequest.AddCookie(&http.Cookie{Name: "access_token", Value: forgedTokenString})
	response, err = forgedClient.Do(forgedRequest)
	if err != nil {
		t.Fatalf("send forged-token request: %v", err)
	}
	profileLiveRequireStatus(t, response, http.StatusUnauthorized)

	accessToken := securityLiveCookieValue(t, firstClient, baseURL, "access_token")
	if got := accountLiveRefreshWithRawToken(t, baseURL, accessToken); got != http.StatusUnauthorized {
		t.Fatalf("access token used as refresh status = %d, want %d", got, http.StatusUnauthorized)
	}

	response = securityLiveRawRequest(
		t,
		firstClient,
		http.MethodPatch,
		baseURL+"/api/accounts/profile",
		"application/json",
		[]byte(`{"unexpected":true}`),
	)
	profileLiveRequireStatus(t, response, http.StatusBadRequest)
	response = securityLiveRawRequest(
		t,
		firstClient,
		http.MethodPatch,
		baseURL+"/api/accounts/profile",
		"application/json",
		[]byte(`{"bio":"first"}{"bio":"second"}`),
	)
	profileLiveRequireStatus(t, response, http.StatusBadRequest)
	oversizedJSON := []byte(`{"bio":"` + strings.Repeat("a", 70<<10) + `"}`)
	response = securityLiveRawRequest(
		t,
		firstClient,
		http.MethodPatch,
		baseURL+"/api/accounts/profile",
		"application/json",
		oversizedJSON,
	)
	profileLiveRequireStatus(t, response, http.StatusRequestEntityTooLarge)

	response = profileLiveJSONRequest(
		t,
		firstClient,
		http.MethodGet,
		baseURL+"/api/accounts/profiles/1%20OR%201=1",
		nil,
	)
	profileLiveRequireStatus(t, response, http.StatusBadRequest)
	response = profileLiveJSONRequest(
		t,
		firstClient,
		http.MethodGet,
		baseURL+"/api/accounts/profiles/search?sort=recommended%27%3BDELETE%20FROM%20users%3B--",
		nil,
	)
	profileLiveRequireStatus(t, response, http.StatusBadRequest)

	response = profileLiveMultipartRequest(
		t,
		firstClient,
		baseURL+"/api/accounts/avatar",
		"avatar",
		[][]byte{[]byte("<script>alert(1)</script>")},
	)
	profileLiveRequireStatus(t, response, http.StatusUnsupportedMediaType)
	response = profileLiveMultipartRequest(
		t,
		firstClient,
		baseURL+"/api/accounts/photos",
		"photos",
		[][]byte{[]byte("not an image")},
	)
	profileLiveRequireStatus(t, response, http.StatusUnsupportedMediaType)
	response = profileLiveMultipartRequest(
		t,
		firstClient,
		baseURL+"/api/accounts/avatar",
		"avatar",
		[][]byte{make([]byte, (5<<20)+1)},
	)
	profileLiveRequireStatus(t, response, http.StatusRequestEntityTooLarge)

	response = profileLiveMultipartRequest(
		t,
		firstClient,
		baseURL+"/api/accounts/photos",
		"photos",
		[][]byte{profileLivePNG(t, 200)},
	)
	body := profileLiveRequireStatus(t, response, http.StatusCreated)
	var uploaded profileLivePhotosResponse
	if err := json.Unmarshal(body, &uploaded); err != nil || len(uploaded.Photos) != 1 {
		t.Fatalf("decode ownership photo: %v; body = %q", err, body)
	}
	photoID := uploaded.Photos[0].ID
	t.Cleanup(func() {
		cleanupResponse := profileLiveJSONRequest(
			t,
			firstClient,
			http.MethodDelete,
			fmt.Sprintf("%s/api/accounts/photos/%d", baseURL, photoID),
			nil,
		)
		if cleanupResponse.StatusCode != http.StatusNoContent && cleanupResponse.StatusCode != http.StatusNotFound {
			cleanupBody, _ := io.ReadAll(cleanupResponse.Body)
			cleanupResponse.Body.Close()
			t.Errorf("cleanup photo status = %d; body = %q", cleanupResponse.StatusCode, cleanupBody)
			return
		}
		cleanupResponse.Body.Close()
	})
	response = profileLiveJSONRequest(
		t,
		secondClient,
		http.MethodDelete,
		fmt.Sprintf("%s/api/accounts/photos/%d", baseURL, photoID),
		nil,
	)
	profileLiveRequireStatus(t, response, http.StatusNotFound)
	response = profileLiveJSONRequest(
		t,
		firstClient,
		http.MethodDelete,
		fmt.Sprintf("%s/api/accounts/photos/%d", baseURL, photoID),
		nil,
	)
	profileLiveRequireStatus(t, response, http.StatusNoContent)
}

func TestLiveRejectsMalformedImageContent(t *testing.T) {
	baseURL := strings.TrimRight(os.Getenv("SECURITY_E2E_BASE_URL"), "/")
	seedPassword := os.Getenv("SECURITY_E2E_SEED_PASSWORD")
	if baseURL == "" || seedPassword == "" {
		t.Skip("set SECURITY_E2E_BASE_URL and SECURITY_E2E_SEED_PASSWORD to run malformed image test")
	}
	client, status, _ := accountLiveLoginStatus(t, baseURL, "seed_0203", seedPassword)
	if status != http.StatusOK {
		t.Fatalf("seed login status = %d, want %d", status, http.StatusOK)
	}
	malformedPNG := append(
		[]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'},
		[]byte("<script>alert('not a real image')</script>")...,
	)
	response := profileLiveMultipartRequest(
		t,
		client,
		baseURL+"/api/accounts/avatar",
		"avatar",
		[][]byte{malformedPNG},
	)
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	response.Body.Close()
	if err != nil {
		t.Fatalf("read malformed image response: %v", err)
	}
	if response.StatusCode == http.StatusOK {
		restoreResponse := profileLiveMultipartRequest(
			t,
			client,
			baseURL+"/api/accounts/avatar",
			"avatar",
			[][]byte{profileLivePNG(t, 220)},
		)
		profileLiveRequireStatus(t, restoreResponse, http.StatusOK)
	}
	if response.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("malformed image status = %d, want %d; body = %q", response.StatusCode, http.StatusUnsupportedMediaType, body)
	}
}
