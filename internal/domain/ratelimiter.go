package domain

import "time"

type RateLimiter interface {
	IsAllowed(ip string) bool           // Allow or reject a request based on IP
	GetState() map[string]int           // Get the current state of the rate limiter
	GetRateLimit() (int, time.Duration) // Returns the request limit and the time window
}
