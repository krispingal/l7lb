package httphandler

import (
	"net"
	"net/http"

	"github.com/krispingal/l7lb/internal/domain"
	"go.uber.org/zap"
)

func getClientIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return ""
	}
	return ip
}

func NewMiddleware(limiter domain.RateLimiter, next http.Handler, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		if clientIP == "" {
			logger.Error("Could not determine client IP from req", zap.Any("request_header", r.Header))
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
