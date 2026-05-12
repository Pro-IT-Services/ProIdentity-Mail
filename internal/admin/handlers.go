package admin

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/mail"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"proidentity-mail/internal/configdrift"
	"proidentity-mail/internal/domain"
	"proidentity-mail/internal/i18n"
	"proidentity-mail/internal/security"
	"proidentity-mail/internal/session"
)

type Store interface {
	CreateTenant(ctx context.Context, tenant domain.Tenant) (domain.Tenant, error)
	ListTenants(ctx context.Context) ([]domain.Tenant, error)
	UpdateTenant(ctx context.Context, tenant domain.Tenant) (domain.Tenant, error)
	DeleteTenant(ctx context.Context, tenantID uint64) error
	CreateDomain(ctx context.Context, mailDomain domain.Domain) (domain.Domain, error)
	ListDomains(ctx context.Context) ([]domain.Domain, error)
	UpdateDomain(ctx context.Context, mailDomain domain.Domain) (domain.Domain, error)
	DeleteDomain(ctx context.Context, domainID uint64) error
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	ListUsers(ctx context.Context) ([]domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) (domain.User, error)
	DeleteUser(ctx context.Context, userID uint64) error
	UnlockUser(ctx context.Context, userID uint64) (domain.User, error)
	ResetUserMFA(ctx context.Context, userID uint64) error
	CreateTenantAdmin(ctx context.Context, admin domain.TenantAdmin) (domain.TenantAdmin, error)
	ListTenantAdmins(ctx context.Context) ([]domain.TenantAdmin, error)
	DeleteTenantAdmin(ctx context.Context, adminID uint64) error
	ListLoginRateLimits(ctx context.Context) ([]domain.LoginRateLimit, error)
	ClearLoginRateLimit(ctx context.Context, limitID uint64) error
	GetAdminMFASettings(ctx context.Context) (domain.AdminMFASettings, error)
	SaveAdminMFASettings(ctx context.Context, settings domain.AdminMFASettings) (domain.AdminMFASettings, error)
	CreateAdminMFAChallenge(ctx context.Context, challenge domain.AdminMFAChallenge) error
	GetAdminMFAChallenge(ctx context.Context, token string) (domain.AdminMFAChallenge, error)
	DeleteAdminMFAChallenge(ctx context.Context, token string) error
	ListAdminWebAuthnCredentials(ctx context.Context) ([]domain.AdminWebAuthnCredential, error)
	CreateAdminWebAuthnCredential(ctx context.Context, credential domain.AdminWebAuthnCredential) (domain.AdminWebAuthnCredential, error)
	UpdateAdminWebAuthnCredential(ctx context.Context, credential domain.AdminWebAuthnCredential) error
	CreateAdminWebAuthnSession(ctx context.Context, session domain.AdminWebAuthnSession) error
	GetAdminWebAuthnSession(ctx context.Context, token string) (domain.AdminWebAuthnSession, error)
	DeleteAdminWebAuthnSession(ctx context.Context, token string) error
	CreateAlias(ctx context.Context, alias domain.Alias) (domain.Alias, error)
	ListAliases(ctx context.Context) ([]domain.Alias, error)
	UpdateAlias(ctx context.Context, alias domain.Alias) (domain.Alias, error)
	DeleteAlias(ctx context.Context, aliasID uint64) error
	CreateCatchAllRoute(ctx context.Context, route domain.CatchAllRoute) (domain.CatchAllRoute, error)
	ListCatchAllRoutes(ctx context.Context) ([]domain.CatchAllRoute, error)
	UpdateCatchAllRoute(ctx context.Context, route domain.CatchAllRoute) (domain.CatchAllRoute, error)
	DeleteCatchAllRoute(ctx context.Context, routeID uint64) error
	CreateSharedMailboxPermission(ctx context.Context, permission domain.SharedMailboxPermission) (domain.SharedMailboxPermission, error)
	ListSharedMailboxPermissions(ctx context.Context) ([]domain.SharedMailboxPermission, error)
	UpdateSharedMailboxPermission(ctx context.Context, permission domain.SharedMailboxPermission) (domain.SharedMailboxPermission, error)
	DeleteSharedMailboxPermission(ctx context.Context, permissionID uint64) error
	ListQuarantineEvents(ctx context.Context) ([]domain.QuarantineEvent, error)
	ResolveQuarantineEvent(ctx context.Context, eventID uint64, status, note string) (domain.QuarantineEvent, error)
	ListAuditEvents(ctx context.Context) ([]domain.AuditEvent, error)
	RecordAuditEvent(ctx context.Context, event domain.AuditEvent) error
	ListTenantPolicies(ctx context.Context) ([]domain.TenantPolicy, error)
	UpdateTenantPolicy(ctx context.Context, policy domain.TenantPolicy) (domain.TenantPolicy, error)
	GetMailServerSettings(ctx context.Context) (domain.MailServerSettings, error)
	UpdateMailServerSettings(ctx context.Context, settings domain.MailServerSettings) (domain.MailServerSettings, error)
	GetDomainDNS(ctx context.Context, domainID uint64) (domain.DomainDNS, error)
	GetDomainTLS(ctx context.Context, domainID uint64) (domain.DomainTLS, error)
	ListAvailableTLSCertificates(ctx context.Context) ([]domain.TLSCertificate, error)
	UpdateDomainTLSSettings(ctx context.Context, settings domain.DomainTLSSettings) (domain.DomainTLSSettings, error)
	CreateTLSCertificateJob(ctx context.Context, job domain.TLSCertificateJob) (domain.TLSCertificateJob, error)
	GetCloudflareConfig(ctx context.Context, domainID uint64) (domain.CloudflareConfig, error)
	SaveCloudflareConfig(ctx context.Context, domainID uint64, zoneID, apiToken string) (domain.CloudflareConfig, error)
	CheckCloudflareDNS(ctx context.Context, domainID uint64) (domain.DNSProvisionPlan, error)
	ApplyCloudflareDNS(ctx context.Context, domainID uint64, replace bool) (domain.DNSProvisionResult, error)
}

type TenantAdminGrantStore interface {
	GetTenantAdminGrants(ctx context.Context, email string) ([]domain.TenantAdmin, error)
}

type handler struct {
	store  Store
	auth   AuthConfig
	system SystemConfig
}

type adminPrincipal struct {
	Subject string
	Super   bool
	Grants  []domain.TenantAdmin
}

type adminPrincipalContextKey struct{}

const defaultSystemSharedMailboxQuotaBytes uint64 = 1073741824

var defaultSystemSharedMailboxes = []struct {
	LocalPart   string
	DisplayName string
}{
	{LocalPart: "postmaster", DisplayName: "Postmaster"},
	{LocalPart: "abuse", DisplayName: "Abuse Desk"},
	{LocalPart: "dmarc", DisplayName: "DMARC Reports"},
	{LocalPart: "tlsrpt", DisplayName: "TLS Reports"},
}

type AuthConfig struct {
	Username          string
	Password          string
	Sessions          *session.Manager
	Limiter           session.Limiter
	AuthPolicyLimiter session.Limiter
	DiscoveryLimiter  session.Limiter
	AuthPolicyToken   string
	System            SystemConfig
}

type CommandRunner func(ctx context.Context, command string, args ...string) ([]byte, error)

type SystemConfig struct {
	MailctlPath            string
	ConfigApplyRequestPath string
	LiveRoot               string
	CommandTimeout         time.Duration
	CommandRunner          CommandRunner
}

