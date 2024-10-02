package usecases

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

func (rl *RateLimiter) GetClientIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("Could not parse client IP from RemoteAddr: %v", err)
		return ""
	}
	return ip
}
