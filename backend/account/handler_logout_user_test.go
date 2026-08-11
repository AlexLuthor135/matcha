package account

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func logoutRefreshTokenForTest(t *testing.T, secret string, userID uint, sessionID string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    userID,
		"session_id": sessionID,
		"type":       "refresh",
		"exp":        time.Now().Add(time.Hour).Unix(),
	})
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign logout refresh token: %v", err)
	}
	return signedToken
}

func TestLogoutUserRevokesSessionAndClearsCookies(t *testing.T) {
	const secret = "logout-handler-test-secret"
	t.Setenv("JWT_SECRET", secret)
	revoked := false
	repository := &fakeUserRepository{
		revokeAuthSessionFn: func(_ context.Context, sessionID string, userID uint) error {
			if sessionID != "session-42" || userID != 42 {
				t.Fatalf("RevokeAuthSession() got session %q and user %d", sessionID, userID)
			}
			revoked = true
			return nil
		},
	}
	handler := NewHandler(NewService(repository, nil))
	request := httptest.NewRequest(http.MethodPost, "/api/accounts/logout/", nil)
	request.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: logoutRefreshTokenForTest(t, secret, 42, "session-42"),
	})
	response := httptest.NewRecorder()

	handler.LogoutUser(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !revoked {
		t.Fatal("RevokeAuthSession() was not called")
	}
	deletedCookies := make(map[string]bool)
	for _, cookie := range response.Result().Cookies() {
		if cookie.MaxAge < 0 {
			deletedCookies[cookie.Name] = true
		}
	}
	if !deletedCookies["access_token"] || !deletedCookies["refresh_token"] {
		t.Fatalf("deleted cookies = %v, want access and refresh tokens", deletedCookies)
	}
}

func TestLogoutUserWithoutRefreshCookieStillClearsCookies(t *testing.T) {
	handler := NewHandler(NewService(&fakeUserRepository{}, nil))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/accounts/logout/", nil)

	handler.LogoutUser(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if len(response.Result().Cookies()) != 2 {
		t.Fatalf("cookie count = %d, want 2", len(response.Result().Cookies()))
	}
}