func NewRouter(store Store, authConfig ...AuthConfig) http.Handler {
	var auth AuthConfig
	if len(authConfig) > 0 {
		auth = authConfig[0]
	}
	if auth.DiscoveryLimiter == nil {
		auth.DiscoveryLimiter = session.NewLoginLimiter(discoveryLimiterOptions())
	}
	h := handler{store: store, auth: auth, system: normalizeSystemConfig(auth.System)}
	r := chi.NewRouter()
	r.Use(security.BrowserHeaders)
	r.Use(security.LimitRequestBody(1 << 20))
	r.Get("/healthz", health)
	r.Get("/.well-known/autoconfig/mail/config-v1.1.xml", h.mailAutoconfig)
	r.Get("/mail/config-v1.1.xml", h.mailAutoconfig)
	r.Get("/autodiscover/autodiscover.xml", h.outlookAutodiscover)
	r.Post("/autodiscover/autodiscover.xml", h.outlookAutodiscover)
	r.Get("/Autodiscover/Autodiscover.xml", h.outlookAutodiscover)
	r.Post("/Autodiscover/Autodiscover.xml", h.outlookAutodiscover)
	r.Get("/.well-known/proidentity-mail/config.json", h.serviceDiscovery)
	r.Get("/.well-known/caldav", wellKnownDAV)
	r.Head("/.well-known/caldav", wellKnownDAV)
	r.Get("/.well-known/carddav", wellKnownDAV)
	r.Head("/.well-known/carddav", wellKnownDAV)
	r.Get("/", h.index)
	r.Get("/api/v1/session", h.login)
	r.Post("/api/v1/session", h.login)
	r.Post("/api/v1/session/mfa", h.verifySessionMFA)
	r.Post("/api/v1/session/mfa/webauthn", h.finishSessionWebAuthn)
	r.Delete("/api/v1/session", h.logout)
	r.Post("/internal/dovecot/auth-policy", h.dovecotAuthPolicy)
	r.Group(func(protected chi.Router) {
		protected.Use(h.requireAdmin)
		protected.Get("/api/v1/tenants", h.listTenants)
		protected.Post("/api/v1/session/step-up", h.beginAdminStepUp)
		protected.Post("/api/v1/session/step-up/verify", h.verifyAdminStepUp)
		protected.Post("/api/v1/session/step-up/webauthn", h.finishAdminStepUpWebAuthn)
		protected.Post("/api/v1/tenants", h.createTenant)
		protected.Put("/api/v1/tenants/{tenantID}", h.updateTenant)
		protected.Delete("/api/v1/tenants/{tenantID}", h.deleteTenant)
		protected.Get("/api/v1/domains", h.listDomains)
		protected.Post("/api/v1/domains", h.createDomain)
		protected.Put("/api/v1/domains/{domainID}", h.updateDomain)
		protected.Delete("/api/v1/domains/{domainID}", h.deleteDomain)
		protected.Get("/api/v1/domains/{domainID}/dns", h.getDomainDNS)
		protected.Get("/api/v1/domains/{domainID}/tls", h.getDomainTLS)
		protected.Put("/api/v1/domains/{domainID}/tls/settings", h.updateDomainTLSSettings)
		protected.Post("/api/v1/domains/{domainID}/tls/jobs", h.createTLSCertificateJob)
		protected.Get("/api/v1/tls/certificates", h.listAvailableTLSCertificates)
		protected.Get("/api/v1/domains/{domainID}/cloudflare", h.getCloudflareConfig)
		protected.Put("/api/v1/domains/{domainID}/cloudflare", h.saveCloudflareConfig)
		protected.Post("/api/v1/domains/{domainID}/cloudflare/check", h.checkCloudflareDNS)
		protected.Post("/api/v1/domains/{domainID}/cloudflare/apply", h.applyCloudflareDNS)
		protected.Get("/api/v1/users", h.listUsers)
		protected.Post("/api/v1/users", h.createUser)
		protected.Put("/api/v1/users/{userID}", h.updateUser)
		protected.Delete("/api/v1/users/{userID}", h.deleteUser)
		protected.Post("/api/v1/users/{userID}/unlock", h.unlockUser)
		protected.Post("/api/v1/users/{userID}/mfa/reset", h.resetUserMFA)
		protected.Get("/api/v1/tenant-admins", h.listTenantAdmins)
		protected.Post("/api/v1/tenant-admins", h.createTenantAdmin)
		protected.Delete("/api/v1/tenant-admins/{tenantAdminID}", h.deleteTenantAdmin)
		protected.Get("/api/v1/security/login-rate-limits", h.listLoginRateLimits)
		protected.Delete("/api/v1/security/login-rate-limits/{limitID}", h.clearLoginRateLimit)
		protected.Get("/api/v1/admin-mfa/settings", h.getAdminMFASettings)
		protected.Put("/api/v1/admin-mfa/proidentity", h.updateProIdentityAuthSettings)
		protected.Post("/api/v1/admin-mfa/totp/enroll", h.createAdminTOTPEnrollment)
		protected.Post("/api/v1/admin-mfa/totp/verify", h.verifyAdminTOTPEnrollment)
		protected.Post("/api/v1/admin-mfa/proidentity/totp/enroll", h.createProIdentityTOTPEnrollment)
		protected.Post("/api/v1/admin-mfa/proidentity/totp/verify", h.verifyProIdentityTOTPEnrollment)
		protected.Post("/api/v1/admin-mfa/proidentity/totp/confirm", h.confirmProIdentityTOTPEnrollment)
		protected.Post("/api/v1/admin-mfa/webauthn/register/begin", h.beginAdminWebAuthnRegistration)
		protected.Post("/api/v1/admin-mfa/webauthn/register/finish", h.finishAdminWebAuthnRegistration)
		protected.Get("/api/v1/aliases", h.listAliases)
		protected.Post("/api/v1/aliases", h.createAlias)
		protected.Put("/api/v1/aliases/{aliasID}", h.updateAlias)
		protected.Delete("/api/v1/aliases/{aliasID}", h.deleteAlias)
		protected.Get("/api/v1/catch-all", h.listCatchAllRoutes)
		protected.Post("/api/v1/catch-all", h.createCatchAllRoute)
		protected.Put("/api/v1/catch-all/{routeID}", h.updateCatchAllRoute)
		protected.Delete("/api/v1/catch-all/{routeID}", h.deleteCatchAllRoute)
		protected.Get("/api/v1/shared-permissions", h.listSharedMailboxPermissions)
		protected.Post("/api/v1/shared-permissions", h.createSharedMailboxPermission)
		protected.Put("/api/v1/shared-permissions/{permissionID}", h.updateSharedMailboxPermission)
		protected.Delete("/api/v1/shared-permissions/{permissionID}", h.deleteSharedMailboxPermission)
		protected.Get("/api/v1/quarantine", h.listQuarantineEvents)
		protected.Post("/api/v1/quarantine/{eventID}/release", h.releaseQuarantineEvent)
		protected.Post("/api/v1/quarantine/{eventID}/delete", h.deleteQuarantineEvent)
		protected.Get("/api/v1/audit", h.listAuditEvents)
		protected.Get("/api/v1/policies", h.listTenantPolicies)
		protected.Put("/api/v1/policies/{tenantID}", h.updateTenantPolicy)
		protected.Get("/api/v1/mail-server-settings", h.getMailServerSettings)
		protected.Put("/api/v1/mail-server-settings", h.updateMailServerSettings)
		protected.Get("/api/v1/system/config-drift", h.getConfigDrift)
		protected.Post("/api/v1/system/config-apply", h.applyConfigDrift)
	})
	return r
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func normalizeSystemConfig(cfg SystemConfig) SystemConfig {
	if strings.TrimSpace(cfg.MailctlPath) == "" {
		cfg.MailctlPath = "/opt/proidentity-mail/bin/mailctl"
	}
	if strings.TrimSpace(cfg.ConfigApplyRequestPath) == "" {
		cfg.ConfigApplyRequestPath = "/etc/proidentity-mail/apply-request"
	}
	if cfg.CommandTimeout <= 0 {
		cfg.CommandTimeout = 45 * time.Second
	}
	if cfg.CommandRunner == nil {
		cfg.CommandRunner = execSystemCommand
	}
	return cfg
}

func execSystemCommand(ctx context.Context, command string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	return cmd.CombinedOutput()
}

func (h handler) getConfigDrift(w http.ResponseWriter, r *http.Request) {
	if !requireSuperAdmin(w, r) {
		return
	}
	renderRoot, err := os.MkdirTemp("", "proidentity-config-drift-*")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create drift workspace failed")
		return
	}
	defer os.RemoveAll(renderRoot)

	mailDir := filepath.Join(renderRoot, "mail")
	proxyDir := filepath.Join(renderRoot, "proxy")
	if err := h.runSystemCommand(r.Context(), h.system.MailctlPath, "render", "--target-dir", mailDir); err != nil {
		writeError(w, http.StatusInternalServerError, "render mail config failed: "+err.Error())
		return
	}
	if err := h.runSystemCommand(r.Context(), h.system.MailctlPath, "render-proxy", "--target-dir", proxyDir); err != nil {
		writeError(w, http.StatusInternalServerError, "render proxy config failed: "+err.Error())
		return
	}
	report := configdrift.Compare(r.Context(), configdrift.DefaultMappings(mailDir, proxyDir, h.system.LiveRoot))
	writeJSON(w, http.StatusOK, report)
}

func (h handler) applyConfigDrift(w http.ResponseWriter, r *http.Request) {
	if !requireSuperAdmin(w, r) {
		return
	}
	if !h.requireAdminStepUp(w, r) {
		return
	}
	path, err := h.queueConfigApplyRequest()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	metadata, _ := json.Marshal(map[string]string{"request_path": path})
	h.recordAudit(r.Context(), domain.AuditEvent{
		ActorType:    "admin",
		Action:       "system.config_apply_requested",
		TargetType:   "system_config",
		TargetID:     "all",
		Category:     "system",
		Severity:     "warning",
		MetadataJSON: string(metadata),
	})
	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":       "queued",
		"request_path": path,
	})
}

func (h handler) queueConfigApplyRequest() (string, error) {
	path := strings.TrimSpace(h.system.ConfigApplyRequestPath)
	if path == "" {
		return "", errors.New("config apply request path is not configured")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return "", fmt.Errorf("create apply request directory failed: %w", err)
	}
	body := "requested_at=" + time.Now().UTC().Format(time.RFC3339Nano) + "\n"
	if err := os.WriteFile(path, []byte(body), 0640); err != nil {
		return "", fmt.Errorf("queue config apply failed: %w", err)
	}
	return path, nil
}

func (h handler) runSystemCommand(ctx context.Context, command string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, h.system.CommandTimeout)
	defer cancel()
	output, err := h.system.CommandRunner(ctx, command, args...)
	if err == nil {
		return nil
	}
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return err
	}
	if len(detail) > 1200 {
		detail = detail[:1200] + "..."
	}
	return fmt.Errorf("%w: %s", err, detail)
}

func (h handler) index(w http.ResponseWriter, r *http.Request) {
	nonce, err := security.NewCSPNonce()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create security nonce failed")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", security.BrowserCSP(nonce))
	html := strings.Replace(adminIndexHTML, "__PROIDENTITY_I18N_CATALOG__", i18n.CatalogJSON(), 1)
	html = strings.ReplaceAll(html, "__PROIDENTITY_CSP_NONCE__", nonce)
	_, _ = w.Write([]byte(html))
}

func (h handler) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.auth.Username == "" && h.auth.Password == "" {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), adminPrincipalContextKey{}, adminPrincipal{Subject: "development", Super: true})))
			return
		}
		if h.auth.Sessions != nil {
			var current session.Session
			var ok bool
			if !safeMethod(r.Method) {
				current, ok = h.auth.Sessions.ValidateUnsafe(r)
				if !ok {
					http.Error(w, "csrf required", http.StatusForbidden)
					return
				}
			} else {
				current, ok = h.auth.Sessions.Validate(r)
				if !ok {
					writeError(w, http.StatusUnauthorized, "unauthorized")
					return
				}
			}
			if current.Kind != "admin" {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			principal, ok := h.adminPrincipalForSubject(r.Context(), current.Subject)
			if !ok {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), adminPrincipalContextKey{}, principal)))
			return
		}
		username, password, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(username), []byte(h.auth.Username)) != 1 || subtle.ConstantTimeCompare([]byte(password), []byte(h.auth.Password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="ProIdentity Admin", charset="UTF-8"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), adminPrincipalContextKey{}, adminPrincipal{Subject: username, Super: true})))
	})
}

func (h handler) adminPrincipalForSubject(ctx context.Context, subject string) (adminPrincipal, bool) {
	subject = strings.ToLower(strings.TrimSpace(subject))
	if subject == "" {
		return adminPrincipal{}, false
	}
	if strings.EqualFold(subject, strings.TrimSpace(h.auth.Username)) {
		return adminPrincipal{Subject: subject, Super: true}, true
	}
	grantStore, ok := h.store.(TenantAdminGrantStore)
	if !ok {
		return adminPrincipal{}, false
	}
	grants, err := grantStore.GetTenantAdminGrants(ctx, subject)
	if err != nil {
		return adminPrincipal{}, false
	}
	active := make([]domain.TenantAdmin, 0, len(grants))
	for _, grant := range grants {
		if grant.TenantID == 0 || !strings.EqualFold(grant.Status, "active") {
			continue
		}
		grant.Role = normalizeTenantAdminRole(grant.Role)
		active = append(active, grant)
	}
	if len(active) == 0 {
		return adminPrincipal{}, false
	}
	return adminPrincipal{Subject: subject, Grants: active}, true
}

func currentAdminPrincipal(ctx context.Context) adminPrincipal {
	principal, _ := ctx.Value(adminPrincipalContextKey{}).(adminPrincipal)
	return principal
}

func requireSuperAdmin(w http.ResponseWriter, r *http.Request) bool {
	if currentAdminPrincipal(r.Context()).Super {
		return true
	}
	writeError(w, http.StatusForbidden, "super admin required")
	return false
}

func (h handler) requireAdminStepUp(w http.ResponseWriter, r *http.Request) bool {
	if h.auth.Sessions == nil {
		return true
	}
	if h.auth.Sessions.HasRecentStepUp(r) {
		return true
	}
	writeError(w, http.StatusPreconditionRequired, "fresh admin mfa step-up required")
	return false
}

func requireTenantRead(w http.ResponseWriter, r *http.Request, tenantID uint64) bool {
	if currentAdminPrincipal(r.Context()).CanReadTenant(tenantID) {
		return true
	}
	writeError(w, http.StatusForbidden, "tenant access denied")
	return false
}

