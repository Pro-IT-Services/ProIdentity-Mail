package session

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Options struct {
	CookieName  string
	TTL         time.Duration
	Secure      bool
	MaxFailures int
	Lockout     time.Duration
}

type Manager struct {
	cookieName string
	ttl        time.Duration
	secure     bool
	mu         sync.RWMutex
	sessions   map[string]Session
}

type LoginLimiter struct {
	maxFailures int
	lockout     time.Duration
	mu          sync.Mutex
	failures    map[string]failureState
}

type failureState struct {
	count    int
	lockedAt time.Time
}

func NewLoginLimiter(options Options) *LoginLimiter {
	maxFailures := options.MaxFailures
	if maxFailures == 0 {
		maxFailures = 5
	}
	lockout := options.Lockout
	if lockout == 0 {
		lockout = 15 * time.Minute
	}
	return &LoginLimiter{maxFailures: maxFailures, lockout: lockout, failures: make(map[string]failureState)}
}

func (l *LoginLimiter) Locked(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	state, ok := l.failures[key]
	if !ok || state.count < l.maxFailures {
		return false
	}
	if time.Since(state.lockedAt) > l.lockout {
		delete(l.failures, key)
		return false
	}
	return true
}

func (l *LoginLimiter) Fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	state := l.failures[key]
	state.count++
	if state.count >= l.maxFailures && state.lockedAt.IsZero() {
		state.lockedAt = time.Now()
	}
	l.failures[key] = state
}

func (l *LoginLimiter) Success(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, key)
}

type Session struct {
	Subject         string
	Kind            string
	CSRFToken       string
	csrfHash        string
	fingerprintHash string
	expiresAt       time.Time
}

type Created struct {
	Cookie    *http.Cookie
	CSRFToken string
}

func NewManager(options Options) *Manager {
	cookieName := options.CookieName
	if cookieName == "" {
		cookieName = "proidentity_session"
	}
	ttl := options.TTL
	if ttl == 0 {
		ttl = 8 * time.Hour
	}
	return &Manager{
		cookieName: cookieName,
		ttl:        ttl,
		secure:     options.Secure,
		sessions:   make(map[string]Session),
	}
}

func (m *Manager) Create(r *http.Request, subject, kind string) (Created, error) {
	token, err := randomToken(32)
	if err != nil {
		return Created{}, err
	}
	csrf, err := randomToken(32)
	if err != nil {
		return Created{}, err
	}
	now := time.Now()
	session := Session{
		Subject:         subject,
		Kind:            kind,
		CSRFToken:       csrf,
		csrfHash:        hash(csrf),
		fingerprintHash: browserFingerprint(r),
		expiresAt:       now.Add(m.ttl),
	}
	m.mu.Lock()
	m.sessions[hash(token)] = session
	m.mu.Unlock()
	return Created{
		Cookie: &http.Cookie{
			Name:     m.cookieName,
			Value:    token,
			Path:     "/",
			Expires:  session.expiresAt,
			MaxAge:   int(m.ttl.Seconds()),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   m.secure,
		},
		CSRFToken: csrf,
	}, nil
}

func (m *Manager) Validate(r *http.Request) (Session, bool) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil || cookie.Value == "" {
		return Session{}, false
	}
	m.mu.RLock()
	session, ok := m.sessions[hash(cookie.Value)]
	m.mu.RUnlock()
	if !ok || time.Now().After(session.expiresAt) {
		return Session{}, false
	}
	if subtle.ConstantTimeCompare([]byte(session.fingerprintHash), []byte(browserFingerprint(r))) != 1 {
		return Session{}, false
	}
	return session, true
}

func (m *Manager) ValidateUnsafe(r *http.Request) (Session, bool) {
	session, ok := m.Validate(r)
	if !ok {
		return Session{}, false
	}
	token := r.Header.Get("X-CSRF-Token")
	if token == "" {
		token = r.Header.Get("X-ProIdentity-CSRF")
	}
	if token == "" || subtle.ConstantTimeCompare([]byte(session.csrfHash), []byte(hash(token))) != 1 {
		return Session{}, false
	}
	return session, true
}

func (m *Manager) Clear(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(m.cookieName); err == nil {
		m.mu.Lock()
		delete(m.sessions, hash(cookie.Value))
		m.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: m.cookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: m.secure})
}

func browserFingerprint(r *http.Request) string {
	source := strings.Join([]string{
		r.Header.Get("User-Agent"),
		r.Header.Get("Accept-Language"),
		ipPrefix(r.RemoteAddr),
	}, "\n")
	return hash(source)
}

func ipPrefix(remoteAddr string) string {
	host := remoteAddr
	if idx := strings.LastIndex(host, ":"); idx > 0 {
		host = host[:idx]
	}
	parts := strings.Split(host, ".")
	if len(parts) == 4 {
		return strings.Join(parts[:3], ".")
	}
	return host
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
