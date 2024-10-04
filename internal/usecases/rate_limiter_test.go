package usecases

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
