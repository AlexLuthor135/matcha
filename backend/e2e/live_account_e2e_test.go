package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

type accountLiveUser struct {
	UserName string
	Email    string
	Password string
	Client   *http.Client
	ID       uint
}

func accountLiveRegisterAndLogin(
	t *testing.T,
	baseURL string,
	mailpitBaseURL string,
) accountLiveUser {
	t.Helper()
	suffix := fmt.Sprintf("%x", time.Now().UnixNano())
	if len(suffix) > 10 {
		suffix = suffix[len(suffix)-10:]
	}
	user := accountLiveUser{
		UserName: "livea_" + suffix,
		Password: "LiveAccount9!GreenOrbit",
		Client:   profileLiveClient(t),
	}
	user.Email = user.UserName + "@example.invalid"
	response := profileLiveJSONRequest(t, user.Client, http.MethodPost, baseURL+"/api/register", map[string]string{
		"user_name":  user.UserName,
		"first_name": "Live",
		"last_name":  "Account",
		"email":      user.Email,
		"password":   user.Password,
	})
	profileLiveRequireStatus(t, response, http.StatusCreated)
	token := profileLiveVerificationToken(t, mailpitBaseURL, user.Email, "Verify")
	response = profileLiveJSONRequest(
		t,
		user.Client,
		http.MethodGet,
		baseURL+"/api/verify-email?token="+url.QueryEscape(token),
		nil,
	)
	profileLiveRequireStatus(t, response, http.StatusOK)
	response = profileLiveJSONRequest(t, user.Client, http.MethodPost, baseURL+"/api/login", map[string]string{
		"login":    user.UserName,
		"password": user.Password,
	})
	profileLiveRequireStatus(t, response, http.StatusOK)
	response = profileLiveJSONRequest(t, user.Client, http.MethodGet, baseURL+"/api/accounts/verify_login", nil)
	body := profileLiveRequireStatus(t, response, http.StatusOK)
	var verified struct {
		ID uint `json:"id"`
	}
	if err := json.Unmarshal(body, &verified); err != nil || verified.ID == 0 {
		t.Fatalf("decode account live user: %v; body = %q", err, body)
	}
	user.ID = verified.ID
	return user
}

func accountLiveLoginStatus(
	t *testing.T,
	baseURL string,
	login string,
	password string,
) (*http.Client, int, []byte) {
	t.Helper()
	client := profileLiveClient(t)
	response := profileLiveJSONRequest(t, client, http.MethodPost, baseURL+"/api/login", map[string]string{
		"login":    login,
		"password": password,
	})
	body := profileLiveRequireStatus(t, response, response.StatusCode)
	return client, response.StatusCode, body
}

func accountLiveRefreshToken(t *testing.T, client *http.Client, baseURL string) string {
	t.Helper()
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("parse account live URL: %v", err)
	}
	for _, cookie := range client.Jar.Cookies(parsedURL) {
		if cookie.Name == "refresh_token" {
			return cookie.Value
		}
	}
	t.Fatal("refresh_token cookie is missing")
	return ""
}

func accountLiveRefreshWithRawToken(
	t *testing.T,
	baseURL string,
	rawToken string,
) int {
	t.Helper()
	client := profileLiveClient(t)
	request, err := http.NewRequest(http.MethodPost, baseURL+"/api/accounts/token/refresh", nil)
	if err != nil {
		t.Fatalf("create raw refresh request: %v", err)
	}
	request.AddCookie(&http.Cookie{Name: "refresh_token", Value: rawToken})
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("send raw refresh request: %v", err)
	}
	profileLiveRequireStatus(t, response, response.StatusCode)
	return response.StatusCode
}

