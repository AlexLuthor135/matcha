package e2e

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

const profileLiveTimeout = 10 * time.Second

type profileLiveMailpitMessages struct {
	Messages []struct {
		Subject string `json:"Subject"`
		To      []struct {
			Address string `json:"Address"`
		} `json:"To"`
		Snippet string `json:"Snippet"`
	} `json:"messages"`
}

type profileLiveAvatarResponse struct {
	AvatarURL string `json:"avatar_url"`
}

type profileLivePhoto struct {
	ID  uint   `json:"id"`
	URL string `json:"url"`
}

type profileLivePhotosResponse struct {
	Photos []profileLivePhoto `json:"photos"`
}

type profileLivePrivateProfile struct {
	Avatar            string             `json:"avatar"`
	Photos            []profileLivePhoto `json:"photos"`
	LocationSource    string             `json:"location_source"`
	LocationName      string             `json:"location_name"`
	LocationConsentAt *time.Time         `json:"location_consent_at"`
	Latitude          *float64           `json:"latitude"`
	Longitude         *float64           `json:"longitude"`
}

func profileLiveClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("create profile live cookie jar: %v", err)
	}
	return &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: true,
			},
		},
		Timeout: profileLiveTimeout,
	}
}

func profileLiveJSONRequest(
	t *testing.T,
	client *http.Client,
	method string,
	requestURL string,
	body any,
) *http.Response {
	t.Helper()
	var requestBody io.Reader
	if body != nil {
		encodedBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("encode profile live request: %v", err)
		}
		requestBody = bytes.NewReader(encodedBody)
	}
	request, err := http.NewRequest(method, requestURL, requestBody)
	if err != nil {
		t.Fatalf("create profile live request: %v", err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("send profile live request: %v", err)
	}
	return response
}

func profileLiveRequireStatus(t *testing.T, response *http.Response, wantStatus int) []byte {
	t.Helper()
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		t.Fatalf("read profile live response: %v", err)
	}
	if response.StatusCode != wantStatus {
		t.Fatalf("status = %d, want %d; body = %q", response.StatusCode, wantStatus, body)
	}
	return body
}

func profileLivePNG(t *testing.T, variant uint8) []byte {
	t.Helper()
	imageData := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			imageData.SetRGBA(x, y, color.RGBA{
				R: variant,
				G: uint8(x * 20),
				B: uint8(y * 20),
				A: 255,
			})
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, imageData); err != nil {
		t.Fatalf("encode profile live PNG: %v", err)
	}
	return encoded.Bytes()
}

