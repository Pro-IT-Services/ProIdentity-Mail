package admin

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"proidentity-mail/internal/domain"
	"proidentity-mail/internal/security"
	"proidentity-mail/internal/session"
)

type Store interface {
	CreateTenant(ctx context.Context, tenant domain.Tenant) (domain.Tenant, error)
	ListTenants(ctx context.Context) ([]domain.Tenant, error)
	CreateDomain(ctx context.Context, mailDomain domain.Domain) (domain.Domain, error)
	ListDomains(ctx context.Context) ([]domain.Domain, error)
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	ListUsers(ctx context.Context) ([]domain.User, error)
	ListQuarantineEvents(ctx context.Context) ([]domain.QuarantineEvent, error)
	ResolveQuarantineEvent(ctx context.Context, eventID uint64, status, note string) (domain.QuarantineEvent, error)
	ListAuditEvents(ctx context.Context) ([]domain.AuditEvent, error)
	RecordAuditEvent(ctx context.Context, event domain.AuditEvent) error
	ListTenantPolicies(ctx context.Context) ([]domain.TenantPolicy, error)
	UpdateTenantPolicy(ctx context.Context, policy domain.TenantPolicy) (domain.TenantPolicy, error)
	GetDomainDNS(ctx context.Context, domainID uint64) (domain.DomainDNS, error)
}

type handler struct {
	store Store
	auth  AuthConfig
}

type AuthConfig struct {
	Username string
	Password string
	Sessions *session.Manager
	Limiter  *session.LoginLimiter
}

func NewRouter(store Store, authConfig ...AuthConfig) http.Handler {
	var auth AuthConfig
	if len(authConfig) > 0 {
		auth = authConfig[0]
	}
	h := handler{store: store, auth: auth}
	r := chi.NewRouter()
	r.Get("/healthz", health)
	r.Get("/.well-known/autoconfig/mail/config-v1.1.xml", h.mailAutoconfig)
	r.Get("/mail/config-v1.1.xml", h.mailAutoconfig)
	r.Get("/.well-known/proidentity-mail/config.json", h.serviceDiscovery)
	r.Get("/.well-known/caldav", wellKnownDAV)
	r.Head("/.well-known/caldav", wellKnownDAV)
	r.Get("/.well-known/carddav", wellKnownDAV)
	r.Head("/.well-known/carddav", wellKnownDAV)
	r.Get("/", h.index)
	r.Get("/api/v1/session", h.login)
	r.Post("/api/v1/session", h.login)
	r.Delete("/api/v1/session", h.logout)
	r.Group(func(protected chi.Router) {
		protected.Use(h.requireAdmin)
		protected.Get("/api/v1/tenants", h.listTenants)
		protected.Post("/api/v1/tenants", h.createTenant)
		protected.Get("/api/v1/domains", h.listDomains)
		protected.Post("/api/v1/domains", h.createDomain)
		protected.Get("/api/v1/domains/{domainID}/dns", h.getDomainDNS)
		protected.Get("/api/v1/users", h.listUsers)
		protected.Post("/api/v1/users", h.createUser)
		protected.Get("/api/v1/quarantine", h.listQuarantineEvents)
		protected.Post("/api/v1/quarantine/{eventID}/release", h.releaseQuarantineEvent)
		protected.Post("/api/v1/quarantine/{eventID}/delete", h.deleteQuarantineEvent)
		protected.Get("/api/v1/audit", h.listAuditEvents)
		protected.Get("/api/v1/policies", h.listTenantPolicies)
		protected.Put("/api/v1/policies/{tenantID}", h.updateTenantPolicy)
	})
	return r
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h handler) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(adminIndexHTML))
}