func requireTenantWrite(w http.ResponseWriter, r *http.Request, tenantID uint64) bool {
	if currentAdminPrincipal(r.Context()).CanWriteTenant(tenantID) {
		return true
	}
	writeError(w, http.StatusForbidden, "tenant write access denied")
	return false
}

func (p adminPrincipal) CanReadTenant(tenantID uint64) bool {
	if p.Super {
		return true
	}
	for _, grant := range p.Grants {
		if grant.TenantID == tenantID && strings.EqualFold(grant.Status, "active") {
			return true
		}
	}
	return false
}

func (p adminPrincipal) CanWriteTenant(tenantID uint64) bool {
	if p.Super {
		return true
	}
	for _, grant := range p.Grants {
		if grant.TenantID == tenantID && strings.EqualFold(grant.Status, "active") && normalizeTenantAdminRole(grant.Role) == "tenant_admin" {
			return true
		}
	}
	return false
}

func filterTenantsForPrincipal(items []domain.Tenant, principal adminPrincipal) []domain.Tenant {
	if principal.Super {
		return items
	}
	filtered := make([]domain.Tenant, 0, len(items))
	for _, item := range items {
		if principal.CanReadTenant(item.ID) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func filterDomainsForPrincipal(items []domain.Domain, principal adminPrincipal) []domain.Domain {
	if principal.Super {
		return items
	}
	filtered := make([]domain.Domain, 0, len(items))
	for _, item := range items {
		if principal.CanReadTenant(item.TenantID) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func filterUsersForPrincipal(items []domain.User, principal adminPrincipal) []domain.User {
	if principal.Super {
		return items
	}
	filtered := make([]domain.User, 0, len(items))
	for _, item := range items {
		if principal.CanReadTenant(item.TenantID) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func (h handler) requireDomainRead(w http.ResponseWriter, r *http.Request, domainID uint64) bool {
	principal := currentAdminPrincipal(r.Context())
	if principal.Super {
		return true
	}
	tenantID, found, err := h.tenantIDForDomainID(r.Context(), domainID)
	if err != nil {
		writeStoreError(w, err, "load domain failed")
		return false
	}
	if !found {
		writeError(w, http.StatusNotFound, "domain not found")
		return false
	}
	return requireTenantRead(w, r, tenantID)
}

func (h handler) requireDomainWrite(w http.ResponseWriter, r *http.Request, domainID uint64) bool {
	principal := currentAdminPrincipal(r.Context())
	if principal.Super {
		return true
	}
	tenantID, found, err := h.tenantIDForDomainID(r.Context(), domainID)
	if err != nil {
		writeStoreError(w, err, "load domain failed")
		return false
	}
	if !found {
		writeError(w, http.StatusNotFound, "domain not found")
		return false
	}
	return requireTenantWrite(w, r, tenantID)
}

func (h handler) requireUserWrite(w http.ResponseWriter, r *http.Request, userID uint64) bool {
	principal := currentAdminPrincipal(r.Context())
	if principal.Super {
		return true
	}
	tenantID, found, err := h.tenantIDForUserID(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err, "load user failed")
		return false
	}
	if !found {
		writeError(w, http.StatusNotFound, "user not found")
		return false
	}
	return requireTenantWrite(w, r, tenantID)
}

func (h handler) tenantIDForDomainID(ctx context.Context, domainID uint64) (uint64, bool, error) {
	domains, err := h.store.ListDomains(ctx)
	if err != nil {
		return 0, false, err
	}
	for _, item := range domains {
		if item.ID == domainID {
			return item.TenantID, true, nil
		}
	}
	return 0, false, nil
}

func (h handler) tenantIDForUserID(ctx context.Context, userID uint64) (uint64, bool, error) {
	users, err := h.store.ListUsers(ctx)
	if err != nil {
		return 0, false, err
	}
	for _, item := range users {
		if item.ID == userID {
			return item.TenantID, true, nil
		}
	}
	return 0, false, nil
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
	keys := loginKeys("admin", req.Username, r)
	if session.AnyLocked(h.auth.Limiter, keys) {
		writeError(w, http.StatusTooManyRequests, "login temporarily locked")
		return
	}
	if subtle.ConstantTimeCompare([]byte(req.Username), []byte(h.auth.Username)) != 1 || subtle.ConstantTimeCompare([]byte(req.Password), []byte(h.auth.Password)) != 1 {
		session.FailAll(h.auth.Limiter, keys)
		h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "admin.login_failed", TargetType: "admin", TargetID: req.Username, MetadataJSON: auditJSON(map[string]any{"client_ip": requestClientIP(r), "user_agent": r.UserAgent()})})
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if h.store != nil {
		settings, err := h.store.GetAdminMFASettings(r.Context())
		if err != nil {
			writeStoreError(w, err, "mfa settings unavailable")
			return
		}
		mfaResponse, err := h.beginAdminMFA(r.Context(), r, req.Username, settings)
		if err != nil {
			if writeProIdentityAPIError(w, err) {
				return
			}
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		if mfaResponse != nil {
			writeJSON(w, http.StatusOK, mfaResponse)
			return
		}
	}
	session.SuccessAll(h.auth.Limiter, keys)
	created, err := h.auth.Sessions.Create(r, req.Username, "admin")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create session failed")
		return
	}
	http.SetCookie(w, created.Cookie)
	h.recordAdminLoginSuccess(r.Context(), r, req.Username, "")
	writeJSON(w, http.StatusOK, map[string]string{"csrf_token": created.CSRFToken})
}

func loginKeys(service, subject string, r *http.Request) []string {
	subject = strings.ToLower(strings.TrimSpace(subject))
	host := strings.ToLower(strings.TrimSpace(requestClientIP(r)))
	return []string{
		service + "|ip|" + host,
		service + "|account|" + subject,
		service + "|pair|" + subject + "|" + host,
	}
}

type dovecotAuthPolicyRequest struct {
	Login     string `json:"login"`
	Remote    string `json:"remote"`
	Protocol  string `json:"protocol"`
	Success   any    `json:"success"`
	FailType  string `json:"fail_type"`
	SessionID string `json:"session_id"`
}

func (h handler) dovecotAuthPolicy(w http.ResponseWriter, r *http.Request) {
	if !isLoopbackRequest(r) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	token := h.auth.AuthPolicyToken
	if token == "" {
		writeError(w, http.StatusServiceUnavailable, "auth policy unavailable")
		return
	}
	if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-ProIdentity-Auth-Policy")), []byte(token)) != 1 {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var req dovecotAuthPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	limiter := h.auth.AuthPolicyLimiter
	if limiter == nil {
		limiter = h.auth.Limiter
	}
	keys := dovecotAuthPolicyKeys(req.Protocol, req.Login, req.Remote, r.RemoteAddr)
	command := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("command")))
	if command == "" {
		command = "allow"
	}

	switch command {
	case "allow":
		if session.AnyLocked(limiter, keys) {
			writeJSON(w, http.StatusOK, map[string]any{"status": -1, "msg": "login temporarily locked"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]int{"status": 0})
	case "report":
		if dovecotPolicySuccess(req.Success) {
			session.SuccessAll(limiter, keys)
		} else if shouldRecordDovecotPolicyFailure(req.FailType) {
			session.FailAll(limiter, keys)
		}
		writeJSON(w, http.StatusOK, map[string]int{"status": 0})
	default:
		writeError(w, http.StatusBadRequest, "unsupported command")
	}
}

func isLoopbackRequest(r *http.Request) bool {
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		if value := strings.TrimSpace(r.Header.Get(header)); value != "" {
			if header == "X-Forwarded-For" {
				value = strings.TrimSpace(strings.Split(value, ",")[0])
			}
			ip := net.ParseIP(value)
			if ip == nil || !ip.IsLoopback() {
				return false
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(strings.TrimSpace(host))
	return ip != nil && ip.IsLoopback()
}

func dovecotAuthPolicyKeys(protocol, login, remote, fallbackRemote string) []string {
	protocol = normalizeDovecotPolicyKeyPart(protocol, "unknown")
	login = strings.ToLower(strings.TrimSpace(login))
	remote = normalizeDovecotRemote(remote, fallbackRemote)
	keys := make([]string, 0, 6)
	if remote != "" {
		keys = append(keys, protocol+"|ip|"+remote)
	}
	if login != "" {
		keys = append(keys, protocol+"|account|"+login)
		if remote != "" {
			keys = append(keys, protocol+"|pair|"+login+"|"+remote)
		}
	}
	if remote != "" {
		keys = append(keys, "dovecot|ip|"+remote)
	}
	if login != "" {
		keys = append(keys, "dovecot|account|"+login)
		if remote != "" {
			keys = append(keys, "dovecot|pair|"+login+"|"+remote)
		}
	}
	return keys
}

func normalizeDovecotRemote(remote, fallbackRemote string) string {
	for i, candidate := range []string{remote, fallbackRemote} {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if host, _, err := net.SplitHostPort(candidate); err == nil {
			candidate = host
		}
		if ip := net.ParseIP(candidate); ip != nil {
			if i > 0 && ip.IsLoopback() {
				continue
			}
			return ip.String()
		}
	}
	return "unknown"
}

func normalizeDovecotPolicyKeyPart(value, fallback string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return fallback
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return fallback
	}
	return value
}

func dovecotPolicySuccess(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "1", "true", "yes", "y", "ok", "success":
			return true
		}
	case float64:
		return typed != 0
	}
	return false
}

func shouldRecordDovecotPolicyFailure(failType string) bool {
	switch strings.ToLower(strings.TrimSpace(failType)) {
	case "", "credentials", "account", "expired", "disabled":
		return true
	case "policy", "internal":
		return false
	default:
		return true
	}
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

func (h handler) recordAdminLoginSuccess(ctx context.Context, r *http.Request, username, provider string) {
	clientIP := requestClientIP(r)
	if clientIP != "" && !h.adminLoginIPSeen(ctx, username, clientIP) {
		h.recordAudit(ctx, domain.AuditEvent{
			ActorType:    "system",
			Action:       "security.alert.admin_new_ip",
			TargetType:   "admin",
			TargetID:     username,
			MetadataJSON: auditJSON(map[string]any{"username": username, "client_ip": clientIP, "mfa_provider": provider}),
		})
	}
	metadata := map[string]any{"client_ip": clientIP, "user_agent": r.UserAgent()}
	if provider != "" {
		metadata["mfa_provider"] = provider
	}
	h.recordAudit(ctx, domain.AuditEvent{ActorType: "admin", Action: "admin.login", TargetType: "admin", TargetID: username, MetadataJSON: auditJSON(metadata)})
}

func (h handler) adminLoginIPSeen(ctx context.Context, username, clientIP string) bool {
	if h.store == nil {
		return true
	}
	events, err := h.store.ListAuditEvents(ctx)
	if err != nil {
		return true
	}
	for _, event := range events {
		if event.Action != "admin.login" || !strings.EqualFold(event.TargetID, username) {
			continue
		}
		if metaString(auditMetadata(event.MetadataJSON), "client_ip") == clientIP {
			return true
		}
	}
	return false
}

func auditJSON(metadata map[string]any) string {
	body, err := json.Marshal(metadata)
	if err != nil {
		return `{}`
	}
	return string(body)
}

func safeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

func discoveryLimiterOptions() session.Options {
	return session.Options{MaxFailures: 60, Lockout: 5 * time.Minute, Window: time.Minute}
}

func (h handler) allowDiscoveryClientRequest(w http.ResponseWriter, r *http.Request, endpoint string) bool {
	return h.allowDiscoveryRateLimitKeys(w, discoveryClientRateLimitKeys(endpoint, r))
}

func (h handler) allowDiscoveryDomainRequest(w http.ResponseWriter, endpoint, email string) bool {
	return h.allowDiscoveryRateLimitKeys(w, discoveryDomainRateLimitKeys(endpoint, email))
}

func (h handler) allowDiscoveryRateLimitKeys(w http.ResponseWriter, keys []string) bool {
	if session.AnyLocked(h.auth.DiscoveryLimiter, keys) {
		writeError(w, http.StatusTooManyRequests, "automatic setup is temporarily rate limited")
		return false
	}
	session.FailAll(h.auth.DiscoveryLimiter, keys)
	return true
}

func discoveryClientRateLimitKeys(endpoint string, r *http.Request) []string {
	endpoint = normalizeDiscoveryRateLimitPart(endpoint)
	host := strings.ToLower(strings.TrimSpace(requestClientIP(r)))
	return []string{endpoint + "|ip|" + host}
}

func discoveryDomainRateLimitKeys(endpoint, email string) []string {
	endpoint = normalizeDiscoveryRateLimitPart(endpoint)
	if _, domainName, ok := strings.Cut(strings.ToLower(strings.TrimSpace(email)), "@"); ok {
		if domainName = normalizeDiscoveryRateLimitPart(domainName); domainName != "" {
			return []string{endpoint + "|domain|" + domainName}
		}
	}
	return nil
}

func normalizeDiscoveryRateLimitPart(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "unknown"
	}
	var builder strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '-' || r == '_' || r == '.':
			builder.WriteRune(r)
		default:
			builder.WriteByte('_')
		}
	}
	if builder.Len() == 0 {
		return "unknown"
	}
	return builder.String()
}