func TestLiveAccountEmailPasswordAndSessionLifecycle(t *testing.T) {
	baseURL := strings.TrimRight(os.Getenv("ACCOUNT_E2E_BASE_URL"), "/")
	mailpitBaseURL := strings.TrimRight(os.Getenv("ACCOUNT_E2E_MAILPIT_URL"), "/")
	if baseURL == "" || mailpitBaseURL == "" {
		t.Skip("set ACCOUNT_E2E_BASE_URL and ACCOUNT_E2E_MAILPIT_URL to run live account test")
	}
	user := accountLiveRegisterAndLogin(t, baseURL, mailpitBaseURL)
	newEmail := "updated_" + user.Email

	response := profileLiveJSONRequest(t, user.Client, http.MethodPatch, baseURL+"/api/accounts/user", map[string]string{
		"first_name": "Updated",
		"email":      newEmail,
	})
	body := profileLiveRequireStatus(t, response, http.StatusOK)
	var updateResult struct {
		VerificationRequired bool   `json:"verification_required"`
		PendingEmail         string `json:"pending_email"`
	}
	if err := json.Unmarshal(body, &updateResult); err != nil {
		t.Fatalf("decode account update: %v", err)
	}
	if !updateResult.VerificationRequired || updateResult.PendingEmail != newEmail {
		t.Fatalf("account update result = %+v", updateResult)
	}
	response = profileLiveJSONRequest(t, user.Client, http.MethodGet, baseURL+"/api/accounts/profile", nil)
	body = profileLiveRequireStatus(t, response, http.StatusOK)
	var privateProfile struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
	}
	if err := json.Unmarshal(body, &privateProfile); err != nil {
		t.Fatalf("decode account profile before email verification: %v", err)
	}
	if privateProfile.Email != user.Email || privateProfile.FirstName != "Updated" {
		t.Fatalf("profile before email verification = %+v", privateProfile)
	}

	emailToken := profileLiveVerificationToken(t, mailpitBaseURL, newEmail, "Verify")
	response = profileLiveJSONRequest(
		t,
		user.Client,
		http.MethodGet,
		baseURL+"/api/verify-email?token="+url.QueryEscape(emailToken),
		nil,
	)
	profileLiveRequireStatus(t, response, http.StatusOK)
	response = profileLiveJSONRequest(t, user.Client, http.MethodGet, baseURL+"/api/accounts/profile", nil)
	body = profileLiveRequireStatus(t, response, http.StatusOK)
	if err := json.Unmarshal(body, &privateProfile); err != nil || privateProfile.Email != newEmail {
		t.Fatalf("verified profile email = %q; decode error = %v", privateProfile.Email, err)
	}

	originalRefreshToken := accountLiveRefreshToken(t, user.Client, baseURL)
	firstNewPassword := "ChangedAccount8!BlueOrbit"
	response = profileLiveJSONRequest(t, user.Client, http.MethodPatch, baseURL+"/api/accounts/user/password", map[string]string{
		"current_password": user.Password,
		"new_password":     firstNewPassword,
	})
	profileLiveRequireStatus(t, response, http.StatusOK)
	if got := accountLiveRefreshWithRawToken(t, baseURL, originalRefreshToken); got != http.StatusUnauthorized {
		t.Fatalf("refresh after password change status = %d, want %d", got, http.StatusUnauthorized)
	}
	_, status, _ := accountLiveLoginStatus(t, baseURL, user.UserName, user.Password)
	if status != http.StatusUnauthorized {
		t.Fatalf("old password login status = %d, want %d", status, http.StatusUnauthorized)
	}
	passwordClient, status, _ := accountLiveLoginStatus(t, baseURL, newEmail, firstNewPassword)
	if status != http.StatusOK {
		t.Fatalf("new password/email login status = %d, want %d", status, http.StatusOK)
	}
	passwordRefreshToken := accountLiveRefreshToken(t, passwordClient, baseURL)

	response = profileLiveJSONRequest(t, profileLiveClient(t), http.MethodPost, baseURL+"/api/password/forgot", map[string]string{
		"email": newEmail,
	})
	profileLiveRequireStatus(t, response, http.StatusOK)
	resetToken := profileLiveVerificationToken(t, mailpitBaseURL, newEmail, "Reset")
	finalPassword := "ResetAccount7!RedOrbit"
	response = profileLiveJSONRequest(t, profileLiveClient(t), http.MethodPost, baseURL+"/api/password/reset", map[string]string{
		"token":       resetToken,
		"newPassword": finalPassword,
	})
	profileLiveRequireStatus(t, response, http.StatusOK)
	if got := accountLiveRefreshWithRawToken(t, baseURL, passwordRefreshToken); got != http.StatusUnauthorized {
		t.Fatalf("refresh after password reset status = %d, want %d", got, http.StatusUnauthorized)
	}
	response = profileLiveJSONRequest(t, profileLiveClient(t), http.MethodPost, baseURL+"/api/password/reset", map[string]string{
		"token":       resetToken,
		"newPassword": "AnotherAccount6!GoldOrbit",
	})
	profileLiveRequireStatus(t, response, http.StatusBadRequest)
	_, status, _ = accountLiveLoginStatus(t, baseURL, newEmail, finalPassword)
	if status != http.StatusOK {
		t.Fatalf("final password login status = %d, want %d", status, http.StatusOK)
	}
}
