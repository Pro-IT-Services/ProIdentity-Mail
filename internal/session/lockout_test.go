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

func TestDefaultPenaltyScheduleMatchesSecurityPolicy(t *testing.T) {
	tests := []struct {
		failures int
		want     time.Duration
	}{
		{failures: 3, want: 0},
		{failures: 4, want: 30 * time.Second},
		{failures: 5, want: 30 * time.Second},
		{failures: 6, want: 5 * time.Minute},
		{failures: 10, want: 5 * time.Minute},
		{failures: 11, want: time.Hour},
		{failures: 20, want: time.Hour},
		{failures: 21, want: 24 * time.Hour},
	}
	for _, tt := range tests {
		if got := LockoutForFailureCount(tt.failures, DefaultPenaltySchedule()); got != tt.want {
			t.Fatalf("default lockout for %d failures = %s, want %s", tt.failures, got, tt.want)
		}
	}
}

func TestAdminPenaltyScheduleIsStricter(t *testing.T) {
	tests := []struct {
		failures int
		want     time.Duration
	}{
		{failures: 1, want: 0},
		{failures: 2, want: 30 * time.Second},
		{failures: 3, want: 5 * time.Minute},
		{failures: 6, want: time.Hour},
		{failures: 11, want: 24 * time.Hour},
	}
	for _, tt := range tests {
		if got := LockoutForFailureCount(tt.failures, AdminPenaltySchedule()); got != tt.want {
			t.Fatalf("admin lockout for %d failures = %s, want %s", tt.failures, got, tt.want)
		}
	}
}

func TestLoginLimiterDefaultUsesProgressiveSchedule(t *testing.T) {
	limiter := NewLoginLimiter(Options{})
	key := "webmail|ip|198.51.100.10"

	for i := 0; i < 3; i++ {
		limiter.Fail(key)
	}
	if limiter.Locked(key) {
		t.Fatal("key locked before default progressive threshold")
	}
	limiter.Fail(key)
	if !limiter.Locked(key) {
		t.Fatal("key not locked at default progressive threshold")
	}
}

func TestLoginLimiterKeepsFailureCountAfterTemporaryLockExpires(t *testing.T) {
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	limiter := NewLoginLimiter(Options{Window: time.Hour})
	limiter.now = func() time.Time { return now }
	key := "dovecot|pair|tester@example.com|198.51.100.10"

	for i := 0; i < 4; i++ {
		limiter.Fail(key)
	}
	if !limiter.Locked(key) {
		t.Fatal("key should be locked after four failures")
	}
	now = now.Add(31 * time.Second)
	if limiter.Locked(key) {
		t.Fatal("key should be allowed after the 30 second penalty expires")
	}
	limiter.Fail(key)
	limiter.Fail(key)
	if !limiter.Locked(key) {
		t.Fatal("key should relock after escalating to six failures in the same window")
	}
	state := limiter.failures[key]
	if got := state.lockedUntil.Sub(now); got != 5*time.Minute {
		t.Fatalf("escalated lockout = %s, want 5m", got)
	}
}

func TestSQLLoginLimiterUsesSamePenaltySchedule(t *testing.T) {
	limiter := NewSQLLoginLimiter(nil, "webmail", Options{})
	if got := LockoutForFailureCount(4, limiter.penalties); got != 30*time.Second {
		t.Fatalf("sql default four failure lockout = %s, want 30s", got)
	}
	if got := LockoutForFailureCount(21, limiter.penalties); got != 24*time.Hour {
		t.Fatalf("sql default twenty-one failure lockout = %s, want 24h", got)
	}
	if limiter.window != time.Hour {
		t.Fatalf("sql limiter window = %s, want 1h", limiter.window)
	}
	if limiter.accountLockoutFailures != 10 {
		t.Fatalf("sql limiter account lockout threshold = %d, want 10", limiter.accountLockoutFailures)
	}
}

func TestAccountEmailFromLimiterKeyOnlyAcceptsAccountEmailKeys(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{key: "webmail|account|Tester@Example.COM", want: "tester@example.com"},
		{key: "dovecot|account|tester@example.com", want: "tester@example.com"},
		{key: "webmail|pair|tester@example.com|198.51.100.10", want: ""},
		{key: "webmail|ip|198.51.100.10", want: ""},
		{key: "admin|account|admin", want: ""},
		{key: "dovecot|account|not an email", want: ""},
	}
	for _, tt := range tests {
		if got := AccountEmailFromLimiterKey(tt.key); got != tt.want {
			t.Fatalf("AccountEmailFromLimiterKey(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestParsePairLimiterKeyValidatesShapeAndControlCharacters(t *testing.T) {
	pair, ok := parsePairLimiterKey("webmail|pair|Tester@Example.COM|203.0.113.10")
	if !ok {
		t.Fatal("expected valid pair key")
	}
	if pair.Scope != "webmail" || pair.Account != "tester@example.com" || pair.Client != "203.0.113.10" {
		t.Fatalf("pair = %+v", pair)
	}

	invalid := []string{
		"webmail|account|tester@example.com",
		"webmail|pair|tester@example.com",
		"webmail|pair|tester@example.com|203.0.113.10|extra",
		"webmail|pair|tester@example.com\nother|203.0.113.10",
		"webmail|pair||203.0.113.10",
		"webmail|pair|tester@example.com|",
	}
	for _, key := range invalid {
		if _, ok := parsePairLimiterKey(key); ok {
			t.Fatalf("parsePairLimiterKey(%q) unexpectedly succeeded", key)
		}
	}
}