func (h handler) mailAutoconfig(w http.ResponseWriter, r *http.Request) {
	if !h.allowDiscoveryClientRequest(w, r, "autoconfig") {
		return
	}
	email, ok := normalizeEmailAddress(r.URL.Query().Get("emailaddress"))
	if !ok {
		writeError(w, http.StatusBadRequest, "emailaddress is required")
		return
	}
	if !h.allowDiscoveryDomainRequest(w, "autoconfig", email) {
		return
	}
	domainName := email[strings.LastIndex(email, "@")+1:]
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
	if !h.allowDiscoveryClientRequest(w, r, "service-discovery") {
		return
	}
	email, ok := normalizeEmailAddress(r.URL.Query().Get("emailaddress"))
	if !ok {
		writeError(w, http.StatusBadRequest, "emailaddress is required")
		return
	}
	if !h.allowDiscoveryDomainRequest(w, "service-discovery", email) {
		return
	}
	domainName := email[strings.LastIndex(email, "@")+1:]
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

func (h handler) outlookAutodiscover(w http.ResponseWriter, r *http.Request) {
	if !h.allowDiscoveryClientRequest(w, r, "autodiscover") {
		return
	}
	email, ok := normalizeEmailAddress(autodiscoverEmail(r))
	if !ok {
		writeError(w, http.StatusBadRequest, "emailaddress is required")
		return
	}
	if !h.allowDiscoveryDomainRequest(w, "autodiscover", email) {
		return
	}
	domainName := email[strings.LastIndex(email, "@")+1:]
	host := "mail." + domainName
	base := "https://" + host
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	_, _ = fmt.Fprintf(w, `<?xml version="1.0" encoding="utf-8"?>
<Autodiscover xmlns="http://schemas.microsoft.com/exchange/autodiscover/responseschema/2006">
  <Response xmlns="http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a">
    <User>
      <DisplayName>%s</DisplayName>
      <EMailAddress>%s</EMailAddress>
    </User>
    <Account>
      <AccountType>email</AccountType>
      <Action>settings</Action>
      <Protocol>
        <Type>IMAP</Type>
        <Server>%s</Server>
        <Port>993</Port>
        <DomainRequired>off</DomainRequired>
        <LoginName>%s</LoginName>
        <SPA>off</SPA>
        <SSL>on</SSL>
        <AuthRequired>on</AuthRequired>
      </Protocol>
      <Protocol>
        <Type>POP3</Type>
        <Server>%s</Server>
        <Port>995</Port>
        <DomainRequired>off</DomainRequired>
        <LoginName>%s</LoginName>
        <SPA>off</SPA>
        <SSL>on</SSL>
        <AuthRequired>on</AuthRequired>
      </Protocol>
      <Protocol>
        <Type>SMTP</Type>
        <Server>%s</Server>
        <Port>587</Port>
        <DomainRequired>off</DomainRequired>
        <LoginName>%s</LoginName>
        <SPA>off</SPA>
        <SSL>on</SSL>
        <Encryption>TLS</Encryption>
        <AuthRequired>on</AuthRequired>
        <UsePOPAuth>off</UsePOPAuth>
      </Protocol>
      <Protocol>
        <Type>DAV</Type>
        <Server>%s/dav/</Server>
      </Protocol>
    </Account>
  </Response>
</Autodiscover>
`, xmlEscape(email), xmlEscape(email), xmlEscape(host), xmlEscape(email), xmlEscape(host), xmlEscape(email), xmlEscape(host), xmlEscape(email), xmlEscape(base))
}

func autodiscoverEmail(r *http.Request) string {
	for _, key := range []string{"emailaddress", "email", "Email"} {
		if value := strings.ToLower(strings.TrimSpace(r.URL.Query().Get(key))); value != "" {
			return value
		}
	}
	var req struct {
		Request struct {
			EMailAddress string `xml:"EMailAddress"`
		} `xml:"Request"`
	}
	if err := xml.NewDecoder(r.Body).Decode(&req); err != nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(req.Request.EMailAddress))
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
	tenants = filterTenantsForPrincipal(tenants, currentAdminPrincipal(r.Context()))
	writeJSON(w, http.StatusOK, tenants)
}

