package httphandler

import (
	"log"
	"net"
	"net/http"

	"github.com/krispingal/l7lb/internal/usecases/ratelimiting"
)

func getClientIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return ""
	}
	return ip
}

func NewMiddleware(limiter ratelimiting.RateLimiterInterface, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		if clientIP == "" {
			log.Printf("Could not determine client IP from req %v", r.Header)
			http.Error(w, "Could not determine client IP", http.StatusInternalServerError)
			return
		}
		if !limiter.IsAllowed(clientIP) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
