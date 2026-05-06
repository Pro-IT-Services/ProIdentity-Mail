package session

import (
	"testing"
	"time"
)

func TestLoginLimiterLocksAfterFailures(t *testing.T) {
	limiter := NewLoginLimiter(Options{MaxFailures: 3, Lockout: time.Minute})
	key := "admin|127.0.0"

	if limiter.Locked(key) {
		t.Fatal("new key should not be locked")
	}
	limiter.Fail(key)
	limiter.Fail(key)
	if limiter.Locked(key) {
		t.Fatal("key locked too early")
	}
	limiter.Fail(key)
	if !limiter.Locked(key) {
		t.Fatal("key not locked after max failures")
	}
	limiter.Success(key)
	if limiter.Locked(key) {
		t.Fatal("success did not clear lockout")
	}
}
