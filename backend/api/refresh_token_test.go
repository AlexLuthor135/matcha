package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type refreshSessionValidatorFunc func(context.Context, string, uint) (bool, error)

func (fn refreshSessionValidatorFunc) ValidateAuthSession(
	ctx context.Context,
	sessionID string,
	userID uint,
) (bool, error) {
	return fn(ctx, sessionID, userID)
}

func signedRefreshTokenForTest(
	t *testing.T,
	secret string,
	userID uint,
	sessionID string,
	tokenType string,
	expiresAt time.Time,
) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    userID,
		"session_id": sessionID,
		"type":       tokenType,
		"exp":        expiresAt.Unix(),
	})
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign refresh token: %v", err)
	}
	return signedToken
}

func TestParseRefreshToken(t *testing.T) {
	const secret = "refresh-token-test-secret"
	t.Setenv("JWT_SECRET", secret)

	rawToken := signedRefreshTokenForTest(
		t,
		secret,
		42,
		"session-42",
		"refresh",
		time.Now().Add(time.Hour),
	)
	claims, err := ParseRefreshToken(rawToken)
	if err != nil {
		t.Fatalf("ParseRefreshToken() unexpected error: %v", err)
	}
	if claims.UserID != 42 || claims.SessionID != "session-42" || claims.TokenType != "refresh" {
		t.Fatalf("ParseRefreshToken() claims = %+v", claims)
	}
}

func TestParseRefreshTokenRejectsInvalidTokens(t *testing.T) {
	const secret = "refresh-token-test-secret"
	t.Setenv("JWT_SECRET", secret)

	tests := []struct {
		name      string
		userID    uint
		sessionID string
		tokenType string
		expiresAt time.Time
	}{
		{name: "access token", userID: 42, sessionID: "session-42", tokenType: "access", expiresAt: time.Now().Add(time.Hour)},
		{name: "missing user", sessionID: "session-42", tokenType: "refresh", expiresAt: time.Now().Add(time.Hour)},
		{name: "missing session", userID: 42, tokenType: "refresh", expiresAt: time.Now().Add(time.Hour)},
		{name: "expired", userID: 42, sessionID: "session-42", tokenType: "refresh", expiresAt: time.Now().Add(-time.Hour)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rawToken := signedRefreshTokenForTest(
				t,
				secret,
				test.userID,
				test.sessionID,
				test.tokenType,
				test.expiresAt,
			)
			if _, err := ParseRefreshToken(rawToken); !errors.Is(err, ErrInvalidRefreshToken) {
				t.Fatalf("ParseRefreshToken() error = %v, want %v", err, ErrInvalidRefreshToken)
			}
		})
	}
}

func TestRefreshTokenValidatesSessionAndSetsAccessCookie(t *testing.T) {
	const secret = "refresh-token-test-secret"
	t.Setenv("JWT_SECRET", secret)
	rawToken := signedRefreshTokenForTest(
		t,
		secret,
		42,
		"session-42",
		"refresh",
		time.Now().Add(time.Hour),
	)

	validator := refreshSessionValidatorFunc(func(_ context.Context, sessionID string, userID uint) (bool, error) {
		if sessionID != "session-42" || userID != 42 {
			t.Fatalf("ValidateAuthSession() got session %q and user %d", sessionID, userID)
		}
		return true, nil
	})
	request := httptest.NewRequest(http.MethodPost, "/api/accounts/token/refresh", nil)
	request.AddCookie(&http.Cookie{Name: "refresh_token", Value: rawToken})
	response := httptest.NewRecorder()

	RefreshToken(validator).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	var accessCookie *http.Cookie
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == "access_token" {
			accessCookie = cookie
			break
		}
	}
	if accessCookie == nil || accessCookie.Value == "" {
		t.Fatal("response does not contain an access_token cookie")
	}
}

func TestRefreshTokenRejectsRevokedSession(t *testing.T) {
	const secret = "refresh-token-test-secret"
	t.Setenv("JWT_SECRET", secret)
	rawToken := signedRefreshTokenForTest(
		t,
		secret,
		42,
		"revoked-session",
		"refresh",
		time.Now().Add(time.Hour),
	)
	request := httptest.NewRequest(http.MethodPost, "/api/accounts/token/refresh", nil)
	request.AddCookie(&http.Cookie{Name: "refresh_token", Value: rawToken})
	response := httptest.NewRecorder()

	RefreshToken(refreshSessionValidatorFunc(
		func(context.Context, string, uint) (bool, error) {
			return false, nil
		},
	)).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}
