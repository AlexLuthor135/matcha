package realtime

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

const liveE2ETimeout = 10 * time.Second

type liveVerifyResponse struct {
	ID uint `json:"id"`
}

type liveSocketEvent struct {
	Type         string `json:"type"`
	Message      string `json:"message"`
	Notification struct {
		Type string `json:"type"`
	} `json:"notification"`
}

type liveNotification struct {
	ID       uint       `json:"id"`
	SenderID *uint      `json:"sender_id"`
	Type     string     `json:"type"`
	ReadAt   *time.Time `json:"read_at"`
}

type liveNotificationsResponse struct {
	Notifications []liveNotification `json:"notifications"`
}

type liveProfileViewer struct {
	ID uint `json:"id"`
}

type liveProfileViewersResponse struct {
	Viewers []liveProfileViewer `json:"viewers"`
}

func liveHTTPClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("create cookie jar: %v", err)
	}
	return &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: true,
			},
		},
		Timeout: liveE2ETimeout,
	}
}

func liveJSONRequest(
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
			t.Fatalf("encode %s request: %v", requestURL, err)
		}
		requestBody = bytes.NewReader(encodedBody)
	}
	request, err := http.NewRequest(method, requestURL, requestBody)
	if err != nil {
		t.Fatalf("create %s request: %v", requestURL, err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("send %s request: %v", requestURL, err)
	}
	return response
}

func requireLiveStatus(t *testing.T, response *http.Response, wantStatus int) {
	t.Helper()
	defer response.Body.Close()
	if response.StatusCode == wantStatus {
		return
	}
	body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	t.Fatalf("status = %d, want %d; body = %q", response.StatusCode, wantStatus, body)
}

func liveLogin(
	t *testing.T,
	baseURL string,
	login string,
	password string,
) (*http.Client, uint) {
	t.Helper()
	client := liveHTTPClient(t)
	response := liveJSONRequest(t, client, http.MethodPost, baseURL+"/api/login", map[string]string{
		"login":    login,
		"password": password,
	})
	requireLiveStatus(t, response, http.StatusOK)

	response = liveJSONRequest(t, client, http.MethodGet, baseURL+"/api/accounts/verify_login", nil)
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		t.Fatalf("verify status = %d; body = %q", response.StatusCode, body)
	}
	var verified liveVerifyResponse
	if err := json.NewDecoder(response.Body).Decode(&verified); err != nil {
		t.Fatalf("decode verify response: %v", err)
	}
	if verified.ID == 0 {
		t.Fatal("verified user ID is zero")
	}
	return client, verified.ID
}

func liveSetDecision(
	t *testing.T,
	client *http.Client,
	baseURL string,
	targetUserID uint,
	decision string,
) {
	t.Helper()
	response := liveJSONRequest(
		t,
		client,
		http.MethodPut,
		fmt.Sprintf("%s/api/accounts/profiles/%d/decision", baseURL, targetUserID),
		map[string]string{"decision": decision},
	)
	requireLiveStatus(t, response, http.StatusOK)
}

func liveNotificationsForUser(
	t *testing.T,
	client *http.Client,
	baseURL string,
) []liveNotification {
	t.Helper()
	response := liveJSONRequest(
		t,
		client,
		http.MethodGet,
		baseURL+"/api/accounts/notifications",
		nil,
	)
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		t.Fatalf("list notifications status = %d; body = %q", response.StatusCode, body)
	}
	var result liveNotificationsResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatalf("decode notification list: %v", err)
	}
	return result.Notifications
}

func liveProfileViewersForUser(
	t *testing.T,
	client *http.Client,
	baseURL string,
) []liveProfileViewer {
	t.Helper()
	response := liveJSONRequest(
		t,
		client,
		http.MethodGet,
		baseURL+"/api/accounts/profile/views",
		nil,
	)
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		t.Fatalf("list profile viewers status = %d; body = %q", response.StatusCode, body)
	}
	var result liveProfileViewersResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatalf("decode profile viewers: %v", err)
	}
	return result.Viewers
}

func liveChangeBlock(
	t *testing.T,
	client *http.Client,
	baseURL string,
	targetUserID uint,
	method string,
) {
	t.Helper()
	response := liveJSONRequest(
		t,
		client,
		method,
		fmt.Sprintf("%s/api/accounts/profiles/%d/block", baseURL, targetUserID),
		nil,
	)
	requireLiveStatus(t, response, http.StatusNoContent)
}

