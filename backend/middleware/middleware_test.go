package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func accessTokenForTest(
	t *testing.T,
	secret string,
	userID any,
	tokenType string,
) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"type":    tokenType,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign access token: %v", err)
	}
	return signedToken
}

func TestAuthMiddlewareAcceptsValidAccessToken(t *testing.T) {
	const secret = "auth-middleware-test-secret"
	t.Setenv("JWT_SECRET", secret)

	nextCalled := false
	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		userID, ok := r.Context().Value(UserIDKey).(uint)
		if !ok || userID != 42 {
			t.Fatalf("context user ID = %v, want 42", r.Context().Value(UserIDKey))
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: accessTokenForTest(t, secret, 42, "access"),
	})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if !nextCalled {
		t.Fatal("next handler was not called")
	}
}

func TestAuthMiddlewareRejectsInvalidAuthentication(t *testing.T) {
	const secret = "auth-middleware-test-secret"
	t.Setenv("JWT_SECRET", secret)

	tests := []struct {
		name   string
		cookie string
	}{
		{name: "missing cookie"},
		{name: "invalid signature", cookie: accessTokenForTest(t, "another-secret", 42, "access")},
		{name: "refresh token", cookie: accessTokenForTest(t, secret, 42, "refresh")},
		{name: "missing user", cookie: accessTokenForTest(t, secret, nil, "access")},
		{name: "fractional user", cookie: accessTokenForTest(t, secret, 42.5, "access")},
		{name: "zero user", cookie: accessTokenForTest(t, secret, 0, "access")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			nextCalled := false
			handler := AuthMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				nextCalled = true
			}))
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if test.cookie != "" {
				request.AddCookie(&http.Cookie{Name: "access_token", Value: test.cookie})
			}
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
			}
			if nextCalled {
				t.Fatal("next handler was called")
			}
		})
	}
}

func TestAuthMiddlewareRequiresJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.AddCookie(&http.Cookie{Name: "access_token", Value: "present"})
	response := httptest.NewRecorder()

	AuthMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler was called")
	})).ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
}
