package httphandler

import (
	"net/http"

	"github.com/krispingal/l7lb/internal/usecases"
)

func NewMiddleware(limiter *usecases.RateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := limiter.GetClientIP(r)
		if clientIP == "" {
			http.Error(w, "Could not determine client IP", http.StatusInternalServerError)
			return
		}
		if !limiter.Allow(clientIP) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
