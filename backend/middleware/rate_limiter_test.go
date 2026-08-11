package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		realIP     string
		want       string
	}{
		{
			name:       "uses valid reverse proxy IP",
			remoteAddr: "172.20.0.5:41000",
			realIP:     "203.0.113.10",
			want:       "203.0.113.10",
		},
		{
			name:       "removes port from remote address",
			remoteAddr: "192.0.2.15:52000",
			want:       "192.0.2.15",
		},
		{
			name:       "ignores invalid reverse proxy IP",
			remoteAddr: "192.0.2.16:53000",
			realIP:     "not-an-ip",
			want:       "192.0.2.16",
		},
		{
			name:       "keeps malformed remote address",
			remoteAddr: "unknown-client",
			want:       "unknown-client",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.RemoteAddr = test.remoteAddr
			if test.realIP != "" {
				request.Header.Set("X-Real-IP", test.realIP)
			}

			if got := clientIP(request); got != test.want {
				t.Fatalf("clientIP() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestRateLimiterLimit(t *testing.T) {
	limiter := NewRateLimiter(2, time.Minute)
	handlerCalls := 0
	handler := limiter.Limit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		handlerCalls++
		w.WriteHeader(http.StatusNoContent)
	}))

	for attempt, remoteAddr := range []string{
		"192.0.2.20:40001",
		"192.0.2.20:40002",
	} {
		request := httptest.NewRequest(http.MethodPost, "/login", nil)
		request.RemoteAddr = remoteAddr
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		if response.Code != http.StatusNoContent {
			t.Fatalf(
				"attempt %d status = %d, want %d",
				attempt+1,
				response.Code,
				http.StatusNoContent,
			)
		}
	}

	request := httptest.NewRequest(http.MethodPost, "/login", nil)
	request.RemoteAddr = "192.0.2.20:40003"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"limited status = %d, want %d",
			response.Code,
			http.StatusTooManyRequests,
		)
	}
	if response.Header().Get("Retry-After") == "" {
		t.Fatal("Retry-After header is missing")
	}
	if handlerCalls != 2 {
		t.Fatalf("handler calls = %d, want 2", handlerCalls)
	}

	otherClientRequest := httptest.NewRequest(http.MethodPost, "/login", nil)
	otherClientRequest.RemoteAddr = "192.0.2.21:40001"
	otherClientResponse := httptest.NewRecorder()
	handler.ServeHTTP(otherClientResponse, otherClientRequest)

	if otherClientResponse.Code != http.StatusNoContent {
		t.Fatalf(
			"other client status = %d, want %d",
			otherClientResponse.Code,
			http.StatusNoContent,
		)
	}
}

func TestRateLimiterWindowResets(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)
	now := time.Now()

	allowed, _ := limiter.allow("192.0.2.30", now)
	if !allowed {
		t.Fatal("first request was rejected")
	}

	allowed, retryAfter := limiter.allow("192.0.2.30", now.Add(30*time.Second))
	if allowed {
		t.Fatal("request inside the same window was allowed")
	}
	if retryAfter != 30*time.Second {
		t.Fatalf("retryAfter = %s, want %s", retryAfter, 30*time.Second)
	}

	allowed, _ = limiter.allow("192.0.2.30", now.Add(time.Minute))
	if !allowed {
		t.Fatal("request after window reset was rejected")
	}
}

func TestNewRateLimiterRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		limit  int
		window time.Duration
	}{
		{
			name:   "zero limit",
			limit:  0,
			window: time.Minute,
		},
		{
			name:   "zero window",
			limit:  1,
			window: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("NewRateLimiter() did not panic")
				}
			}()

			NewRateLimiter(test.limit, test.window)
		})
	}
}
