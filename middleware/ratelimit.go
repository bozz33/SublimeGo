package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bozz33/sublimego/auth"
	"golang.org/x/time/rate"
)

// KeyFunc extracts a unique key from a request to identify the client.
type KeyFunc func(r *http.Request) string

// RateLimitConfig configures the rate limiter.
type RateLimitConfig struct {
	RequestsPerMinute int
	Burst             int
	KeyFunc           KeyFunc
	WhitelistIPs      []string
	CleanupInterval   time.Duration
	OnLimitExceeded   func(r *http.Request, key string)
}

// RateLimiter manages rate limiting using the Token Bucket algorithm.
type RateLimiter struct {
	config    *RateLimitConfig
	limiters  sync.Map // map[string]*limiterEntry
	whitelist map[string]bool
	mu        sync.RWMutex
	stopClean chan struct{}
}

// limiterEntry contains a rate limiter and its last access time.
type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(config *RateLimitConfig) *RateLimiter {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	if config.KeyFunc == nil {
		config.KeyFunc = KeyByIP
	}

	if config.CleanupInterval == 0 {
		config.CleanupInterval = 5 * time.Minute
	}

	rl := &RateLimiter{
		config:    config,
		whitelist: make(map[string]bool),
		stopClean: make(chan struct{}),
	}

	for _, ip := range config.WhitelistIPs {
		rl.whitelist[ip] = true
	}

	go rl.cleanupLoop()

	return rl
}

// DefaultRateLimitConfig returns a default configuration.
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		RequestsPerMinute: 60,
		Burst:             10,
		KeyFunc:           KeyByIP,
		CleanupInterval:   5 * time.Minute,
	}
}

// Middleware returns the rate limiting middleware.
func (rl *RateLimiter) Middleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := rl.config.KeyFunc(r)

			if rl.isWhitelisted(key, r) {
				next.ServeHTTP(w, r)
				return
			}

			limiter := rl.getLimiter(key)

			if !limiter.Allow() {
				if rl.config.OnLimitExceeded != nil {
					rl.config.OnLimitExceeded(r, key)
				}

				rl.handleRateLimitExceeded(w, r, key)
				return
			}

			rl.setRateLimitHeaders(w, limiter)
			next.ServeHTTP(w, r)
		})
	}
}

// getLimiter retrieves or creates a limiter for a given key.
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	if entry, ok := rl.limiters.Load(key); ok {
		if e, ok2 := entry.(*limiterEntry); ok2 {
			e.lastSeen = time.Now()
			return e.limiter
		}
	}

	limit := rate.Limit(float64(rl.config.RequestsPerMinute) / 60.0)
	limiter := rate.NewLimiter(limit, rl.config.Burst)

	entry := &limiterEntry{
		limiter:  limiter,
		lastSeen: time.Now(),
	}

	rl.limiters.Store(key, entry)
	return limiter
}

// isWhitelisted checks if a key or IP is in the whitelist.
func (rl *RateLimiter) isWhitelisted(key string, r *http.Request) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	// Check if the key itself is whitelisted
	if rl.whitelist[key] {
		return true
	}

	// Extract the client's IP address
	clientIP := getClientIPFromRequest(r)

	// Check if the IP address is whitelisted
	if rl.whitelist[clientIP] {
		return true
	}

	// Check if the IP address is in a whitelisted CIDR range
	for whitelistIP := range rl.whitelist {
		if strings.Contains(whitelistIP, "/") {
			if isIPInCIDR(clientIP, whitelistIP) {
				return true
			}
		}
	}

	return false
}

// setRateLimitHeaders adds informative rate limiting headers.
func (rl *RateLimiter) setRateLimitHeaders(w http.ResponseWriter, limiter *rate.Limiter) {
	w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.config.RequestsPerMinute))

	tokens := int(limiter.Tokens())
	if tokens < 0 {
		tokens = 0
	}
	w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", tokens))

	resetTime := time.Now().Add(time.Minute).Unix()
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))
}

// handleRateLimitExceeded handles the case when the limit is exceeded.
func (rl *RateLimiter) handleRateLimitExceeded(w http.ResponseWriter, r *http.Request, key string) {
	retryAfter := 60
	w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.config.RequestsPerMinute))
	w.Header().Set("X-RateLimit-Remaining", "0")
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
	w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusTooManyRequests)
	_, _ = fmt.Fprintf(w, `{"error":"Too many requests. Retry in %d seconds."}`, retryAfter)
}

// cleanupLoop periodically cleans up inactive limiters.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopClean:
			return
		}
	}
}

// cleanup removes limiters inactive for more than 2x CleanupInterval.
func (rl *RateLimiter) cleanup() {
	threshold := time.Now().Add(-2 * rl.config.CleanupInterval)

	rl.limiters.Range(func(key, value interface{}) bool {
		entry, ok := value.(*limiterEntry)
		if !ok {
			return true
		}
		if entry.lastSeen.Before(threshold) {
			rl.limiters.Delete(key)
		}
		return true
	})
}

// Stop stops the cleanup loop.
func (rl *RateLimiter) Stop() {
	close(rl.stopClean)
}

// KeyByIP extracts the client IP.
func KeyByIP(r *http.Request) string {
	return getClientIPFromRequest(r)
}

// KeyByUser extracts the authenticated user ID.
func KeyByUser(r *http.Request) string {
	user := auth.CurrentUser(r)
	if user != nil {
		return fmt.Sprintf("user:%d", user.ID)
	}
	return KeyByIP(r)
}

// KeyByHeader returns a KeyFunc that extracts a header value.
func KeyByHeader(header string) KeyFunc {
	return func(r *http.Request) string {
		value := r.Header.Get(header)
		if value == "" {
			return KeyByIP(r)
		}
		return value
	}
}

// getClientIPFromRequest extracts the client IP considering proxies.
func getClientIPFromRequest(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// isIPInCIDR checks if an IP is in a CIDR range.
func isIPInCIDR(ipStr, cidr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	return ipNet.Contains(ip)
}

// WithRateLimitConfig creates a rate limiter with custom config.
func WithRateLimitConfig(requestsPerMinute, burst int, keyFunc KeyFunc) *RateLimiter {
	return NewRateLimiter(&RateLimitConfig{
		RequestsPerMinute: requestsPerMinute,
		Burst:             burst,
		KeyFunc:           keyFunc,
	})
}