func (h handler) createTenant(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	if !requireSuperAdmin(w, r) {
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
	req.Name = strings.TrimSpace(req.Name)
	var ok bool
	req.Slug, ok = normalizeTenantSlug(req.Slug)
	if !ok {
		writeError(w, http.StatusBadRequest, "slug must contain only lowercase letters, numbers, and hyphens")
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
	h.recordAudit(r.Context(), domain.AuditEvent{TenantID: &tenant.ID, ActorType: "admin", Action: "tenant.create", TargetType: "tenant", TargetID: strconv.FormatUint(tenant.ID, 10), MetadataJSON: fmt.Sprintf(`{"name":%q,"slug":%q,"status":%q}`, tenant.Name, tenant.Slug, tenant.Status)})
	writeJSON(w, http.StatusCreated, tenant)
}

func (h handler) updateTenant(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	if !h.requireAdminStepUp(w, r) {
		return
	}
	tenantID, ok := parsePathUint(w, r, "tenantID", "valid tenant id is required")
	if !ok {
		return
	}
	var req struct {
		Name   string `json:"name"`
		Slug   string `json:"slug"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Slug, ok = normalizeTenantSlug(req.Slug)
	if !ok {
		writeError(w, http.StatusBadRequest, "slug must contain only lowercase letters, numbers, and hyphens")
		return
	}
	req.Status = strings.ToLower(strings.TrimSpace(req.Status))
	if req.Status == "" {
		req.Status = "active"
	}
	if req.Name == "" || req.Slug == "" {
		writeError(w, http.StatusBadRequest, "name and slug are required")
		return
	}
	if !validChoice(req.Status, "active", "suspended") {
		writeError(w, http.StatusBadRequest, "status must be active or suspended")
		return
	}
	tenant, err := h.store.UpdateTenant(r.Context(), domain.Tenant{ID: tenantID, Name: req.Name, Slug: req.Slug, Status: req.Status})
	if err != nil {
		writeStoreError(w, err, "update tenant failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{TenantID: &tenantID, ActorType: "admin", Action: "tenant.update", TargetType: "tenant", TargetID: strconv.FormatUint(tenantID, 10), MetadataJSON: fmt.Sprintf(`{"name":%q,"slug":%q,"status":%q}`, tenant.Name, tenant.Slug, tenant.Status)})
	writeJSON(w, http.StatusOK, tenant)
}

func (h handler) deleteTenant(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	tenantID, ok := parsePathUint(w, r, "tenantID", "valid tenant id is required")
	if !ok {
		return
	}
	if err := h.store.DeleteTenant(r.Context(), tenantID); err != nil {
		writeStoreError(w, err, "delete tenant failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{TenantID: &tenantID, ActorType: "admin", Action: "tenant.delete", TargetType: "tenant", TargetID: strconv.FormatUint(tenantID, 10), MetadataJSON: `{}`})
	w.WriteHeader(http.StatusNoContent)
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
	var ok bool
	req.Name, ok = normalizeDomainName(req.Name)
	if !ok {
		writeError(w, http.StatusBadRequest, "valid domain name is required")
		return
	}
	if req.TenantID == 0 || req.Name == "" {
		writeError(w, http.StatusBadRequest, "tenant_id and name are required")
		return
	}
	if !requireTenantWrite(w, r, req.TenantID) {
		return
	}
	mailDomain, err := h.store.CreateDomain(r.Context(), domain.Domain{TenantID: req.TenantID, Name: req.Name})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create domain failed")
		return
	}
	if err := h.createDefaultSystemSharedMailboxes(r.Context(), mailDomain); err != nil {
		writeError(w, http.StatusInternalServerError, "create system shared mailboxes failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{TenantID: &req.TenantID, ActorType: "admin", Action: "domain.create", TargetType: "domain", TargetID: strconv.FormatUint(mailDomain.ID, 10), MetadataJSON: fmt.Sprintf(`{"name":%q,"status":%q,"dkim_selector":%q}`, mailDomain.Name, mailDomain.Status, mailDomain.DKIMSelector)})
	writeJSON(w, http.StatusCreated, mailDomain)
}

func (h handler) createDefaultSystemSharedMailboxes(ctx context.Context, mailDomain domain.Domain) error {
	for _, mailbox := range defaultSystemSharedMailboxes {
		if _, err := h.store.CreateUser(ctx, domain.User{
			TenantID:        mailDomain.TenantID,
			PrimaryDomainID: mailDomain.ID,
			LocalPart:       mailbox.LocalPart,
			DisplayName:     mailbox.DisplayName,
			MailboxType:     "shared",
			QuotaBytes:      defaultSystemSharedMailboxQuotaBytes,
		}); err != nil {
			return err
		}
	}
	return nil
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
	domains = filterDomainsForPrincipal(domains, currentAdminPrincipal(r.Context()))
	writeJSON(w, http.StatusOK, domains)
}

func (h handler) updateDomain(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	domainID, ok := parsePathUint(w, r, "domainID", "valid domain id is required")
	if !ok {
		return
	}
	if !h.requireDomainWrite(w, r, domainID) {
		return
	}
	var req struct {
		TenantID     uint64 `json:"tenant_id"`
		Name         string `json:"name"`
		Status       string `json:"status"`
		DKIMSelector string `json:"dkim_selector"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Name, ok = normalizeDomainName(req.Name)
	if !ok {
		writeError(w, http.StatusBadRequest, "valid domain name is required")
		return
	}
	req.Status = strings.ToLower(strings.TrimSpace(req.Status))
	req.DKIMSelector = strings.TrimSpace(req.DKIMSelector)
	if req.Status == "" {
		req.Status = "pending"
	}
	if req.DKIMSelector == "" {
		req.DKIMSelector = "mail"
	}
	req.DKIMSelector, ok = normalizeDKIMSelector(req.DKIMSelector)
	if !ok {
		writeError(w, http.StatusBadRequest, "valid dkim selector is required")
		return
	}
	if req.TenantID == 0 || req.Name == "" {
		writeError(w, http.StatusBadRequest, "tenant_id and name are required")
		return
	}
	if !requireTenantWrite(w, r, req.TenantID) {
		return
	}
	if !validChoice(req.Status, "pending", "active", "disabled") {
		writeError(w, http.StatusBadRequest, "status must be pending, active, or disabled")
		return
	}
	mailDomain, err := h.store.UpdateDomain(r.Context(), domain.Domain{ID: domainID, TenantID: req.TenantID, Name: req.Name, Status: req.Status, DKIMSelector: req.DKIMSelector})
	if err != nil {
		writeStoreError(w, err, "update domain failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{TenantID: &req.TenantID, ActorType: "admin", Action: "domain.update", TargetType: "domain", TargetID: strconv.FormatUint(domainID, 10), MetadataJSON: fmt.Sprintf(`{"name":%q,"status":%q,"dkim_selector":%q}`, mailDomain.Name, mailDomain.Status, mailDomain.DKIMSelector)})
	writeJSON(w, http.StatusOK, mailDomain)
}

func (h handler) deleteDomain(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	domainID, ok := parsePathUint(w, r, "domainID", "valid domain id is required")
	if !ok {
		return
	}
	if !h.requireDomainWrite(w, r, domainID) {
		return
	}
	if !h.requireAdminStepUp(w, r) {
		return
	}
	if err := h.store.DeleteDomain(r.Context(), domainID); err != nil {
		writeStoreError(w, err, "delete domain failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "domain.delete", TargetType: "domain", TargetID: strconv.FormatUint(domainID, 10), MetadataJSON: `{}`})
	w.WriteHeader(http.StatusNoContent)
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
		MailboxType     string `json:"mailbox_type"`
		QuotaBytes      uint64 `json:"quota_bytes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var ok bool
	req.LocalPart, ok = normalizeLocalPart(req.LocalPart)
	if !ok {
		writeError(w, http.StatusBadRequest, "valid local part is required")
		return
	}
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.MailboxType = strings.ToLower(strings.TrimSpace(req.MailboxType))
	if req.MailboxType == "" {
		req.MailboxType = "user"
	}
	if req.MailboxType != "user" && req.MailboxType != "shared" {
		writeError(w, http.StatusBadRequest, "mailbox_type must be user or shared")
		return
	}
	if req.TenantID == 0 || req.PrimaryDomainID == 0 || req.LocalPart == "" || (req.MailboxType == "user" && req.Password == "") {
		writeError(w, http.StatusBadRequest, "tenant_id, primary_domain_id, local_part, and password are required for users")
		return
	}
	if !requireTenantWrite(w, r, req.TenantID) {
		return
	}
	var hash string
	if req.Password != "" {
		var err error
		hash, err = security.HashPassword(req.Password)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid password")
			return
		}
	}
	user, err := h.store.CreateUser(r.Context(), domain.User{
		TenantID:        req.TenantID,
		PrimaryDomainID: req.PrimaryDomainID,
		LocalPart:       req.LocalPart,
		DisplayName:     req.DisplayName,
		MailboxType:     req.MailboxType,
		PasswordHash:    hash,
		QuotaBytes:      req.QuotaBytes,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create user failed")
		return
	}
	user.PasswordHash = ""
	h.recordAudit(r.Context(), domain.AuditEvent{TenantID: &req.TenantID, ActorType: "admin", Action: "user.create", TargetType: "user", TargetID: strconv.FormatUint(user.ID, 10), MetadataJSON: fmt.Sprintf(`{"local_part":%q,"mailbox_type":%q,"quota_bytes":%d}`, user.LocalPart, user.MailboxType, user.QuotaBytes)})
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
	users = filterUsersForPrincipal(users, currentAdminPrincipal(r.Context()))
	writeJSON(w, http.StatusOK, users)
}

func (h handler) updateUser(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	userID, ok := parsePathUint(w, r, "userID", "valid user id is required")
	if !ok {
		return
	}
	if !h.requireUserWrite(w, r, userID) {
		return
	}
	var req struct {
		TenantID        uint64 `json:"tenant_id"`
		PrimaryDomainID uint64 `json:"primary_domain_id"`
		LocalPart       string `json:"local_part"`
		DisplayName     string `json:"display_name"`
		Password        string `json:"password"`
		MailboxType     string `json:"mailbox_type"`
		Status          string `json:"status"`
		QuotaBytes      uint64 `json:"quota_bytes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.LocalPart, ok = normalizeLocalPart(req.LocalPart)
	if !ok {
		writeError(w, http.StatusBadRequest, "valid local part is required")
		return
	}
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.MailboxType = strings.ToLower(strings.TrimSpace(req.MailboxType))
	req.Status = strings.ToLower(strings.TrimSpace(req.Status))
	if req.MailboxType == "" {
		req.MailboxType = "user"
	}
	if req.Status == "" {
		req.Status = "active"
	}
	if req.QuotaBytes == 0 {
		req.QuotaBytes = 10737418240
	}
	if req.TenantID == 0 || req.PrimaryDomainID == 0 || req.LocalPart == "" {
		writeError(w, http.StatusBadRequest, "tenant_id, primary_domain_id, and local_part are required")
		return
	}
	if !requireTenantWrite(w, r, req.TenantID) {
		return
	}
	if !validChoice(req.MailboxType, "user", "shared") {
		writeError(w, http.StatusBadRequest, "mailbox_type must be user or shared")
		return
	}
	if !validChoice(req.Status, "active", "locked", "disabled") {
		writeError(w, http.StatusBadRequest, "status must be active, locked, or disabled")
		return
	}
	var hash string
	if req.Password != "" && req.MailboxType == "user" {
		var err error
		hash, err = security.HashPassword(req.Password)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid password")
			return
		}
	}
	user, err := h.store.UpdateUser(r.Context(), domain.User{
		ID:              userID,
		TenantID:        req.TenantID,
		PrimaryDomainID: req.PrimaryDomainID,
		LocalPart:       req.LocalPart,
		DisplayName:     req.DisplayName,
		MailboxType:     req.MailboxType,
		PasswordHash:    hash,
		Status:          req.Status,
		QuotaBytes:      req.QuotaBytes,
	})
	if err != nil {
		writeStoreError(w, err, "update user failed")
		return
	}
	user.PasswordHash = ""
	h.recordAudit(r.Context(), domain.AuditEvent{TenantID: &req.TenantID, ActorType: "admin", Action: "user.update", TargetType: "user", TargetID: strconv.FormatUint(userID, 10), MetadataJSON: fmt.Sprintf(`{"local_part":%q,"status":%q,"mailbox_type":%q,"quota_bytes":%d}`, user.LocalPart, user.Status, user.MailboxType, user.QuotaBytes)})
	writeJSON(w, http.StatusOK, user)
}

func (h handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	userID, ok := parsePathUint(w, r, "userID", "valid user id is required")
	if !ok {
		return
	}
	if !h.requireUserWrite(w, r, userID) {
		return
	}
	if !h.requireAdminStepUp(w, r) {
		return
	}
	if err := h.store.DeleteUser(r.Context(), userID); err != nil {
		writeStoreError(w, err, "delete user failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "user.delete", TargetType: "user", TargetID: strconv.FormatUint(userID, 10), MetadataJSON: `{}`})
	w.WriteHeader(http.StatusNoContent)
}

func (h handler) unlockUser(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	userID, ok := parsePathUint(w, r, "userID", "valid user id is required")
	if !ok {
		return
	}
	if !h.requireUserWrite(w, r, userID) {
		return
	}
	if !h.requireAdminStepUp(w, r) {
		return
	}
	user, err := h.store.UnlockUser(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err, "unlock user failed")
		return
	}
	user.PasswordHash = ""
	h.recordAudit(r.Context(), domain.AuditEvent{TenantID: &user.TenantID, ActorType: "admin", Action: "user.unlock", TargetType: "user", TargetID: strconv.FormatUint(userID, 10), MetadataJSON: fmt.Sprintf(`{"local_part":%q,"status":%q}`, user.LocalPart, user.Status)})
	writeJSON(w, http.StatusOK, user)
}

func (h handler) resetUserMFA(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	userID, ok := parsePathUint(w, r, "userID", "valid user id is required")
	if !ok {
		return
	}
	if !h.requireUserWrite(w, r, userID) {
		return
	}
	if !h.requireAdminStepUp(w, r) {
		return
	}
	if err := h.store.ResetUserMFA(r.Context(), userID); err != nil {
		writeStoreError(w, err, "reset user mfa failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "user_mfa.reset", TargetType: "user", TargetID: strconv.FormatUint(userID, 10), MetadataJSON: `{}`})
	w.WriteHeader(http.StatusNoContent)
}

func (h handler) listTenantAdmins(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	admins, err := h.store.ListTenantAdmins(r.Context())
	if err != nil {
		writeStoreError(w, err, "list tenant admins failed")
		return
	}
	if admins == nil {
		admins = []domain.TenantAdmin{}
	}
	writeJSON(w, http.StatusOK, admins)
}

func (h handler) createTenantAdmin(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	var req domain.TenantAdmin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Role = strings.ToLower(strings.TrimSpace(req.Role))
	if req.Role == "" {
		req.Role = "tenant_admin"
	}
	req.Status = strings.ToLower(strings.TrimSpace(req.Status))
	if req.Status == "" {
		req.Status = "active"
	}
	if req.TenantID == 0 || req.UserID == 0 {
		writeError(w, http.StatusBadRequest, "tenant_id and user_id are required")
		return
	}
	if !validChoice(req.Role, "tenant_admin", "read_only") {
		writeError(w, http.StatusBadRequest, "role must be tenant_admin or read_only")
		return
	}
	if !validChoice(req.Status, "active", "disabled") {
		writeError(w, http.StatusBadRequest, "status must be active or disabled")
		return
	}
	admin, err := h.store.CreateTenantAdmin(r.Context(), req)
	if err != nil {
		writeStoreError(w, err, "create tenant admin failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{TenantID: &req.TenantID, ActorType: "admin", Action: "tenant_admin.create", TargetType: "tenant_admin", TargetID: strconv.FormatUint(admin.ID, 10), MetadataJSON: fmt.Sprintf(`{"user_id":%d,"role":%q,"status":%q}`, admin.UserID, admin.Role, admin.Status)})
	writeJSON(w, http.StatusCreated, admin)
}

func (h handler) deleteTenantAdmin(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	adminID, ok := parsePathUint(w, r, "tenantAdminID", "valid tenant admin id is required")
	if !ok {
		return
	}
	if err := h.store.DeleteTenantAdmin(r.Context(), adminID); err != nil {
		writeStoreError(w, err, "delete tenant admin failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "tenant_admin.delete", TargetType: "tenant_admin", TargetID: strconv.FormatUint(adminID, 10), MetadataJSON: `{}`})
	w.WriteHeader(http.StatusNoContent)
}

func (h handler) listLoginRateLimits(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	rows, err := h.store.ListLoginRateLimits(r.Context())
	if err != nil {
		writeStoreError(w, err, "list login rate limits failed")
		return
	}
	if rows == nil {
		rows = []domain.LoginRateLimit{}
	}
	writeJSON(w, http.StatusOK, rows)
}

func (h handler) clearLoginRateLimit(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	if !h.requireAdminStepUp(w, r) {
		return
	}
	limitID, ok := parsePathUint(w, r, "limitID", "valid rate limit id is required")
	if !ok {
		return
	}
	if err := h.store.ClearLoginRateLimit(r.Context(), limitID); err != nil {
		writeStoreError(w, err, "clear login rate limit failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "security.rate_limit.clear", TargetType: "login_rate_limit", TargetID: strconv.FormatUint(limitID, 10), MetadataJSON: `{}`})
	w.WriteHeader(http.StatusNoContent)
}

func (h handler) createAlias(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	var req struct {
		TenantID        uint64 `json:"tenant_id"`
		DomainID        uint64 `json:"domain_id"`
		SourceLocalPart string `json:"source_local_part"`
		Destination     string `json:"destination"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var ok bool
	req.SourceLocalPart, ok = normalizeLocalPart(req.SourceLocalPart)
	if !ok {
		writeError(w, http.StatusBadRequest, "valid source local part is required")
		return
	}
	req.Destination, ok = normalizeAddressList(req.Destination)
	if !ok {
		writeError(w, http.StatusBadRequest, "valid destination email is required")
		return
	}
	if req.TenantID == 0 || req.DomainID == 0 || strings.TrimSpace(req.SourceLocalPart) == "" || strings.TrimSpace(req.Destination) == "" {
		writeError(w, http.StatusBadRequest, "tenant_id, domain_id, source_local_part, and destination are required")
		return
	}
	if !requireTenantWrite(w, r, req.TenantID) {
		return
	}
	alias, err := h.store.CreateAlias(r.Context(), domain.Alias{TenantID: req.TenantID, DomainID: req.DomainID, SourceLocalPart: req.SourceLocalPart, Destination: req.Destination})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create alias failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{TenantID: &req.TenantID, ActorType: "admin", Action: "alias.create", TargetType: "alias", TargetID: strconv.FormatUint(alias.ID, 10), MetadataJSON: fmt.Sprintf(`{"source_local_part":%q,"destination":%q}`, alias.SourceLocalPart, alias.Destination)})
	writeJSON(w, http.StatusCreated, alias)
}

func (h handler) listAliases(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	aliases, err := h.store.ListAliases(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list aliases failed")
		return
	}
	if !currentAdminPrincipal(r.Context()).Super {
		principal := currentAdminPrincipal(r.Context())
		filtered := make([]domain.Alias, 0, len(aliases))
		for _, alias := range aliases {
			if principal.CanReadTenant(alias.TenantID) {
				filtered = append(filtered, alias)
			}
		}
		aliases = filtered
	}
	writeJSON(w, http.StatusOK, aliases)
}

func (h handler) updateAlias(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	aliasID, ok := parsePathUint(w, r, "aliasID", "valid alias id is required")
	if !ok {
		return
	}
	var req struct {
		TenantID        uint64 `json:"tenant_id"`
		DomainID        uint64 `json:"domain_id"`
		SourceLocalPart string `json:"source_local_part"`
		Destination     string `json:"destination"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.SourceLocalPart, ok = normalizeLocalPart(req.SourceLocalPart)
	if !ok {
		writeError(w, http.StatusBadRequest, "valid source local part is required")
		return
	}
	req.Destination, ok = normalizeAddressList(req.Destination)
	if !ok {
		writeError(w, http.StatusBadRequest, "valid destination email is required")
		return
	}
	if req.TenantID == 0 || req.DomainID == 0 || strings.TrimSpace(req.SourceLocalPart) == "" || strings.TrimSpace(req.Destination) == "" {
		writeError(w, http.StatusBadRequest, "tenant_id, domain_id, source_local_part, and destination are required")
		return
	}
	if !requireTenantWrite(w, r, req.TenantID) {
		return
	}
	alias, err := h.store.UpdateAlias(r.Context(), domain.Alias{ID: aliasID, TenantID: req.TenantID, DomainID: req.DomainID, SourceLocalPart: req.SourceLocalPart, Destination: req.Destination})
	if err != nil {
		writeStoreError(w, err, "update alias failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{TenantID: &req.TenantID, ActorType: "admin", Action: "alias.update", TargetType: "alias", TargetID: strconv.FormatUint(aliasID, 10), MetadataJSON: fmt.Sprintf(`{"source_local_part":%q,"destination":%q}`, alias.SourceLocalPart, alias.Destination)})
	writeJSON(w, http.StatusOK, alias)
}

func (h handler) deleteAlias(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	aliasID, ok := parsePathUint(w, r, "aliasID", "valid alias id is required")
	if !ok {
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	if err := h.store.DeleteAlias(r.Context(), aliasID); err != nil {
		writeStoreError(w, err, "delete alias failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "alias.delete", TargetType: "alias", TargetID: strconv.FormatUint(aliasID, 10), MetadataJSON: `{}`})
	w.WriteHeader(http.StatusNoContent)
}

func (h handler) createCatchAllRoute(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	var req struct {
		TenantID    uint64 `json:"tenant_id"`
		DomainID    uint64 `json:"domain_id"`
		Destination string `json:"destination"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var ok bool
	req.Destination, ok = normalizeAddressList(req.Destination)
	if !ok {
		writeError(w, http.StatusBadRequest, "valid destination email is required")
		return
	}
	if req.TenantID == 0 || req.DomainID == 0 || strings.TrimSpace(req.Destination) == "" {
		writeError(w, http.StatusBadRequest, "tenant_id, domain_id, and destination are required")
		return
	}
	if !requireTenantWrite(w, r, req.TenantID) {
		return
	}
	route, err := h.store.CreateCatchAllRoute(r.Context(), domain.CatchAllRoute{TenantID: req.TenantID, DomainID: req.DomainID, Destination: req.Destination})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create catch-all failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{TenantID: &req.TenantID, ActorType: "admin", Action: "catch_all.create", TargetType: "catch_all", TargetID: strconv.FormatUint(route.ID, 10), MetadataJSON: fmt.Sprintf(`{"destination":%q,"status":%q}`, route.Destination, route.Status)})
	writeJSON(w, http.StatusCreated, route)
}

func (h handler) listCatchAllRoutes(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	routes, err := h.store.ListCatchAllRoutes(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list catch-all failed")
		return
	}
	if !currentAdminPrincipal(r.Context()).Super {
		principal := currentAdminPrincipal(r.Context())
		filtered := make([]domain.CatchAllRoute, 0, len(routes))
		for _, route := range routes {
			if principal.CanReadTenant(route.TenantID) {
				filtered = append(filtered, route)
			}
		}
		routes = filtered
	}
	writeJSON(w, http.StatusOK, routes)
}

func (h handler) updateCatchAllRoute(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	routeID, ok := parsePathUint(w, r, "routeID", "valid catch-all id is required")
	if !ok {
		return
	}
	var req struct {
		TenantID    uint64 `json:"tenant_id"`
		DomainID    uint64 `json:"domain_id"`
		Destination string `json:"destination"`
		Status      string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Destination, ok = normalizeAddressList(req.Destination)
	if !ok {
		writeError(w, http.StatusBadRequest, "valid destination email is required")
		return
	}
	req.Status = strings.ToLower(strings.TrimSpace(req.Status))
	if req.Status == "" {
		req.Status = "active"
	}
	if req.TenantID == 0 || req.DomainID == 0 || strings.TrimSpace(req.Destination) == "" {
		writeError(w, http.StatusBadRequest, "tenant_id, domain_id, and destination are required")
		return
	}
	if !requireTenantWrite(w, r, req.TenantID) {
		return
	}
	if !validChoice(req.Status, "active", "disabled") {
		writeError(w, http.StatusBadRequest, "status must be active or disabled")
		return
	}
	route, err := h.store.UpdateCatchAllRoute(r.Context(), domain.CatchAllRoute{ID: routeID, TenantID: req.TenantID, DomainID: req.DomainID, Destination: req.Destination, Status: req.Status})
	if err != nil {
		writeStoreError(w, err, "update catch-all failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{TenantID: &req.TenantID, ActorType: "admin", Action: "catch_all.update", TargetType: "catch_all", TargetID: strconv.FormatUint(routeID, 10), MetadataJSON: fmt.Sprintf(`{"destination":%q,"status":%q}`, route.Destination, route.Status)})
	writeJSON(w, http.StatusOK, route)
}

func (h handler) deleteCatchAllRoute(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	routeID, ok := parsePathUint(w, r, "routeID", "valid catch-all id is required")
	if !ok {
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	if err := h.store.DeleteCatchAllRoute(r.Context(), routeID); err != nil {
		writeStoreError(w, err, "delete catch-all failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "catch_all.delete", TargetType: "catch_all", TargetID: strconv.FormatUint(routeID, 10), MetadataJSON: `{}`})
	w.WriteHeader(http.StatusNoContent)
}

func (h handler) createSharedMailboxPermission(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	var req struct {
		TenantID        uint64 `json:"tenant_id"`
		SharedMailboxID uint64 `json:"shared_mailbox_id"`
		UserID          uint64 `json:"user_id"`
		CanRead         bool   `json:"can_read"`
		CanSendAs       bool   `json:"can_send_as"`
		CanSendOnBehalf bool   `json:"can_send_on_behalf"`
		CanManage       bool   `json:"can_manage"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.TenantID == 0 || req.SharedMailboxID == 0 || req.UserID == 0 {
		writeError(w, http.StatusBadRequest, "tenant_id, shared_mailbox_id, and user_id are required")
		return
	}
	if !requireTenantWrite(w, r, req.TenantID) {
		return
	}
	if !req.CanRead && !req.CanSendAs && !req.CanSendOnBehalf && !req.CanManage {
		req.CanRead = true
	}
	permission, err := h.store.CreateSharedMailboxPermission(r.Context(), domain.SharedMailboxPermission{
		TenantID: req.TenantID, SharedMailboxID: req.SharedMailboxID, UserID: req.UserID,
		CanRead: req.CanRead, CanSendAs: req.CanSendAs, CanSendOnBehalf: req.CanSendOnBehalf, CanManage: req.CanManage,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create shared permission failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{TenantID: &req.TenantID, ActorType: "admin", Action: "shared_permission.create", TargetType: "shared_permission", TargetID: strconv.FormatUint(permission.ID, 10), MetadataJSON: fmt.Sprintf(`{"shared_mailbox_id":%d,"user_id":%d,"rights":%q}`, permission.SharedMailboxID, permission.UserID, sharedPermissionRights(permission))})
	writeJSON(w, http.StatusCreated, permission)
}

func (h handler) listSharedMailboxPermissions(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	permissions, err := h.store.ListSharedMailboxPermissions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list shared permissions failed")
		return
	}
	if !currentAdminPrincipal(r.Context()).Super {
		principal := currentAdminPrincipal(r.Context())
		filtered := make([]domain.SharedMailboxPermission, 0, len(permissions))
		for _, permission := range permissions {
			if principal.CanReadTenant(permission.TenantID) {
				filtered = append(filtered, permission)
			}
		}
		permissions = filtered
	}
	writeJSON(w, http.StatusOK, permissions)
}

func (h handler) updateSharedMailboxPermission(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	permissionID, ok := parsePathUint(w, r, "permissionID", "valid shared permission id is required")
	if !ok {
		return
	}
	var req struct {
		TenantID        uint64 `json:"tenant_id"`
		SharedMailboxID uint64 `json:"shared_mailbox_id"`
		UserID          uint64 `json:"user_id"`
		CanRead         bool   `json:"can_read"`
		CanSendAs       bool   `json:"can_send_as"`
		CanSendOnBehalf bool   `json:"can_send_on_behalf"`
		CanManage       bool   `json:"can_manage"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.TenantID == 0 || req.SharedMailboxID == 0 || req.UserID == 0 {
		writeError(w, http.StatusBadRequest, "tenant_id, shared_mailbox_id, and user_id are required")
		return
	}
	if !requireTenantWrite(w, r, req.TenantID) {
		return
	}
	if !req.CanRead && !req.CanSendAs && !req.CanSendOnBehalf && !req.CanManage {
		req.CanRead = true
	}
	permission, err := h.store.UpdateSharedMailboxPermission(r.Context(), domain.SharedMailboxPermission{
		ID: permissionID, TenantID: req.TenantID, SharedMailboxID: req.SharedMailboxID, UserID: req.UserID,
		CanRead: req.CanRead, CanSendAs: req.CanSendAs, CanSendOnBehalf: req.CanSendOnBehalf, CanManage: req.CanManage,
	})
	if err != nil {
		writeStoreError(w, err, "update shared permission failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{TenantID: &req.TenantID, ActorType: "admin", Action: "shared_permission.update", TargetType: "shared_permission", TargetID: strconv.FormatUint(permissionID, 10), MetadataJSON: fmt.Sprintf(`{"shared_mailbox_id":%d,"user_id":%d,"rights":%q}`, permission.SharedMailboxID, permission.UserID, sharedPermissionRights(permission))})
	writeJSON(w, http.StatusOK, permission)
}

func (h handler) deleteSharedMailboxPermission(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	permissionID, ok := parsePathUint(w, r, "permissionID", "valid shared permission id is required")
	if !ok {
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	if err := h.store.DeleteSharedMailboxPermission(r.Context(), permissionID); err != nil {
		writeStoreError(w, err, "delete shared permission failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "shared_permission.delete", TargetType: "shared_permission", TargetID: strconv.FormatUint(permissionID, 10), MetadataJSON: `{}`})
	w.WriteHeader(http.StatusNoContent)
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
	if !h.requireDomainRead(w, r, domainID) {
		return
	}
	dns, err := h.store.GetDomainDNS(r.Context(), domainID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get domain dns failed")
		return
	}
	writeJSON(w, http.StatusOK, dns)
}

func (h handler) getDomainTLS(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	domainID, ok := parsePathUint(w, r, "domainID", "valid domain id is required")
	if !ok {
		return
	}
	if !h.requireDomainRead(w, r, domainID) {
		return
	}
	tlsState, err := h.store.GetDomainTLS(r.Context(), domainID)
	if err != nil {
		writeStoreError(w, err, "get domain tls failed")
		return
	}
	writeJSON(w, http.StatusOK, tlsState)
}

func (h handler) listAvailableTLSCertificates(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	certificates, err := h.store.ListAvailableTLSCertificates(r.Context())
	if err != nil {
		writeStoreError(w, err, "list tls certificates failed")
		return
	}
	writeJSON(w, http.StatusOK, certificates)
}

func (h handler) updateDomainTLSSettings(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	domainID, ok := parsePathUint(w, r, "domainID", "valid domain id is required")
	if !ok {
		return
	}
	if !h.requireDomainWrite(w, r, domainID) {
		return
	}
	if !h.requireAdminStepUp(w, r) {
		return
	}
	var req domain.DomainTLSSettings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.DomainID = domainID
	req.TLSMode = strings.ToLower(strings.TrimSpace(req.TLSMode))
	if req.TLSMode == "" {
		req.TLSMode = "inherit"
	}
	if !validChoice(req.TLSMode, "inherit", "letsencrypt-dns-cloudflare", "letsencrypt-http", "custom", "disabled") {
		writeError(w, http.StatusBadRequest, "tls_mode must be inherit, letsencrypt-dns-cloudflare, letsencrypt-http, custom, or disabled")
		return
	}
	req.ChallengeType = strings.ToLower(strings.TrimSpace(req.ChallengeType))
	if req.ChallengeType == "" {
		req.ChallengeType = "dns-cloudflare"
	}
	if !validChoice(req.ChallengeType, "dns-cloudflare", "http-01", "manual-dns", "custom-import", "none") {
		writeError(w, http.StatusBadRequest, "challenge_type must be dns-cloudflare, http-01, manual-dns, custom-import, or none")
		return
	}
	req.CertificateName = strings.ToLower(strings.TrimSpace(req.CertificateName))
	if req.CertificateName != "" {
		var ok bool
		req.CertificateName, ok = normalizeDomainName(req.CertificateName)
		if !ok {
			writeError(w, http.StatusBadRequest, "valid certificate name is required")
			return
		}
	}
	req.CustomCertPath = strings.TrimSpace(req.CustomCertPath)
	req.CustomKeyPath = strings.TrimSpace(req.CustomKeyPath)
	req.CustomChainPath = strings.TrimSpace(req.CustomChainPath)
	if !allowedCertificatePath(req.CustomCertPath) || !allowedCertificatePath(req.CustomKeyPath) || !allowedCertificatePath(req.CustomChainPath) {
		writeError(w, http.StatusBadRequest, "custom certificate paths must be under /etc/proidentity-mail/certs, /var/lib/proidentity-mail/certs, /etc/letsencrypt/live, or /etc/ssl")
		return
	}
	settings, err := h.store.UpdateDomainTLSSettings(r.Context(), req)
	if err != nil {
		writeStoreError(w, err, "update domain tls settings failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "domain_tls_settings.update", TargetType: "domain", TargetID: strconv.FormatUint(domainID, 10), MetadataJSON: fmt.Sprintf(`{"tls_mode":%q,"challenge_type":%q,"dns_webmail_alias_enabled":%t,"dns_admin_alias_enabled":%t}`, settings.TLSMode, settings.ChallengeType, settings.DNSWebmailAliasEnabled, settings.DNSAdminAliasEnabled)})
	writeJSON(w, http.StatusOK, settings)
}

func (h handler) createTLSCertificateJob(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	domainID, ok := parsePathUint(w, r, "domainID", "valid domain id is required")
	if !ok {
		return
	}
	if !h.requireDomainWrite(w, r, domainID) {
		return
	}
	if !h.requireAdminStepUp(w, r) {
		return
	}
	var req struct {
		JobType       string   `json:"job_type"`
		ChallengeType string   `json:"challenge_type"`
		Hostnames     []string `json:"hostnames"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && r.Body != http.NoBody {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	job := domain.TLSCertificateJob{
		DomainID:      domainID,
		JobType:       strings.ToLower(strings.TrimSpace(req.JobType)),
		ChallengeType: strings.ToLower(strings.TrimSpace(req.ChallengeType)),
		Hostnames:     req.Hostnames,
		RequestedBy:   "admin",
	}
	if job.JobType == "" {
		job.JobType = "issue"
	}
	if !validChoice(job.JobType, "issue", "renew", "import", "deploy", "check") {
		writeError(w, http.StatusBadRequest, "job_type must be issue, renew, import, deploy, or check")
		return
	}
	if job.ChallengeType != "" && !validChoice(job.ChallengeType, "dns-cloudflare", "http-01", "manual-dns", "custom-import", "none") {
		writeError(w, http.StatusBadRequest, "challenge_type must be dns-cloudflare, http-01, manual-dns, custom-import, or none")
		return
	}
	hostnames, ok := normalizeHostnames(job.Hostnames)
	if !ok {
		writeError(w, http.StatusBadRequest, "hostnames must be valid domain names")
		return
	}
	job.Hostnames = hostnames
	created, err := h.store.CreateTLSCertificateJob(r.Context(), job)
	if err != nil {
		writeStoreError(w, err, "create tls certificate job failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "tls_certificate_job.create", TargetType: "domain", TargetID: strconv.FormatUint(domainID, 10), MetadataJSON: fmt.Sprintf(`{"job_type":%q,"challenge_type":%q,"hostnames":%d}`, created.JobType, created.ChallengeType, len(created.Hostnames))})
	writeJSON(w, http.StatusAccepted, created)
}

func (h handler) getCloudflareConfig(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	domainID, ok := parsePathUint(w, r, "domainID", "valid domain id is required")
	if !ok {
		return
	}
	if !h.requireDomainRead(w, r, domainID) {
		return
	}
	config, err := h.store.GetCloudflareConfig(r.Context(), domainID)
	if err != nil {
		writeStoreError(w, err, "get cloudflare config failed")
		return
	}
	writeJSON(w, http.StatusOK, config)
}

func (h handler) saveCloudflareConfig(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	domainID, ok := parsePathUint(w, r, "domainID", "valid domain id is required")
	if !ok {
		return
	}
	if !h.requireDomainWrite(w, r, domainID) {
		return
	}
	if !h.requireAdminStepUp(w, r) {
		return
	}
	var req struct {
		ZoneID   string `json:"zone_id"`
		APIToken string `json:"api_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	config, err := h.store.SaveCloudflareConfig(r.Context(), domainID, strings.TrimSpace(req.ZoneID), strings.TrimSpace(req.APIToken))
	if err != nil {
		writeStoreError(w, err, "save cloudflare config failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "cloudflare_config.update", TargetType: "domain", TargetID: strconv.FormatUint(domainID, 10), MetadataJSON: fmt.Sprintf(`{"zone_id":%q,"token_configured":%t}`, config.ZoneID, config.TokenConfigured)})
	writeJSON(w, http.StatusOK, config)
}

func (h handler) checkCloudflareDNS(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	domainID, ok := parsePathUint(w, r, "domainID", "valid domain id is required")
	if !ok {
		return
	}
	if !h.requireDomainRead(w, r, domainID) {
		return
	}
	plan, err := h.store.CheckCloudflareDNS(r.Context(), domainID)
	if err != nil {
		writeStoreError(w, err, "check cloudflare dns failed")
		return
	}
	writeJSON(w, http.StatusOK, plan)
}

func (h handler) applyCloudflareDNS(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	domainID, ok := parsePathUint(w, r, "domainID", "valid domain id is required")
	if !ok {
		return
	}
	if !h.requireDomainWrite(w, r, domainID) {
		return
	}
	if !h.requireAdminStepUp(w, r) {
		return
	}
	var req struct {
		Replace bool `json:"replace"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && r.Body != http.NoBody {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := h.store.ApplyCloudflareDNS(r.Context(), domainID, req.Replace)
	if err != nil {
		writeStoreError(w, err, "apply cloudflare dns failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "cloudflare_dns.apply", TargetType: "domain", TargetID: strconv.FormatUint(domainID, 10), MetadataJSON: fmt.Sprintf(`{"replace":%t,"changed":%d,"backup_id":%d}`, req.Replace, result.Changed, result.BackupID)})
	writeJSON(w, http.StatusOK, result)
}

func (h handler) listQuarantineEvents(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	if !requireSuperAdmin(w, r) {
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
	if !requireSuperAdmin(w, r) {
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
	if !requireSuperAdmin(w, r) {
		return
	}
	events, err := h.store.ListAuditEvents(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list audit events failed")
		return
	}
	writeJSON(w, http.StatusOK, enrichAuditEvents(events))
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
	if !currentAdminPrincipal(r.Context()).Super {
		filtered := make([]domain.TenantPolicy, 0, len(policies))
		principal := currentAdminPrincipal(r.Context())
		for _, policy := range policies {
			if principal.CanReadTenant(policy.TenantID) {
				filtered = append(filtered, policy)
			}
		}
		policies = filtered
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
	if !requireTenantWrite(w, r, tenantID) {
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

func (h handler) getMailServerSettings(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	settings, err := h.store.GetMailServerSettings(r.Context())
	if err != nil {
		writeStoreError(w, err, "get mail server settings failed")
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (h handler) updateMailServerSettings(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	if !h.requireAdminStepUp(w, r) {
		return
	}
	var req domain.MailServerSettings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.HostnameMode = strings.ToLower(strings.TrimSpace(req.HostnameMode))
	if req.HostnameMode == "" {
		req.HostnameMode = "shared"
	}
	if !validChoice(req.HostnameMode, "shared", "head-domain", "per-domain") {
		writeError(w, http.StatusBadRequest, "hostname_mode must be shared, head-domain, or per-domain")
		return
	}
	req.MailHostname = strings.ToLower(strings.TrimSpace(req.MailHostname))
	if req.MailHostname != "" {
		var ok bool
		req.MailHostname, ok = normalizeDomainName(req.MailHostname)
		if !ok {
			writeError(w, http.StatusBadRequest, "mail_hostname must be a valid domain name")
			return
		}
	}
	req.PublicIPv4 = strings.TrimSpace(req.PublicIPv4)
	req.PublicIPv6 = strings.TrimSpace(req.PublicIPv6)
	if req.PublicIPv4 != "" && net.ParseIP(req.PublicIPv4).To4() == nil {
		writeError(w, http.StatusBadRequest, "public_ipv4 must be a valid IPv4 address")
		return
	}
	if req.PublicIPv6 != "" {
		parsed := net.ParseIP(req.PublicIPv6)
		if parsed == nil || parsed.To4() != nil {
			writeError(w, http.StatusBadRequest, "public_ipv6 must be a valid IPv6 address")
			return
		}
	}
	req.TLSMode = strings.ToLower(strings.TrimSpace(req.TLSMode))
	if req.TLSMode == "" {
		req.TLSMode = "system"
	}
	if !validChoice(req.TLSMode, "system", "none", "behind-proxy", "letsencrypt-http", "letsencrypt-dns-cloudflare", "custom-cert") {
		writeError(w, http.StatusBadRequest, "tls_mode must be system, none, behind-proxy, letsencrypt-http, letsencrypt-dns-cloudflare, or custom-cert")
		return
	}
	req.HTTPSCertificateID = cleanUintPtr(req.HTTPSCertificateID)
	req.DefaultLanguage = i18n.NormalizeLanguage(req.DefaultLanguage)
	if req.DefaultLanguage == "" {
		writeError(w, http.StatusBadRequest, "default_language is not supported")
		return
	}
	settings, err := h.store.UpdateMailServerSettings(r.Context(), req)
	if err != nil {
		writeStoreError(w, err, "update mail server settings failed")
		return
	}
	if path, err := h.queueConfigApplyRequest(); err != nil {
		settings.ConfigApplyError = err.Error()
	} else {
		settings.ConfigApplyQueued = true
		metadata, _ := json.Marshal(map[string]string{"request_path": path, "source": "mail_server_settings"})
		h.recordAudit(r.Context(), domain.AuditEvent{
			ActorType:    "admin",
			Action:       "system.config_apply_requested",
			TargetType:   "system_config",
			TargetID:     "mail-server",
			Category:     "system",
			Severity:     "warning",
			MetadataJSON: string(metadata),
		})
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "mail_server_settings.update", TargetType: "system", TargetID: "mail-server", MetadataJSON: fmt.Sprintf(`{"hostname_mode":%q,"mail_hostname":%q,"sni_enabled":%t,"tls_mode":%q,"force_https":%t,"https_certificate_id":%q,"cloudflare_real_ip_enabled":%t}`, settings.HostnameMode, settings.MailHostname, settings.SNIEnabled, settings.TLSMode, settings.ForceHTTPS, uintPtrString(settings.HTTPSCertificateID), settings.CloudflareRealIPEnabled)})
	writeJSON(w, http.StatusOK, settings)
}

func parsePathUint(w http.ResponseWriter, r *http.Request, name, message string) (uint64, bool) {
	id, err := strconv.ParseUint(chi.URLParam(r, name), 10, 64)
	if err != nil || id == 0 {
		writeError(w, http.StatusBadRequest, message)
		return 0, false
	}
	return id, true
}

func uintPtrString(value *uint64) string {
	if value == nil {
		return ""
	}
	return strconv.FormatUint(*value, 10)
}

func normalizeTenantSlug(value string) (string, bool) {
	slug := strings.ToLower(strings.TrimSpace(value))
	if len(slug) == 0 || len(slug) > 63 {
		return "", false
	}
	if slug[0] == '-' || slug[len(slug)-1] == '-' {
		return "", false
	}
	for i := 0; i < len(slug); i++ {
		ch := slug[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			continue
		}
		return "", false
	}
	return slug, true
}

func normalizeDomainName(value string) (string, bool) {
	name := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(value), "."))
	if len(name) == 0 || len(name) > 253 || strings.Contains(name, "..") {
		return "", false
	}
	labels := strings.Split(name, ".")
	if len(labels) < 2 {
		return "", false
	}
	for _, label := range labels {
		if !validDNSLabel(label) {
			return "", false
		}
	}
	return name, true
}

func validDNSLabel(label string) bool {
	if len(label) == 0 || len(label) > 63 {
		return false
	}
	if label[0] == '-' || label[len(label)-1] == '-' {
		return false
	}
	for i := 0; i < len(label); i++ {
		ch := label[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			continue
		}
		return false
	}
	return true
}

func normalizeDKIMSelector(value string) (string, bool) {
	selector := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(value), "."))
	if len(selector) == 0 || len(selector) > 253 || strings.Contains(selector, "..") {
		return "", false
	}
	for _, label := range strings.Split(selector, ".") {
		if !validDNSLabel(label) {
			return "", false
		}
	}
	return selector, true
}

func normalizeLocalPart(value string) (string, bool) {
	local := strings.ToLower(strings.TrimSpace(value))
	if len(local) == 0 || len(local) > 64 || strings.Contains(local, "..") {
		return "", false
	}
	if local[0] == '.' || local[len(local)-1] == '.' {
		return "", false
	}
	for i := 0; i < len(local); i++ {
		ch := local[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '.' || ch == '_' || ch == '+' || ch == '-' {
			continue
		}
		return "", false
	}
	return local, true
}

func normalizeAddressList(value string) (string, bool) {
	raw := strings.ReplaceAll(value, ";", ",")
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		address, ok := normalizeEmailAddress(part)
		if !ok {
			return "", false
		}
		out = append(out, address)
	}
	if len(out) == 0 {
		return "", false
	}
	return strings.Join(out, ","), true
}

func normalizeEmailAddress(value string) (string, bool) {
	parsed, err := mail.ParseAddress(strings.TrimSpace(value))
	if err != nil || parsed.Address == "" {
		return "", false
	}
	parts := strings.Split(parsed.Address, "@")
	if len(parts) != 2 {
		return "", false
	}
	local, ok := normalizeLocalPart(parts[0])
	if !ok {
		return "", false
	}
	domainName, ok := normalizeDomainName(parts[1])
	if !ok {
		return "", false
	}
	return local + "@" + domainName, true
}

func normalizeHostnames(values []string) ([]string, bool) {
	if len(values) == 0 {
		return nil, true
	}
	out := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		hostname, ok := normalizeDomainName(value)
		if !ok {
			return nil, false
		}
		if !seen[hostname] {
			seen[hostname] = true
			out = append(out, hostname)
		}
	}
	return out, true
}

func allowedCertificatePath(value string) bool {
	path := strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	if path == "" {
		return true
	}
	if !strings.HasPrefix(path, "/") || strings.Contains(path, "\x00") || strings.Contains(path, "/../") || strings.HasSuffix(path, "/..") {
		return false
	}
	for _, prefix := range []string{
		"/etc/proidentity-mail/certs/",
		"/var/lib/proidentity-mail/certs/",
		"/etc/letsencrypt/live/",
		"/etc/ssl/certs/",
		"/etc/ssl/private/",
	} {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func validChoice(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

func writeStoreError(w http.ResponseWriter, err error, message string) {
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if errors.Is(err, errCloudflareTokenRequired) {
		writeError(w, http.StatusBadRequest, "cloudflare api token is required")
		return
	}
	if errors.Is(err, errCloudflareDNSConflicts) {
		writeError(w, http.StatusConflict, "dns conflicts detected; enable backup and replace after review")
		return
	}
	if errors.Is(err, errDomainDNSNotReady) {
		writeError(w, http.StatusConflict, "domain dns settings are incomplete; configure public mail hostname or server public IP first")
		return
	}
	if strings.Contains(strings.ToLower(err.Error()), "cloudflare") {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeError(w, http.StatusInternalServerError, message)
}

func sharedPermissionRights(permission domain.SharedMailboxPermission) string {
	rights := make([]string, 0, 4)
	if permission.CanRead {
		rights = append(rights, "read")
	}
	if permission.CanSendAs {
		rights = append(rights, "send_as")
	}
	if permission.CanSendOnBehalf {
		rights = append(rights, "send_on_behalf")
	}
	if permission.CanManage {
		rights = append(rights, "manage")
	}
	return strings.Join(rights, ",")
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
