package session

import (
	"net/http/httptest"
	"testing"
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