func (h handler) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.auth.Username == "" && h.auth.Password == "" {
			next.ServeHTTP(w, r)
			return
		}
		if h.auth.Sessions != nil {
			if !safeMethod(r.Method) {
				if _, ok := h.auth.Sessions.ValidateUnsafe(r); ok {
					next.ServeHTTP(w, r)
					return
				}
				if _, _, ok := r.BasicAuth(); !ok {
					http.Error(w, "csrf required", http.StatusForbidden)
					return
				}
			} else if _, ok := h.auth.Sessions.Validate(r); ok {
				next.ServeHTTP(w, r)
				return
			}
			if _, _, ok := r.BasicAuth(); !ok {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
		}
		username, password, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(username), []byte(h.auth.Username)) != 1 || subtle.ConstantTimeCompare([]byte(password), []byte(h.auth.Password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="ProIdentity Admin", charset="UTF-8"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h handler) login(w http.ResponseWriter, r *http.Request) {
	if h.auth.Sessions == nil {
		writeError(w, http.StatusServiceUnavailable, "sessions unavailable")
		return
	}
	if r.Method == http.MethodGet {
		current, ok := h.auth.Sessions.Validate(r)
		if !ok || current.Kind != "admin" {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"csrf_token": current.CSRFToken, "username": current.Subject})
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	key := loginKey(req.Username, r)
	if h.auth.Limiter != nil && h.auth.Limiter.Locked(key) {
		writeError(w, http.StatusTooManyRequests, "login temporarily locked")
		return
	}
	if subtle.ConstantTimeCompare([]byte(req.Username), []byte(h.auth.Username)) != 1 || subtle.ConstantTimeCompare([]byte(req.Password), []byte(h.auth.Password)) != 1 {
		if h.auth.Limiter != nil {
			h.auth.Limiter.Fail(key)
		}
		h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "admin.login_failed", TargetType: "admin", TargetID: req.Username, MetadataJSON: `{}`})
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if h.auth.Limiter != nil {
		h.auth.Limiter.Success(key)
	}
	created, err := h.auth.Sessions.Create(r, req.Username, "admin")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create session failed")
		return
	}
	http.SetCookie(w, created.Cookie)
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "admin.login", TargetType: "admin", TargetID: req.Username, MetadataJSON: `{}`})
	writeJSON(w, http.StatusOK, map[string]string{"csrf_token": created.CSRFToken})
}

func loginKey(subject string, r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	return strings.ToLower(strings.TrimSpace(subject)) + "|" + host
}

func (h handler) logout(w http.ResponseWriter, r *http.Request) {
	if h.auth.Sessions != nil {
		h.auth.Sessions.Clear(w, r)
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "admin.logout", TargetType: "admin_session", TargetID: "current", MetadataJSON: `{}`})
	w.WriteHeader(http.StatusNoContent)
}

func (h handler) recordAudit(ctx context.Context, event domain.AuditEvent) {
	if h.store == nil {
		return
	}
	if event.MetadataJSON == "" {
		event.MetadataJSON = `{}`
	}
	_ = h.store.RecordAuditEvent(ctx, event)
}

func safeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

