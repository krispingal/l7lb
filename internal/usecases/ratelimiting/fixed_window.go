package ratelimiting

import (
	"sync"
	"time"
)

type RateLimiterInterface interface {
	IsAllowed(ip string) bool           // Allow or reject a request based on IP
	GetState() map[string]int           // Get the current state of the rate limiter
	GetRateLimit() (int, time.Duration) // Returns the request limit and the time window
}

type FixedWindowRateLimiter struct {
	ipRequestCount map[string]int
	requestLimit   int
	windowDuration time.Duration
	mu             sync.RWMutex // Mutex for protecting the request map
	resetTicker    *time.Ticker // Ticker to reset requests after each window
}

func NewFixedWindowRateLimiter(requestLimit int, windowDuration time.Duration) *FixedWindowRateLimiter {
	rl := &FixedWindowRateLimiter{
		ipRequestCount: make(map[string]int),
		requestLimit:   requestLimit,
		windowDuration: windowDuration,
		resetTicker:    time.NewTicker(windowDuration),
	}
	go rl.reset() // Background routine to reset counts every window
	return rl
}

func (rl *FixedWindowRateLimiter) IsAllowed(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.ipRequestCount[ip] >= rl.requestLimit {
		return false
	}
	rl.ipRequestCount[ip]++
	return true
}

func (rl *FixedWindowRateLimiter) GetState() map[string]int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	//Create a copy of the current state to avoid modifying the original map
	stateCopy := make(map[string]int, len(rl.ipRequestCount))
	for ip, count := range rl.ipRequestCount {
		stateCopy[ip] = count
	}
	return stateCopy
}

func (rl *FixedWindowRateLimiter) GetRateLimit() (int, time.Duration) {
	return rl.requestLimit, rl.windowDuration
}

// reset periodically clears the rquest counts after the time window
func (rl *FixedWindowRateLimiter) reset() {
	for range rl.resetTicker.C {
		rl.mu.Lock()
		rl.ipRequestCount = make(map[string]int) // Clear the map to reset all counts
		rl.mu.Unlock()
	}
}
