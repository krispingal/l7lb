package ratelimiting

import (
	"time"
)

type NoOpRateLimiter struct{}

func (d NoOpRateLimiter) IsAllowed(ip string) bool { return true }

func (d NoOpRateLimiter) GetState() map[string]int { return map[string]int{} }

func (d NoOpRateLimiter) GetRateLimit() (int, time.Duration) { return 0, 0 }