func profileLiveMultipartRequest(
	t *testing.T,
	client *http.Client,
	requestURL string,
	fieldName string,
	files [][]byte,
) *http.Response {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for index, fileData := range files {
		part, err := writer.CreateFormFile(fieldName, fmt.Sprintf("image-%d.png", index+1))
		if err != nil {
			t.Fatalf("create multipart image: %v", err)
		}
		if _, err := part.Write(fileData); err != nil {
			t.Fatalf("write multipart image: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	request, err := http.NewRequest(http.MethodPost, requestURL, &body)
	if err != nil {
		t.Fatalf("create multipart request: %v", err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("send multipart request: %v", err)
	}
	return response
}

func profileLiveVerificationToken(
	t *testing.T,
	mailpitBaseURL string,
	email string,
	subjectContains string,
) string {
	t.Helper()
	tokenPattern := regexp.MustCompile(`token=([A-Za-z0-9_-]+)`)
	deadline := time.Now().Add(profileLiveTimeout)
	for time.Now().Before(deadline) {
		response, err := http.Get(strings.TrimRight(mailpitBaseURL, "/") + "/api/v1/messages")
		if err == nil {
			var messages profileLiveMailpitMessages
			decodeErr := json.NewDecoder(response.Body).Decode(&messages)
			response.Body.Close()
			if decodeErr == nil {
				for _, message := range messages.Messages {
					if !strings.Contains(message.Subject, subjectContains) {
						continue
					}
					for _, recipient := range message.To {
						if recipient.Address != email {
							continue
						}
						matches := tokenPattern.FindStringSubmatch(message.Snippet)
						if len(matches) == 2 {
							return matches[1]
						}
					}
				}
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("verification email for %s was not received", email)
	return ""
}

func profileLiveRegisterAndLogin(
	t *testing.T,
	baseURL string,
	mailpitBaseURL string,
) (*http.Client, uint) {
	t.Helper()
	suffix := fmt.Sprintf("%x", time.Now().UnixNano())
	if len(suffix) > 10 {
		suffix = suffix[len(suffix)-10:]
	}
	userName := "livep_" + suffix
	email := userName + "@example.invalid"
	password := "LiveProfile9!GreenOrbit"
	client := profileLiveClient(t)
	response := profileLiveJSONRequest(t, client, http.MethodPost, baseURL+"/api/register", map[string]string{
		"user_name":  userName,
		"first_name": "Live",
		"last_name":  "Profile",
		"email":      email,
		"password":   password,
	})
	profileLiveRequireStatus(t, response, http.StatusCreated)
	token := profileLiveVerificationToken(t, mailpitBaseURL, email, "Verify")
	response = profileLiveJSONRequest(
		t,
		client,
		http.MethodGet,
		baseURL+"/api/verify-email?token="+url.QueryEscape(token),
		nil,
	)
	profileLiveRequireStatus(t, response, http.StatusOK)
	response = profileLiveJSONRequest(t, client, http.MethodPost, baseURL+"/api/login", map[string]string{
		"login":    userName,
		"password": password,
	})
	profileLiveRequireStatus(t, response, http.StatusOK)
	response = profileLiveJSONRequest(t, client, http.MethodGet, baseURL+"/api/accounts/verify_login", nil)
	body := profileLiveRequireStatus(t, response, http.StatusOK)
	var verified struct {
		ID uint `json:"id"`
	}
	if err := json.Unmarshal(body, &verified); err != nil || verified.ID == 0 {
		t.Fatalf("decode live profile user ID: %v; body = %q", err, body)
	}
	return client, verified.ID
}

func TestLiveProfileLocationUploadsFeedAndSearch(t *testing.T) {
	baseURL := strings.TrimRight(os.Getenv("PROFILE_E2E_BASE_URL"), "/")
	mailpitBaseURL := strings.TrimRight(os.Getenv("PROFILE_E2E_MAILPIT_URL"), "/")
	if baseURL == "" || mailpitBaseURL == "" {
		t.Skip("set PROFILE_E2E_BASE_URL and PROFILE_E2E_MAILPIT_URL to run live profile test")
	}
	client, _ := profileLiveRegisterAndLogin(t, baseURL, mailpitBaseURL)

	manualLocation := map[string]any{
		"source":    "manual",
		"name":      "Berlin",
		"latitude":  52.52,
		"longitude": 13.405,
		"consent":   false,
	}
	completion := map[string]any{
		"gender":      "Male",
		"preferences": "Everyone",
		"bio":         "Live profile integration test",
		"interests":   []string{"coding", "travel"},
		"birth_date":  "1995-06-15",
		"location":    manualLocation,
	}
	response := profileLiveJSONRequest(t, client, http.MethodPost, baseURL+"/api/accounts/profile/complete", completion)
	profileLiveRequireStatus(t, response, http.StatusBadRequest)

	response = profileLiveMultipartRequest(
		t,
		client,
		baseURL+"/api/accounts/avatar",
		"avatar",
		[][]byte{profileLivePNG(t, 40)},
	)
	avatarBody := profileLiveRequireStatus(t, response, http.StatusOK)
	var avatarResponse profileLiveAvatarResponse
	if err := json.Unmarshal(avatarBody, &avatarResponse); err != nil || avatarResponse.AvatarURL == "" {
		t.Fatalf("decode avatar response: %v; body = %q", err, avatarBody)
	}

	response = profileLiveJSONRequest(t, client, http.MethodPost, baseURL+"/api/accounts/profile/complete", completion)
	profileLiveRequireStatus(t, response, http.StatusOK)

	photoFiles := [][]byte{
		profileLivePNG(t, 60),
		profileLivePNG(t, 80),
		profileLivePNG(t, 100),
		profileLivePNG(t, 120),
	}
	response = profileLiveMultipartRequest(t, client, baseURL+"/api/accounts/photos", "photos", photoFiles)
	photosBody := profileLiveRequireStatus(t, response, http.StatusCreated)
	var photosResponse profileLivePhotosResponse
	if err := json.Unmarshal(photosBody, &photosResponse); err != nil {
		t.Fatalf("decode uploaded photos: %v", err)
	}
	if len(photosResponse.Photos) != 4 {
		t.Fatalf("uploaded photo count = %d, want 4", len(photosResponse.Photos))
	}
	response = profileLiveMultipartRequest(
		t,
		client,
		baseURL+"/api/accounts/photos",
		"photos",
		[][]byte{profileLivePNG(t, 140)},
	)
	profileLiveRequireStatus(t, response, http.StatusConflict)

	response = profileLiveJSONRequest(t, client, http.MethodGet, baseURL+"/api/accounts/profile", nil)
	profileBody := profileLiveRequireStatus(t, response, http.StatusOK)
	var profile profileLivePrivateProfile
	if err := json.Unmarshal(profileBody, &profile); err != nil {
		t.Fatalf("decode private profile: %v", err)
	}
	if profile.Avatar != avatarResponse.AvatarURL || len(profile.Photos) != 4 {
		t.Fatalf("private profile avatar/photos = %q/%d", profile.Avatar, len(profile.Photos))
	}
	if profile.LocationSource != "manual" || profile.LocationName != "Berlin" || profile.LocationConsentAt != nil {
		t.Fatalf("manual location = source %q, name %q, consent %v", profile.LocationSource, profile.LocationName, profile.LocationConsentAt)
	}

	baseParsed, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("parse profile E2E base URL: %v", err)
	}
	assetURL := baseParsed.Scheme + "://" + baseParsed.Host + avatarResponse.AvatarURL
	assetResponse := profileLiveJSONRequest(t, client, http.MethodGet, assetURL, nil)
	profileLiveRequireStatus(t, assetResponse, http.StatusOK)

	response = profileLiveJSONRequest(t, client, http.MethodPatch, baseURL+"/api/accounts/profile", map[string]any{
		"location": map[string]any{
			"source":    "gps",
			"latitude":  48.8566,
			"longitude": 2.3522,
			"consent":   false,
		},
	})
	profileLiveRequireStatus(t, response, http.StatusBadRequest)
	response = profileLiveJSONRequest(t, client, http.MethodPatch, baseURL+"/api/accounts/profile", map[string]any{
		"location": map[string]any{
			"source":    "gps",
			"latitude":  48.8566,
			"longitude": 2.3522,
			"consent":   true,
		},
	})
	profileLiveRequireStatus(t, response, http.StatusOK)
	response = profileLiveJSONRequest(t, client, http.MethodGet, baseURL+"/api/accounts/profile", nil)
	profileBody = profileLiveRequireStatus(t, response, http.StatusOK)
	if err := json.Unmarshal(profileBody, &profile); err != nil {
		t.Fatalf("decode GPS profile: %v", err)
	}
	if profile.LocationSource != "gps" || profile.LocationName != "" || profile.LocationConsentAt == nil {
		t.Fatalf("GPS location = source %q, name %q, consent %v", profile.LocationSource, profile.LocationName, profile.LocationConsentAt)
	}

	response = profileLiveJSONRequest(
		t,
		client,
		http.MethodGet,
		baseURL+"/api/accounts/profiles/feed?limit=20&sort=recommended",
		nil,
	)
	feedBody := profileLiveRequireStatus(t, response, http.StatusOK)
	var feed struct {
		Profiles []map[string]any `json:"profiles"`
	}
	if err := json.Unmarshal(feedBody, &feed); err != nil || len(feed.Profiles) == 0 {
		t.Fatalf("decode non-empty feed: %v; body = %q", err, feedBody)
	}
	for _, forbidden := range []string{"email", "password", "latitude", "longitude", "birth_date"} {
		if _, exposed := feed.Profiles[0][forbidden]; exposed {
			t.Fatalf("feed profile exposes private field %q", forbidden)
		}
	}

	response = profileLiveJSONRequest(
		t,
		client,
		http.MethodGet,
		baseURL+"/api/accounts/profiles/search?limit=50&min_age=18&max_age=80&max_distance=20000&min_fame=0&interests=coding&sort=interests",
		nil,
	)
	searchBody := profileLiveRequireStatus(t, response, http.StatusOK)
	var search struct {
		Profiles []map[string]any `json:"profiles"`
	}
	if err := json.Unmarshal(searchBody, &search); err != nil || len(search.Profiles) == 0 {
		t.Fatalf("decode non-empty search: %v; body = %q", err, searchBody)
	}
	response = profileLiveJSONRequest(
		t,
		client,
		http.MethodGet,
		baseURL+"/api/accounts/profiles/search?sort=unknown",
		nil,
	)
	profileLiveRequireStatus(t, response, http.StatusBadRequest)

	response = profileLiveJSONRequest(
		t,
		client,
		http.MethodDelete,
		fmt.Sprintf("%s/api/accounts/photos/%d", baseURL, photosResponse.Photos[0].ID),
		nil,
	)
	profileLiveRequireStatus(t, response, http.StatusNoContent)
	response = profileLiveMultipartRequest(
		t,
		client,
		baseURL+"/api/accounts/photos",
		"photos",
		[][]byte{profileLivePNG(t, 160)},
	)
	profileLiveRequireStatus(t, response, http.StatusCreated)
}
