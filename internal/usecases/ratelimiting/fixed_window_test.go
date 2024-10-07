package ratelimiting

import (
	"testing"
	"time"
)

func TestFixedWindowRateLimiter_allow(t *testing.T) {
	rl := NewFixedWindowRateLimiter(2, time.Second*1)

	ip := "192.168.1.1"
	if !rl.IsAllowed(ip) {
		t.Errorf("Expected IsAllow to return true for the first request, but got false")
	}
	if !rl.IsAllowed(ip) {
		t.Errorf("Expected IsAllow to return true for the second request, but got false")
	}
	if rl.IsAllowed(ip) {
		t.Errorf("Expected IsAllow to return false for the third request, but got true")
	}
}

func TestFixedWindowRateLimiter_Reset(t *testing.T) {
	rl := NewFixedWindowRateLimiter(2, time.Millisecond*500)

	ip := "192.168.1.1"

	if !rl.IsAllowed(ip) || !rl.IsAllowed(ip) || rl.IsAllowed(ip) {
		t.Errorf("Expected IsAllow behavior to match request limit before reset")
	}

	time.Sleep(time.Millisecond * 600)

	if !rl.IsAllowed(ip) {
		t.Errorf("Expected IsAllow to return true after reset, but got false")
	}
	if !rl.IsAllowed(ip) {
		t.Errorf("Expected IsAllow to return true after reset, but got false")
	}
}
func TestFixedWindowRateLimiter_GetState(t *testing.T) {
	rl := NewFixedWindowRateLimiter(3, time.Second*1)

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	// Make requests for both IPs
	rl.IsAllowed(ip1)
	rl.IsAllowed(ip1)
	rl.IsAllowed(ip2)

	state := rl.GetState()

	// Assert the request counts
	if state[ip1] != 2 {
		t.Errorf("Expected IP1 to have 2 requests, but got %d", state[ip1])
	}

	if state[ip2] != 1 {
		t.Errorf("Expected IP2 to have 1 request, but got %d", state[ip2])
	}
}

func TestFixedWindowRateLimiter_GetRateLimit(t *testing.T) {
	requestLimit := 5
	windowDuration := time.Second * 10
	rl := NewFixedWindowRateLimiter(requestLimit, windowDuration)

	limit, duration := rl.GetRateLimit()

	// Assert that the returned values match the ones provided at initialization
	if limit != requestLimit {
		t.Errorf("Expected request limit to be %d, but got %d", requestLimit, limit)
	}

	if duration != windowDuration {
		t.Errorf("Expected window duration to be %v, but got %v", windowDuration, duration)
	}
}