func liveWebSocket(
	t *testing.T,
	baseURL string,
	client *http.Client,
) *websocket.Conn {
	t.Helper()
	httpURL, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("parse live base URL: %v", err)
	}
	websocketURL := *httpURL
	if httpURL.Scheme == "https" {
		websocketURL.Scheme = "wss"
	} else {
		websocketURL.Scheme = "ws"
	}
	websocketURL.Path = strings.TrimRight(websocketURL.Path, "/") + "/api/accounts/ws"

	cookieParts := make([]string, 0)
	for _, cookie := range client.Jar.Cookies(httpURL) {
		cookieParts = append(cookieParts, cookie.Name+"="+cookie.Value)
	}
	header := make(http.Header)
	header.Set("Cookie", strings.Join(cookieParts, "; "))
	dialer := websocket.Dialer{
		HandshakeTimeout: liveE2ETimeout,
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true,
		},
	}
	connection, response, err := dialer.Dial(websocketURL.String(), header)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		t.Fatalf("connect %s: %v", websocketURL.String(), err)
	}
	return connection
}

func readLiveEvent(
	t *testing.T,
	connection *websocket.Conn,
	predicate func(liveSocketEvent) bool,
) liveSocketEvent {
	t.Helper()
	if err := connection.SetReadDeadline(time.Now().Add(liveE2ETimeout)); err != nil {
		t.Fatalf("set WebSocket read deadline: %v", err)
	}
	for {
		var event liveSocketEvent
		if err := connection.ReadJSON(&event); err != nil {
			t.Fatalf("read WebSocket event: %v", err)
		}
		if predicate(event) {
			return event
		}
	}
}

func TestLiveChatMessageNotificationAndUnlike(t *testing.T) {
	baseURL := strings.TrimRight(os.Getenv("REALTIME_E2E_BASE_URL"), "/")
	password := os.Getenv("REALTIME_E2E_SEED_PASSWORD")
	if baseURL == "" || password == "" {
		t.Skip("set REALTIME_E2E_BASE_URL and REALTIME_E2E_SEED_PASSWORD to run live WebSocket test")
	}

	firstClient, firstUserID := liveLogin(t, baseURL, "seed_0001", password)
	secondClient, secondUserID := liveLogin(t, baseURL, "seed_0002", password)
	liveSetDecision(t, firstClient, baseURL, secondUserID, "like")
	liveSetDecision(t, secondClient, baseURL, firstUserID, "like")
	t.Cleanup(func() {
		liveSetDecision(t, firstClient, baseURL, secondUserID, "like")
	})

	firstSocket := liveWebSocket(t, baseURL, firstClient)
	defer firstSocket.Close()
	secondSocket := liveWebSocket(t, baseURL, secondClient)
	defer secondSocket.Close()

	messageText := fmt.Sprintf("live-e2e-%d", time.Now().UnixNano())
	if err := firstSocket.WriteJSON(map[string]any{
		"type":         "chat_message",
		"recipient_id": secondUserID,
		"message":      messageText,
	}); err != nil {
		t.Fatalf("send live chat message: %v", err)
	}
	readLiveEvent(t, firstSocket, func(event liveSocketEvent) bool {
		return event.Type == "chat_message" && event.Message == messageText
	})
	readLiveEvent(t, secondSocket, func(event liveSocketEvent) bool {
		return event.Type == "chat_message" && event.Message == messageText
	})
	readLiveEvent(t, secondSocket, func(event liveSocketEvent) bool {
		return event.Type == "notification" && event.Notification.Type == "message"
	})

	liveSetDecision(t, firstClient, baseURL, secondUserID, "dislike")
	if err := firstSocket.WriteJSON(map[string]any{
		"type":         "chat_message",
		"recipient_id": secondUserID,
		"message":      "must be rejected after unlike",
	}); err != nil {
		t.Fatalf("send message after unlike: %v", err)
	}
	errorEvent := readLiveEvent(t, firstSocket, func(event liveSocketEvent) bool {
		return event.Type == "error"
	})
	if errorEvent.Message != "users are not matched" {
		t.Fatalf("error message = %q, want %q", errorEvent.Message, "users are not matched")
	}
}