func (h handler) mailAutoconfig(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("emailaddress")
	at := strings.LastIndex(email, "@")
	if at <= 0 || at == len(email)-1 {
		writeError(w, http.StatusBadRequest, "emailaddress is required")
		return
	}
	domainName := strings.ToLower(email[at+1:])
	host := "mail." + domainName
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_, _ = fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<clientConfig version="1.1">
  <emailProvider id="%s">
    <domain>%s</domain>
    <displayName>ProIdentity Mail</displayName>
    <displayShortName>ProIdentity</displayShortName>
    <incomingServer type="imap">
      <hostname>%s</hostname>
      <port>993</port>
      <socketType>SSL</socketType>
      <authentication>password-cleartext</authentication>
      <username>%%EMAILADDRESS%%</username>
    </incomingServer>
    <incomingServer type="pop3">
      <hostname>%s</hostname>
      <port>995</port>
      <socketType>SSL</socketType>
      <authentication>password-cleartext</authentication>
      <username>%%EMAILADDRESS%%</username>
    </incomingServer>
    <outgoingServer type="smtp">
      <hostname>%s</hostname>
      <port>587</port>
      <socketType>STARTTLS</socketType>
      <authentication>password-cleartext</authentication>
      <username>%%EMAILADDRESS%%</username>
    </outgoingServer>
  </emailProvider>
</clientConfig>
`, xmlEscape(domainName), xmlEscape(domainName), xmlEscape(host), xmlEscape(host), xmlEscape(host))
}

func (h handler) serviceDiscovery(w http.ResponseWriter, r *http.Request) {
	email := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("emailaddress")))
	at := strings.LastIndex(email, "@")
	if at <= 0 || at == len(email)-1 {
		writeError(w, http.StatusBadRequest, "emailaddress is required")
		return
	}
	domainName := email[at+1:]
	host := "mail." + domainName
	base := "https://" + host
	writeJSON(w, http.StatusOK, map[string]any{
		"email": email,
		"services": map[string]any{
			"imap":    map[string]any{"hostname": host, "port": 993, "security": "tls", "username": email},
			"pop3":    map[string]any{"hostname": host, "port": 995, "security": "tls", "username": email},
			"smtp":    map[string]any{"hostname": host, "port": 587, "security": "starttls", "username": email},
			"caldav":  map[string]any{"url": base + "/dav/calendars/" + email + "/default/"},
			"carddav": map[string]any{"url": base + "/dav/addressbooks/" + email + "/default/"},
			"webmail": map[string]any{"url": base + "/"},
		},
	})
}

func wellKnownDAV(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dav/", http.StatusTemporaryRedirect)
}

func xmlEscape(value string) string {
	var builder strings.Builder
	_ = xml.EscapeText(&builder, []byte(value))
	return builder.String()
}

func (h handler) listTenants(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	tenants, err := h.store.ListTenants(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list tenants failed")
		return
	}
	writeJSON(w, http.StatusOK, tenants)
}

func (h handler) createTenant(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Name == "" || req.Slug == "" {
		writeError(w, http.StatusBadRequest, "name and slug are required")
		return
	}
	tenant, err := h.store.CreateTenant(r.Context(), domain.Tenant{Name: req.Name, Slug: req.Slug})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create tenant failed")
		return
	}
	writeJSON(w, http.StatusCreated, tenant)
}

func (h handler) createDomain(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	var req struct {
		TenantID uint64 `json:"tenant_id"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.TenantID == 0 || req.Name == "" {
		writeError(w, http.StatusBadRequest, "tenant_id and name are required")
		return
	}
	mailDomain, err := h.store.CreateDomain(r.Context(), domain.Domain{TenantID: req.TenantID, Name: req.Name})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create domain failed")
		return
	}
	writeJSON(w, http.StatusCreated, mailDomain)
}

func (h handler) listDomains(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	domains, err := h.store.ListDomains(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list domains failed")
		return
	}
	writeJSON(w, http.StatusOK, domains)
}

func (h handler) createUser(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	var req struct {
		TenantID        uint64 `json:"tenant_id"`
		PrimaryDomainID uint64 `json:"primary_domain_id"`
		LocalPart       string `json:"local_part"`
		DisplayName     string `json:"display_name"`
		Password        string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.TenantID == 0 || req.PrimaryDomainID == 0 || req.LocalPart == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "tenant_id, primary_domain_id, local_part, and password are required")
		return
	}
	hash, err := security.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid password")
		return
	}
	user, err := h.store.CreateUser(r.Context(), domain.User{
		TenantID:        req.TenantID,
		PrimaryDomainID: req.PrimaryDomainID,
		LocalPart:       req.LocalPart,
		DisplayName:     req.DisplayName,
		PasswordHash:    hash,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create user failed")
		return
	}
	user.PasswordHash = ""
	writeJSON(w, http.StatusCreated, user)
}

