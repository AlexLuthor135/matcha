package middleware

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type clientWindow struct {
	requests int
	resetAt  time.Time
}

type RateLimiter struct {
	mutex       sync.Mutex
	clients     map[string]clientWindow
	limit       int
	window      time.Duration
	lastCleanup time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit < 1 {
		panic("rate limiter limit must be positive")
	}
	if window <= 0 {
		panic("rate limiter window must be positive")
	}
	return &RateLimiter{
		clients:     make(map[string]clientWindow),
		limit:       limit,
		window:      window,
		lastCleanup: time.Now(),
	}
}

func (limiter *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		allowed, retryAfter := limiter.allow(ip, time.Now())
		if !allowed {
			retryAfterSeconds := int(math.Ceil(retryAfter.Seconds()))
			if retryAfterSeconds < 1 {
				retryAfterSeconds = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (limiter *RateLimiter) allow(clientID string, now time.Time) (bool, time.Duration) {
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()
	limiter.cleanup(now)
	window, exists := limiter.clients[clientID]
	if !exists || !now.Before(window.resetAt) {
		limiter.clients[clientID] = clientWindow{
			requests: 1,
			resetAt:  now.Add(limiter.window),
		}
		return true, 0
	}
	if window.requests >= limiter.limit {
		return false, window.resetAt.Sub(now)
	}
	window.requests++
	limiter.clients[clientID] = window
	return true, 0
}

func (limiter *RateLimiter) cleanup(now time.Time) {
	if now.Sub(limiter.lastCleanup) < limiter.window {
		return
	}
	for clientID, window := range limiter.clients {
		if !now.Before(window.resetAt) {
			delete(limiter.clients, clientID)
		}
	}
	limiter.lastCleanup = now
}

func clientIP(r *http.Request) string {
	if forwardedIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); forwardedIP != "" {
		if parsedIP := net.ParseIP(forwardedIP); parsedIP != nil {
			return parsedIP.String()
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
