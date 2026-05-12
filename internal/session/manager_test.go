package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestManagerCreatesAndValidatesSessionWithFingerprint(t *testing.T) {
	manager := NewManager(Options{CookieName: "sid"})
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("User-Agent", "Browser A")
	req.Header.Set("Accept-Language", "en-US")

	created, err := manager.Create(req, "admin", "admin")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	req.AddCookie(created.Cookie)
	req.Header.Set("X-CSRF-Token", created.CSRFToken)

	session, ok := manager.Validate(req)
	if !ok {
		t.Fatal("Validate rejected created session")
	}
	if session.Subject != "admin" || session.Kind != "admin" {
		t.Fatalf("unexpected session: %+v", session)
	}
	if created.Cookie.SameSite != http.SameSiteStrictMode || !created.Cookie.HttpOnly {
		t.Fatalf("cookie flags = SameSite:%v HttpOnly:%t, want Strict and HttpOnly", created.Cookie.SameSite, created.Cookie.HttpOnly)
	}

	req.Header.Set("User-Agent", "Other Browser")
	if _, ok := manager.Validate(req); ok {
		t.Fatal("Validate accepted changed browser fingerprint")
	}
}

func TestManagerRequiresCSRFForUnsafeRequest(t *testing.T) {
	manager := NewManager(Options{CookieName: "sid"})
	req := httptest.NewRequest("GET", "/", nil)
	created, err := manager.Create(req, "user@example.com", "webmail")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	post := httptest.NewRequest("POST", "/api/v1/send", nil)
	post.AddCookie(created.Cookie)
	if _, ok := manager.ValidateUnsafe(post); ok {
		t.Fatal("ValidateUnsafe accepted missing CSRF token")
	}
	post.Header.Set("X-CSRF-Token", created.CSRFToken)
	if _, ok := manager.ValidateUnsafe(post); !ok {
		t.Fatal("ValidateUnsafe rejected matching CSRF token")
	}
}

func TestManagerInvalidatesSubjectSessions(t *testing.T) {
	manager := NewManager(Options{CookieName: "sid"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	createdUser, err := manager.Create(req, "User@Example.com", "webmail")
	if err != nil {
		t.Fatalf("Create user session: %v", err)
	}
	createdAdmin, err := manager.Create(req, "user@example.com", "admin")
	if err != nil {
		t.Fatalf("Create admin session: %v", err)
	}

	if removed := manager.InvalidateSubject("user@example.com", "webmail"); removed != 1 {
		t.Fatalf("removed = %d, want 1", removed)
	}
	userReq := httptest.NewRequest(http.MethodGet, "/", nil)
	userReq.AddCookie(createdUser.Cookie)
	if _, ok := manager.Validate(userReq); ok {
		t.Fatal("webmail session still valid after invalidation")
	}
	adminReq := httptest.NewRequest(http.MethodGet, "/", nil)
	adminReq.AddCookie(createdAdmin.Cookie)
	if _, ok := manager.Validate(adminReq); !ok {
		t.Fatal("admin session with same subject should remain valid")
	}
}

func TestManagerMarksRecentStepUpOnlyWithCSRF(t *testing.T) {
	manager := NewManager(Options{CookieName: "sid"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Browser A")
	created, err := manager.Create(req, "admin", "admin")
	if err != nil {
		t.Fatalf("Create session: %v", err)
	}
	step := httptest.NewRequest(http.MethodPost, "/api/v1/session/step-up/verify", nil)
	step.Header.Set("User-Agent", "Browser A")
	step.AddCookie(created.Cookie)
	if manager.MarkStepUp(step, 5*time.Minute) {
		t.Fatal("MarkStepUp accepted request without CSRF token")
	}
	step.Header.Set("X-CSRF-Token", created.CSRFToken)
	if !manager.MarkStepUp(step, 5*time.Minute) {
		t.Fatal("MarkStepUp rejected valid session and CSRF token")
	}
	check := httptest.NewRequest(http.MethodGet, "/api/v1/domains", nil)
	check.Header.Set("User-Agent", "Browser A")
	check.AddCookie(created.Cookie)
	if !manager.HasRecentStepUp(check) {
		t.Fatal("HasRecentStepUp rejected freshly marked session")
	}
}