func TestLiveRequiredNotificationAndBlockFlow(t *testing.T) {
	baseURL := strings.TrimRight(os.Getenv("REALTIME_E2E_BASE_URL"), "/")
	password := os.Getenv("REALTIME_E2E_SEED_PASSWORD")
	if baseURL == "" || password == "" {
		t.Skip("set REALTIME_E2E_BASE_URL and REALTIME_E2E_SEED_PASSWORD to run live notification test")
	}

	firstClient, firstUserID := liveLogin(t, baseURL, "seed_0101", password)
	secondClient, secondUserID := liveLogin(t, baseURL, "seed_0102", password)
	liveSetDecision(t, firstClient, baseURL, secondUserID, "dislike")
	liveSetDecision(t, secondClient, baseURL, firstUserID, "dislike")
	t.Cleanup(func() {
		liveChangeBlock(t, secondClient, baseURL, firstUserID, http.MethodDelete)
		liveSetDecision(t, firstClient, baseURL, secondUserID, "like")
		liveSetDecision(t, secondClient, baseURL, firstUserID, "like")
	})

	firstSocket := liveWebSocket(t, baseURL, firstClient)
	defer firstSocket.Close()
	secondSocket := liveWebSocket(t, baseURL, secondClient)
	defer secondSocket.Close()

	liveSetDecision(t, firstClient, baseURL, secondUserID, "like")
	readLiveEvent(t, secondSocket, func(event liveSocketEvent) bool {
		return event.Type == "notification" && event.Notification.Type == "like"
	})

	liveSetDecision(t, secondClient, baseURL, firstUserID, "like")
	readLiveEvent(t, firstSocket, func(event liveSocketEvent) bool {
		return event.Type == "notification" && event.Notification.Type == "like"
	})
	readLiveEvent(t, firstSocket, func(event liveSocketEvent) bool {
		return event.Type == "notification" && event.Notification.Type == "match"
	})

	profileResponse := liveJSONRequest(
		t,
		firstClient,
		http.MethodGet,
		fmt.Sprintf("%s/api/accounts/profiles/%d", baseURL, secondUserID),
		nil,
	)
	requireLiveStatus(t, profileResponse, http.StatusOK)
	readLiveEvent(t, secondSocket, func(event liveSocketEvent) bool {
		return event.Type == "notification" && event.Notification.Type == "profile_view"
	})

	viewers := liveProfileViewersForUser(t, secondClient, baseURL)
	foundViewer := false
	for _, viewer := range viewers {
		if viewer.ID == firstUserID {
			foundViewer = true
			break
		}
	}
	if !foundViewer {
		t.Fatalf("profile viewers do not contain user %d", firstUserID)
	}

	liveSetDecision(t, firstClient, baseURL, secondUserID, "dislike")
	readLiveEvent(t, secondSocket, func(event liveSocketEvent) bool {
		return event.Type == "notification" && event.Notification.Type == "unlike"
	})

	notifications := liveNotificationsForUser(t, secondClient, baseURL)
	requiredUnreadTypes := map[string]bool{
		"like":         false,
		"profile_view": false,
		"unlike":       false,
	}
	var profileViewNotificationID uint
	for _, notification := range notifications {
		if notification.SenderID == nil || *notification.SenderID != firstUserID {
			continue
		}
		if _, required := requiredUnreadTypes[notification.Type]; !required {
			continue
		}
		if notification.ReadAt != nil {
			t.Fatalf("new %s notification is already read", notification.Type)
		}
		requiredUnreadTypes[notification.Type] = true
		if notification.Type == "profile_view" {
			profileViewNotificationID = notification.ID
		}
	}
	for notificationType, found := range requiredUnreadTypes {
		if !found {
			t.Fatalf("missing unread %s notification from user %d", notificationType, firstUserID)
		}
	}

	markReadResponse := liveJSONRequest(
		t,
		secondClient,
		http.MethodPatch,
		fmt.Sprintf("%s/api/accounts/notifications/%d/read", baseURL, profileViewNotificationID),
		nil,
	)
	requireLiveStatus(t, markReadResponse, http.StatusOK)
	notifications = liveNotificationsForUser(t, secondClient, baseURL)
	markedRead := false
	for _, notification := range notifications {
		if notification.ID == profileViewNotificationID && notification.ReadAt != nil {
			markedRead = true
			break
		}
	}
	if !markedRead {
		t.Fatalf("notification %d was not marked read", profileViewNotificationID)
	}

	notificationCountBeforeBlock := len(notifications)
	liveChangeBlock(t, secondClient, baseURL, firstUserID, http.MethodPut)
	blockedProfileResponse := liveJSONRequest(
		t,
		firstClient,
		http.MethodGet,
		fmt.Sprintf("%s/api/accounts/profiles/%d", baseURL, secondUserID),
		nil,
	)
	requireLiveStatus(t, blockedProfileResponse, http.StatusNotFound)
	blockedLikeResponse := liveJSONRequest(
		t,
		firstClient,
		http.MethodPut,
		fmt.Sprintf("%s/api/accounts/profiles/%d/decision", baseURL, secondUserID),
		map[string]string{"decision": "like"},
	)
	requireLiveStatus(t, blockedLikeResponse, http.StatusNotFound)
	if got := len(liveNotificationsForUser(t, secondClient, baseURL)); got != notificationCountBeforeBlock {
		t.Fatalf("notification count after blocked actions = %d, want %d", got, notificationCountBeforeBlock)
	}
	for _, viewer := range liveProfileViewersForUser(t, secondClient, baseURL) {
		if viewer.ID == firstUserID {
			t.Fatalf("blocked user %d remains in profile viewers", firstUserID)
		}
	}
}
