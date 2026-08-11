package api

import (
	"backend/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateTokensSetsValidCookiesAndResponse(t *testing.T) {
	const secret = "generate-token-test-secret"
	t.Setenv("JWT_SECRET", secret)
	refreshExpiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	response := httptest.NewRecorder()

	failed := GenerateTokens(response, models.User{
		ID:          42,
		IsCompleted: true,
	}, "session-42", refreshExpiresAt)

	if failed {
		t.Fatal("GenerateTokens() reported failure")
	}
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var body LoginResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Message != "Login successful" || !body.IsCompleted {
		t.Fatalf("response = %+v", body)
	}

	cookies := make(map[string]*http.Cookie)
	for _, cookie := range response.Result().Cookies() {
		cookies[cookie.Name] = cookie
	}
	for _, name := range []string{"access_token", "refresh_token"} {
		cookie := cookies[name]
		if cookie == nil {
			t.Fatalf("%s cookie is missing", name)
		}
		if !cookie.HttpOnly || !cookie.Secure || cookie.Path != "/" || cookie.SameSite != http.SameSiteLaxMode {
			t.Fatalf("%s cookie attributes = %+v", name, cookie)
		}
	}

	accessClaims := jwt.MapClaims{}
	accessToken, err := jwt.ParseWithClaims(cookies["access_token"].Value, accessClaims, func(*jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !accessToken.Valid {
		t.Fatalf("parse access token: %v", err)
	}
	if accessClaims["user_id"] != float64(42) || accessClaims["type"] != "access" {
		t.Fatalf("access claims = %+v", accessClaims)
	}

	refreshClaims, err := ParseRefreshToken(cookies["refresh_token"].Value)
	if err != nil {
		t.Fatalf("parse refresh token: %v", err)
	}
	if refreshClaims.UserID != 42 || refreshClaims.SessionID != "session-42" {
		t.Fatalf("refresh claims = %+v", refreshClaims)
	}
	if !refreshClaims.ExpiresAt.Time.Equal(refreshExpiresAt) {
		t.Fatalf("refresh expiry = %s, want %s", refreshClaims.ExpiresAt.Time, refreshExpiresAt)
	}
}

func TestGenerateTokensRequiresJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	response := httptest.NewRecorder()

	failed := GenerateTokens(response, models.User{ID: 42}, "session-42", time.Now().Add(time.Hour))

	if !failed {
		t.Fatal("GenerateTokens() did not report failure")
	}
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
}

func TestClearAuthCookies(t *testing.T) {
	response := httptest.NewRecorder()

	ClearAuthCookies(response)

	cookies := response.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("cookie count = %d, want 2", len(cookies))
	}
	for _, cookie := range cookies {
		if cookie.Name != "access_token" && cookie.Name != "refresh_token" {
			t.Fatalf("unexpected cookie %q", cookie.Name)
		}
		if cookie.Value != "" || cookie.MaxAge != -1 || !cookie.HttpOnly || !cookie.Secure || cookie.Path != "/" {
			t.Fatalf("deletion cookie = %+v", cookie)
		}
	}
}
