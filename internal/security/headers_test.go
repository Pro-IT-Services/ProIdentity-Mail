package security

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBrowserHeaders(t *testing.T) {
	handler := BrowserHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	for name, want := range map[string]string{
		"X-Content-Type-Options":     "nosniff",
		"X-Frame-Options":            "DENY",
		"Referrer-Policy":            "same-origin",
		"Cross-Origin-Opener-Policy": "same-origin",
	} {
		if got := rec.Header().Get(name); got != want {
			t.Fatalf("%s = %q, want %q", name, got, want)
		}
	}
	if got := rec.Header().Get("Content-Security-Policy"); got == "" {
		t.Fatal("Content-Security-Policy is empty")
	}
	if got := rec.Header().Get("Content-Security-Policy"); !strings.Contains(got, "object-src 'none'") || strings.Contains(got, "unsafe-eval") {
		t.Fatalf("Content-Security-Policy is not hardened enough: %q", got)
	}
	if got := rec.Header().Get("Permissions-Policy"); got == "" {
		t.Fatal("Permissions-Policy is empty")
	}
}

func TestBrowserCSPWithNonceAvoidsUnsafeInlineScripts(t *testing.T) {
	csp := BrowserCSP("abc123")
	for _, want := range []string{
		"default-src 'self'",
		"script-src 'self' 'nonce-abc123'",
		"style-src 'self' 'nonce-abc123' https://fonts.googleapis.com",
		"object-src 'none'",
		"frame-ancestors 'none'",
	} {
		if !strings.Contains(csp, want) {
			t.Fatalf("CSP missing %q: %s", want, csp)
		}
	}
	if strings.Contains(csp, "unsafe-inline") || strings.Contains(csp, "unsafe-eval") {
		t.Fatalf("CSP should not allow unsafe inline/eval: %s", csp)
	}
}

func TestLimitRequestBody(t *testing.T) {
	handler := LimitRequestBody(4)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err == nil {
			t.Fatal("expected body too large error")
		}
		var maxBytesError *http.MaxBytesError
		if !errors.As(err, &maxBytesError) {
			t.Fatalf("error = %T %v, want MaxBytesError", err, err)
		}
		w.WriteHeader(http.StatusRequestEntityTooLarge)
	}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("too large")))

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}
