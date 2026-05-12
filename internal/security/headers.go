package security

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
)

func BrowserHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "DENY")
		header.Set("Referrer-Policy", "same-origin")
		header.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=(), usb=()")
		header.Set("Cross-Origin-Opener-Policy", "same-origin")
		if header.Get("Content-Security-Policy") == "" {
			header.Set("Content-Security-Policy", BrowserCSP(""))
		}
		next.ServeHTTP(w, r)
	})
}

func NewCSPNonce() (string, error) {
	var buf [18]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(buf[:]), nil
}

func BrowserCSP(nonce string) string {
	scriptSrc := "script-src 'self'"
	styleSrc := "style-src 'self' https://fonts.googleapis.com"
	nonce = strings.TrimSpace(nonce)
	if nonce != "" {
		scriptSrc += " 'nonce-" + nonce + "'"
		styleSrc = "style-src 'self' 'nonce-" + nonce + "' https://fonts.googleapis.com"
	}
	return strings.Join([]string{
		"default-src 'self'",
		scriptSrc,
		styleSrc,
		"font-src 'self' https://fonts.gstatic.com data:",
		"img-src 'self' data: blob: https: http:",
		"connect-src 'self'",
		"frame-src 'self' blob: data:",
		"media-src 'self' data: blob:",
		"object-src 'none'",
		"base-uri 'self'",
		"form-action 'self'",
		"frame-ancestors 'none'",
	}, "; ")
}

func LimitRequestBody(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}
