package session

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type Penalty struct {
	Failures int
	Lockout  time.Duration
}

type Options struct {
	CookieName             string
	TTL                    time.Duration
	SameSite               http.SameSite
	Secure                 bool
	MaxFailures            int
	Lockout                time.Duration
	Window                 time.Duration
	Penalties              []Penalty
	AccountLockoutFailures int
}

type Manager struct {
	cookieName string
	ttl        time.Duration
	sameSite   http.SameSite
	secure     bool
	mu         sync.RWMutex
	sessions   map[string]Session
}

type LoginLimiter struct {
	penalties []Penalty
	window    time.Duration
	now       func() time.Time
	mu        sync.Mutex
	failures  map[string]failureState
}

type Limiter interface {
	Locked(key string) bool
	Fail(key string)
	Success(key string)
}

type failureState struct {
	count         int
	firstFailedAt time.Time
	lockedUntil   time.Time
}

func DefaultPenaltySchedule() []Penalty {
	return []Penalty{
		{Failures: 4, Lockout: 30 * time.Second},
		{Failures: 6, Lockout: 5 * time.Minute},
		{Failures: 11, Lockout: time.Hour},
		{Failures: 21, Lockout: 24 * time.Hour},
	}
}

func AdminPenaltySchedule() []Penalty {
	return []Penalty{
		{Failures: 2, Lockout: 30 * time.Second},
		{Failures: 3, Lockout: 5 * time.Minute},
		{Failures: 6, Lockout: time.Hour},
		{Failures: 11, Lockout: 24 * time.Hour},
	}
}

func NewLoginLimiter(options Options) *LoginLimiter {
	window := options.Window
	if window == 0 {
		window = time.Hour
	}
	return &LoginLimiter{
		penalties: normalizePenaltySchedule(options),
		window:    window,
		now:       time.Now,
		failures:  make(map[string]failureState),
	}
}

func (l *LoginLimiter) Locked(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	state, ok := l.failures[key]
	if !ok {
		return false
	}
	now := l.now().UTC()
	if state.lockedUntil.IsZero() {
		return false
	}
	if now.Before(state.lockedUntil) {
		return true
	}
	state.lockedUntil = time.Time{}
	if state.firstFailedAt.IsZero() || now.Sub(state.firstFailedAt) > l.window {
		delete(l.failures, key)
	} else {
		l.failures[key] = state
	}
	return false
}

func (l *LoginLimiter) Fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now().UTC()
	state := l.failures[key]
	if state.firstFailedAt.IsZero() || now.Sub(state.firstFailedAt) > l.window {
		state = failureState{firstFailedAt: now}
	}
	state.count++
	if lockout := LockoutForFailureCount(state.count, l.penalties); lockout > 0 {
		state.lockedUntil = now.Add(lockout)
	} else {
		state.lockedUntil = time.Time{}
	}
	l.failures[key] = state
}

func (l *LoginLimiter) Success(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, key)
}

func LockoutForFailureCount(failures int, penalties []Penalty) time.Duration {
	if failures <= 0 {
		return 0
	}
	var lockout time.Duration
	for _, penalty := range normalizePenaltyList(penalties) {
		if failures >= penalty.Failures {
			lockout = penalty.Lockout
		}
	}
	return lockout
}

func normalizePenaltySchedule(options Options) []Penalty {
	if len(options.Penalties) > 0 {
		return normalizePenaltyList(options.Penalties)
	}
	if options.MaxFailures > 0 || options.Lockout > 0 {
		maxFailures := options.MaxFailures
		if maxFailures == 0 {
			maxFailures = 5
		}
		lockout := options.Lockout
		if lockout == 0 {
			lockout = 15 * time.Minute
		}
		return []Penalty{{Failures: maxFailures, Lockout: lockout}}
	}
	return DefaultPenaltySchedule()
}

func normalizePenaltyList(penalties []Penalty) []Penalty {
	normalized := make([]Penalty, 0, len(penalties))
	for _, penalty := range penalties {
		if penalty.Failures <= 0 || penalty.Lockout <= 0 {
			continue
		}
		normalized = append(normalized, penalty)
	}
	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i].Failures < normalized[j].Failures
	})
	return normalized
}

func AnyLocked(limiter Limiter, keys []string) bool {
	if limiter == nil {
		return false
	}
	for _, key := range keys {
		if limiter.Locked(key) {
			return true
		}
	}
	return false
}

func FailAll(limiter Limiter, keys []string) {
	if limiter == nil {
		return
	}
	for _, key := range keys {
		limiter.Fail(key)
	}
}

func SuccessAll(limiter Limiter, keys []string) {
	if limiter == nil {
		return
	}
	for _, key := range keys {
		limiter.Success(key)
	}
}

type Session struct {
	Subject         string
	Kind            string
	CSRFToken       string
	csrfHash        string
	fingerprintHash string
	expiresAt       time.Time
	stepUpUntil     time.Time
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
	sameSite := options.SameSite
	if sameSite == 0 {
		sameSite = http.SameSiteStrictMode
	}
	return &Manager{
		cookieName: cookieName,
		ttl:        ttl,
		sameSite:   sameSite,
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
			SameSite: m.sameSite,
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
	if !ok {
		return Session{}, false
	}
	if time.Now().After(session.expiresAt) {
		m.mu.Lock()
		delete(m.sessions, hash(cookie.Value))
		m.mu.Unlock()
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
	http.SetCookie(w, &http.Cookie{Name: m.cookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: m.sameSite, Secure: m.secure})
}

func (m *Manager) InvalidateSubject(subject, kind string) int {
	if m == nil {
		return 0
	}
	subject = strings.TrimSpace(subject)
	kind = strings.TrimSpace(kind)
	m.mu.Lock()
	defer m.mu.Unlock()
	removed := 0
	for key, session := range m.sessions {
		if strings.EqualFold(session.Subject, subject) && (kind == "" || session.Kind == kind) {
			delete(m.sessions, key)
			removed++
		}
	}
	return removed
}

func (m *Manager) MarkStepUp(r *http.Request, ttl time.Duration) bool {
	if m == nil || ttl <= 0 {
		return false
	}
	if _, ok := m.ValidateUnsafe(r); !ok {
		return false
	}
	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		return false
	}
	key := hash(cookie.Value)
	m.mu.Lock()
	defer m.mu.Unlock()
	current, ok := m.sessions[key]
	if !ok {
		return false
	}
	current.stepUpUntil = time.Now().UTC().Add(ttl)
	m.sessions[key] = current
	return true
}

func (m *Manager) HasRecentStepUp(r *http.Request) bool {
	if m == nil {
		return false
	}
	session, ok := m.Validate(r)
	if !ok {
		return false
	}
	return time.Now().UTC().Before(session.stepUpUntil)
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