func (h handler) listUsers(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	users, err := h.store.ListUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list users failed")
		return
	}
	for i := range users {
		users[i].PasswordHash = ""
	}
	writeJSON(w, http.StatusOK, users)
}

func (h handler) getDomainDNS(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	domainID, err := strconv.ParseUint(chi.URLParam(r, "domainID"), 10, 64)
	if err != nil || domainID == 0 {
		writeError(w, http.StatusBadRequest, "valid domain id is required")
		return
	}
	dns, err := h.store.GetDomainDNS(r.Context(), domainID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get domain dns failed")
		return
	}
	writeJSON(w, http.StatusOK, dns)
}

func (h handler) listQuarantineEvents(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	events, err := h.store.ListQuarantineEvents(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list quarantine events failed")
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (h handler) releaseQuarantineEvent(w http.ResponseWriter, r *http.Request) {
	h.resolveQuarantineEvent(w, r, "released")
}

func (h handler) deleteQuarantineEvent(w http.ResponseWriter, r *http.Request) {
	h.resolveQuarantineEvent(w, r, "deleted")
}

func (h handler) resolveQuarantineEvent(w http.ResponseWriter, r *http.Request, status string) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	eventID, err := strconv.ParseUint(chi.URLParam(r, "eventID"), 10, 64)
	if err != nil || eventID == 0 {
		writeError(w, http.StatusBadRequest, "valid quarantine event id is required")
		return
	}
	var req struct {
		ResolutionNote string `json:"resolution_note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && r.Body != http.NoBody {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	event, err := h.store.ResolveQuarantineEvent(r.Context(), eventID, status, strings.TrimSpace(req.ResolutionNote))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "resolve quarantine event failed")
		return
	}
	writeJSON(w, http.StatusOK, event)
}

func (h handler) listAuditEvents(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	events, err := h.store.ListAuditEvents(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list audit events failed")
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (h handler) listTenantPolicies(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	policies, err := h.store.ListTenantPolicies(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list tenant policies failed")
		return
	}
	writeJSON(w, http.StatusOK, policies)
}

func (h handler) updateTenantPolicy(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	tenantID, err := strconv.ParseUint(chi.URLParam(r, "tenantID"), 10, 64)
	if err != nil || tenantID == 0 {
		writeError(w, http.StatusBadRequest, "valid tenant id is required")
		return
	}
	var req struct {
		SpamAction        string `json:"spam_action"`
		MalwareAction     string `json:"malware_action"`
		RequireTLSForAuth bool   `json:"require_tls_for_auth"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.SpamAction = strings.ToLower(strings.TrimSpace(req.SpamAction))
	req.MalwareAction = strings.ToLower(strings.TrimSpace(req.MalwareAction))
	if req.SpamAction != "mark" && req.SpamAction != "quarantine" && req.SpamAction != "reject" {
		writeError(w, http.StatusBadRequest, "spam_action must be mark, quarantine, or reject")
		return
	}
	if req.MalwareAction != "quarantine" && req.MalwareAction != "reject" {
		writeError(w, http.StatusBadRequest, "malware_action must be quarantine or reject")
		return
	}
	policy, err := h.store.UpdateTenantPolicy(r.Context(), domain.TenantPolicy{
		TenantID:          tenantID,
		SpamAction:        req.SpamAction,
		MalwareAction:     req.MalwareAction,
		RequireTLSForAuth: req.RequireTLSForAuth,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "update tenant policy failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{TenantID: &tenantID, ActorType: "admin", Action: "tenant_policy.update", TargetType: "tenant_policy", TargetID: strconv.FormatUint(tenantID, 10), MetadataJSON: fmt.Sprintf(`{"spam_action":%q,"malware_action":%q,"require_tls_for_auth":%t}`, policy.SpamAction, policy.MalwareAction, policy.RequireTLSForAuth)})
	writeJSON(w, http.StatusOK, policy)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
