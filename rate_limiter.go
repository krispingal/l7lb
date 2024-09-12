package main

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	requests   map[string]int
	limit      int
	window     time.Duration
	mu         sync.Mutex
	resetTimer *time.Ticker
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests:   make(map[string]int),
		limit:      limit,
		window:     window,
		resetTimer: time.NewTicker(window),
	}
	go rl.reset()

	return rl
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.requests[ip] >= rl.limit {
		return false
	}
	rl.requests[ip]++
	return true
}

func (rl *RateLimiter) reset() {
	for range rl.resetTimer.C {
		rl.mu.Lock()
		rl.requests = make(map[string]int)
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) getClientIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("Could not parse client IP from RemoteAddr: %v", err)
		return ""
	}
	return ip
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := rl.getClientIP(r)
		if clientIP == "" {
			http.Error(w, "Could not determine client IP", http.StatusInternalServerError)
			return
		}
		if !rl.Allow(clientIP) {
			log.Printf("Rate limit exceeded for %s", r.URL.Path)
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
