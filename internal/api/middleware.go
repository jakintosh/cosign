package api

import (
	"cosign/internal/service"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

// Rate limiter map: key is IP address
var (
	rateLimiters = make(map[string]*rate.Limiter)
	rateLimitersMu sync.Mutex
)

// getRateLimiter returns a rate limiter for the given IP
func getRateLimiter(ip string) *rate.Limiter {
	rateLimitersMu.Lock()
	defer rateLimitersMu.Unlock()

	limiter, exists := rateLimiters[ip]
	if !exists {
		// Create new limiter: 10 requests per second, burst of 20
		limiter = rate.NewLimiter(10, 20)
		rateLimiters[ip] = limiter
	}

	return limiter
}

// withRateLimit applies rate limiting based on client IP
func withRateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract IP from request
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			// Use first IP in X-Forwarded-For
			ips := strings.Split(forwarded, ",")
			ip = strings.TrimSpace(ips[0])
		}

		limiter := getRateLimiter(ip)
		if !limiter.Allow() {
			writeError(w, http.StatusTooManyRequests, "Rate limit exceeded")
			return
		}

		next(w, r)
	}
}

// withAuth checks for valid API key in Authorization header
func withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := extractBearerToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "Missing authorization token")
			return
		}

		ok, err := service.VerifyAPIKey(token)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to verify token")
			return
		}

		if !ok {
			writeError(w, http.StatusUnauthorized, "Invalid authorization token")
			return
		}

		next(w, r)
	}
}

// withCORS validates Origin header against whitelist and sets CORS headers
func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		allowed, err := service.IsOriginAllowed(origin)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to verify origin")
			return
		}

		if !allowed && origin != "" {
			writeError(w, http.StatusForbidden, "Origin not allowed")
			return
		}

		// Set CORS headers
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

// extractBearerToken extracts the token from Authorization: Bearer {token} header
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}
