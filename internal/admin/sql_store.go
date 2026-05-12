package admin

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"proidentity-mail/internal/domain"
	"proidentity-mail/internal/i18n"
	"proidentity-mail/internal/quarantine"
)

type SQLStore struct {
	db         *sql.DB
	quarantine quarantine.FileStore
	dns        DNSSettings
}

var (
	errCloudflareTokenRequired = errors.New("cloudflare api token is required")
	errCloudflareDNSConflicts  = errors.New("cloudflare dns conflicts detected")
	errDomainDNSNotReady       = errors.New("domain dns settings are incomplete")
)

type DNSSettings struct {
	MailHostname        string
	AdminHostname       string
	WebmailHostname     string
	PublicIPv4          string
	PublicIPv6          string
	TLSMode             string
	ForceHTTPS          bool
	HostnameMode        string
	HeadTenantID        uint64
	HeadDomainID        uint64
	SNIEnabled          bool
	DisableWebmailAlias bool
	DisableAdminAlias   bool
}

func NewSQLStore(db *sql.DB, stores ...quarantine.FileStore) SQLStore {
	fileStore := quarantine.FileStore{Root: "/var/lib/proidentity-mail/quarantine", MailRoot: "/var/vmail"}
	if len(stores) > 0 {
		fileStore = stores[0]
	}
	return SQLStore{db: db, quarantine: fileStore}
}

func (s SQLStore) WithDNSSettings(settings DNSSettings) SQLStore {
	s.dns = DNSSettings{
		MailHostname:        normalizeDNSName(settings.MailHostname),
		AdminHostname:       normalizeDNSName(settings.AdminHostname),
		WebmailHostname:     normalizeDNSName(settings.WebmailHostname),
		PublicIPv4:          strings.TrimSpace(settings.PublicIPv4),
		PublicIPv6:          strings.TrimSpace(settings.PublicIPv6),
		TLSMode:             normalizeProxyTLSMode(settings.TLSMode),
		ForceHTTPS:          settings.ForceHTTPS,
		HostnameMode:        normalizeHostnameMode(settings.HostnameMode),
		HeadTenantID:        settings.HeadTenantID,
		HeadDomainID:        settings.HeadDomainID,
		SNIEnabled:          settings.SNIEnabled,
		DisableWebmailAlias: settings.DisableWebmailAlias,
		DisableAdminAlias:   settings.DisableAdminAlias,
	}
	return s
}

func (s SQLStore) CreateTenant(ctx context.Context, tenant domain.Tenant) (domain.Tenant, error) {
	result, err := s.db.ExecContext(ctx, `INSERT INTO tenants(name, slug, status) VALUES (?, ?, 'active')`, tenant.Name, tenant.Slug)
	if err != nil {
		return domain.Tenant{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Tenant{}, err
	}
	tenant.ID = uint64(id)
	tenant.Status = "active"
	if _, err := s.db.ExecContext(ctx, `INSERT INTO tenant_policies(tenant_id) VALUES (?)`, tenant.ID); err != nil {
		return domain.Tenant{}, err
	}
	return tenant, nil
}

func (s SQLStore) ListTenants(ctx context.Context) ([]domain.Tenant, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, slug, status, created_at, updated_at
		FROM tenants
		ORDER BY created_at DESC, id DESC
		LIMIT 200`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []domain.Tenant
	for rows.Next() {
		var tenant domain.Tenant
		if err := rows.Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt); err != nil {
			return nil, err
		}
		tenants = append(tenants, tenant)
	}
	return tenants, rows.Err()
}

func (s SQLStore) UpdateTenant(ctx context.Context, tenant domain.Tenant) (domain.Tenant, error) {
	if _, err := execOne(ctx, s.db, `UPDATE tenants SET name = ?, slug = ?, status = ? WHERE id = ?`, tenant.Name, tenant.Slug, tenant.Status, tenant.ID); err != nil {
		return domain.Tenant{}, err
	}
	return scanTenant(ctx, s.db, tenant.ID)
}

func (s SQLStore) DeleteTenant(ctx context.Context, tenantID uint64) error {
	_, err := execOne(ctx, s.db, `UPDATE tenants SET status = 'suspended' WHERE id = ?`, tenantID)
	return err
}

func (s SQLStore) CreateDomain(ctx context.Context, mailDomain domain.Domain) (domain.Domain, error) {
	result, err := s.db.ExecContext(ctx, `INSERT INTO domains(tenant_id, name, status, dkim_selector) VALUES (?, ?, 'active', 'mail')`, mailDomain.TenantID, mailDomain.Name)
	if err != nil {
		return domain.Domain{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Domain{}, err
	}
	mailDomain.ID = uint64(id)
	mailDomain.Status = "active"
	mailDomain.DKIMSelector = "mail"
	return mailDomain, nil
}

func (s SQLStore) ListDomains(ctx context.Context) ([]domain.Domain, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant_id, name, status, dkim_selector, created_at, updated_at
		FROM domains
		ORDER BY created_at DESC, id DESC
		LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []domain.Domain
	for rows.Next() {
		var mailDomain domain.Domain
		if err := rows.Scan(&mailDomain.ID, &mailDomain.TenantID, &mailDomain.Name, &mailDomain.Status, &mailDomain.DKIMSelector, &mailDomain.CreatedAt, &mailDomain.UpdatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, mailDomain)
	}
	return domains, rows.Err()
}

func (s SQLStore) UpdateDomain(ctx context.Context, mailDomain domain.Domain) (domain.Domain, error) {
	if _, err := execOne(ctx, s.db, `UPDATE domains SET tenant_id = ?, name = ?, status = ?, dkim_selector = ? WHERE id = ?`, mailDomain.TenantID, strings.ToLower(strings.TrimSpace(mailDomain.Name)), mailDomain.Status, mailDomain.DKIMSelector, mailDomain.ID); err != nil {
		return domain.Domain{}, err
	}
	return scanDomain(ctx, s.db, mailDomain.ID)
}

func (s SQLStore) DeleteDomain(ctx context.Context, domainID uint64) error {
	_, err := execOne(ctx, s.db, `UPDATE domains SET status = 'disabled' WHERE id = ?`, domainID)
	return err
}

func (s SQLStore) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	if user.MailboxType == "" {
		user.MailboxType = "user"
	}
	if user.QuotaBytes == 0 {
		user.QuotaBytes = 10737418240
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO users(tenant_id, primary_domain_id, local_part, display_name, mailbox_type, password_hash, status, quota_bytes) VALUES (?, ?, ?, ?, ?, ?, 'active', ?)`,
		user.TenantID,
		user.PrimaryDomainID,
		user.LocalPart,
		user.DisplayName,
		user.MailboxType,
		user.PasswordHash,
		user.QuotaBytes,
	)
	if err != nil {
		return domain.User{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.User{}, err
	}
	user.ID = uint64(id)
	user.Status = "active"
	return user, nil
}

func (s SQLStore) ListUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant_id, primary_domain_id, local_part, display_name, mailbox_type, status, quota_bytes, created_at, updated_at
		FROM users
		ORDER BY created_at DESC, id DESC
		LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.TenantID, &user.PrimaryDomainID, &user.LocalPart, &user.DisplayName, &user.MailboxType, &user.Status, &user.QuotaBytes, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s SQLStore) UpdateUser(ctx context.Context, user domain.User) (domain.User, error) {
	localPart := strings.ToLower(strings.TrimSpace(user.LocalPart))
	displayName := strings.TrimSpace(user.DisplayName)
	if user.MailboxType == "shared" {
		if _, err := execOne(ctx, s.db, `
			UPDATE users
			SET tenant_id = ?, primary_domain_id = ?, local_part = ?, display_name = ?, mailbox_type = ?, password_hash = '', status = ?, quota_bytes = ?
			WHERE id = ?`,
			user.TenantID, user.PrimaryDomainID, localPart, displayName, user.MailboxType, user.Status, user.QuotaBytes, user.ID); err != nil {
			return domain.User{}, err
		}
	} else if user.PasswordHash != "" {
		if _, err := execOne(ctx, s.db, `
			UPDATE users
			SET tenant_id = ?, primary_domain_id = ?, local_part = ?, display_name = ?, mailbox_type = ?, password_hash = ?, status = ?, quota_bytes = ?
			WHERE id = ?`,
			user.TenantID, user.PrimaryDomainID, localPart, displayName, user.MailboxType, user.PasswordHash, user.Status, user.QuotaBytes, user.ID); err != nil {
			return domain.User{}, err
		}
	} else {
		if _, err := execOne(ctx, s.db, `
			UPDATE users
			SET tenant_id = ?, primary_domain_id = ?, local_part = ?, display_name = ?, mailbox_type = ?, status = ?, quota_bytes = ?
			WHERE id = ?`,
			user.TenantID, user.PrimaryDomainID, localPart, displayName, user.MailboxType, user.Status, user.QuotaBytes, user.ID); err != nil {
			return domain.User{}, err
		}
	}
	return scanUser(ctx, s.db, user.ID)
}

func (s SQLStore) DeleteUser(ctx context.Context, userID uint64) error {
	_, err := execOne(ctx, s.db, `UPDATE users SET status = 'disabled' WHERE id = ?`, userID)
	return err
}

func (s SQLStore) UnlockUser(ctx context.Context, userID uint64) (domain.User, error) {
	user, err := scanUser(ctx, s.db, userID)
	if err != nil {
		return domain.User{}, err
	}
	email, err := s.emailForUser(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.User{}, err
	}
	defer tx.Rollback()
	if _, err := execOne(ctx, tx, `UPDATE users SET status = 'active' WHERE id = ?`, userID); err != nil {
		return domain.User{}, err
	}
	if email != "" {
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM login_rate_limits
			WHERE limiter_key LIKE ?
			   OR limiter_key LIKE ?
			   OR limiter_key IN (?, ?)`,
			"%|account|"+email,
			"%|pair|"+email+"|%",
			"webmail|account|"+email,
			"dovecot|account|"+email,
		); err != nil {
			return domain.User{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.User{}, err
	}
	user.Status = "active"
	return user, nil
}

func (s SQLStore) ResetUserMFA(ctx context.Context, userID uint64) error {
	if _, err := scanUser(ctx, s.db, userID); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_mfa_challenges WHERE user_id = ?`, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE user_mfa_settings
		SET totp_enabled = FALSE, totp_secret = '', pending_totp_secret = ''
		WHERE user_id = ?`, userID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s SQLStore) CreateTenantAdmin(ctx context.Context, admin domain.TenantAdmin) (domain.TenantAdmin, error) {
	admin.Role = normalizeTenantAdminRole(admin.Role)
	admin.Status = normalizeTenantAdminStatus(admin.Status)
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO tenant_admins(tenant_id, user_id, role, status)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE role = VALUES(role), status = VALUES(status), updated_at = current_timestamp()`,
		admin.TenantID, admin.UserID, admin.Role, admin.Status)
	if err != nil {
		return domain.TenantAdmin{}, err
	}
	id, _ := result.LastInsertId()
	if id == 0 {
		err = s.db.QueryRowContext(ctx, `SELECT id FROM tenant_admins WHERE tenant_id = ? AND user_id = ?`, admin.TenantID, admin.UserID).Scan(&id)
		if err != nil {
			return domain.TenantAdmin{}, err
		}
	}
	return scanTenantAdmin(ctx, s.db, uint64(id))
}

func (s SQLStore) ListTenantAdmins(ctx context.Context) ([]domain.TenantAdmin, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant_id, user_id, role, status, created_at, updated_at
		FROM tenant_admins
		ORDER BY tenant_id, user_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	admins := make([]domain.TenantAdmin, 0)
	for rows.Next() {
		var admin domain.TenantAdmin
		if err := rows.Scan(&admin.ID, &admin.TenantID, &admin.UserID, &admin.Role, &admin.Status, &admin.CreatedAt, &admin.UpdatedAt); err != nil {
			return nil, err
		}
		admins = append(admins, admin)
	}
	return admins, rows.Err()
}

func (s SQLStore) GetTenantAdminGrants(ctx context.Context, email string) ([]domain.TenantAdmin, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT ta.id, ta.tenant_id, ta.user_id, ta.role, ta.status, ta.created_at, ta.updated_at
		FROM tenant_admins ta
		JOIN users u ON u.id = ta.user_id
		JOIN domains d ON d.id = u.primary_domain_id
		JOIN tenants t ON t.id = ta.tenant_id
		WHERE CONCAT(u.local_part, '@', d.name) = ?
		  AND ta.status = 'active'
		  AND u.status = 'active'
		  AND t.status = 'active'
		ORDER BY ta.tenant_id`, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	grants := make([]domain.TenantAdmin, 0)
	for rows.Next() {
		var grant domain.TenantAdmin
		if err := rows.Scan(&grant.ID, &grant.TenantID, &grant.UserID, &grant.Role, &grant.Status, &grant.CreatedAt, &grant.UpdatedAt); err != nil {
			return nil, err
		}
		grants = append(grants, grant)
	}
	return grants, rows.Err()
}

func (s SQLStore) DeleteTenantAdmin(ctx context.Context, adminID uint64) error {
	_, err := execOne(ctx, s.db, `DELETE FROM tenant_admins WHERE id = ?`, adminID)
	return err
}

func (s SQLStore) emailForUser(ctx context.Context, userID uint64) (string, error) {
	var email string
	err := s.db.QueryRowContext(ctx, `
		SELECT CONCAT(u.local_part, '@', d.name)
		FROM users u
		JOIN domains d ON d.id = u.primary_domain_id
		WHERE u.id = ?`, userID).Scan(&email)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return strings.ToLower(strings.TrimSpace(email)), err
}

func (s SQLStore) ListLoginRateLimits(ctx context.Context) ([]domain.LoginRateLimit, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, service, limiter_key, failure_count, first_failed_at, last_failed_at, locked_until, updated_at
		FROM login_rate_limits
		ORDER BY COALESCE(locked_until, last_failed_at, updated_at) DESC, failure_count DESC
		LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	limits := make([]domain.LoginRateLimit, 0)
	now := time.Now().UTC()
	for rows.Next() {
		var limit domain.LoginRateLimit
		var firstFailedAt, lastFailedAt, lockedUntil sql.NullTime
		if err := rows.Scan(&limit.ID, &limit.Service, &limit.LimiterKey, &limit.FailureCount, &firstFailedAt, &lastFailedAt, &lockedUntil, &limit.UpdatedAt); err != nil {
			return nil, err
		}
		limit.Scope, limit.Subject = parseLoginLimiterKey(limit.LimiterKey)
		if firstFailedAt.Valid {
			value := firstFailedAt.Time
			limit.FirstFailedAt = &value
		}
		if lastFailedAt.Valid {
			value := lastFailedAt.Time
			limit.LastFailedAt = &value
		}
		if lockedUntil.Valid {
			value := lockedUntil.Time
			limit.LockedUntil = &value
			limit.Locked = now.Before(value.UTC())
		}
		limits = append(limits, limit)
	}
	return limits, rows.Err()
}

func (s SQLStore) ClearLoginRateLimit(ctx context.Context, limitID uint64) error {
	_, err := execOne(ctx, s.db, `DELETE FROM login_rate_limits WHERE id = ?`, limitID)
	return err
}

func (s SQLStore) GetAdminMFASettings(ctx context.Context) (domain.AdminMFASettings, error) {
	var settings domain.AdminMFASettings
	var apiKey sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT id, local_totp_enabled, local_totp_secret, local_totp_pending_secret,
		       proidentity_enabled, proidentity_base_url, proidentity_api_key, proidentity_user_email,
		       proidentity_timeout_seconds, proidentity_totp_enabled, native_webauthn_enabled, updated_at
		FROM admin_mfa_settings
		WHERE id = 1`).Scan(
		&settings.ID,
		&settings.LocalTOTPEnabled,
		&settings.LocalTOTPSecret,
		&settings.LocalTOTPPendingSecret,
		&settings.ProIdentityEnabled,
		&settings.ProIdentityBaseURL,
		&apiKey,
		&settings.ProIdentityUserEmail,
		&settings.ProIdentityTimeoutSeconds,
		&settings.ProIdentityTOTPEnabled,
		&settings.NativeWebAuthnEnabled,
		&settings.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return domain.AdminMFASettings{ID: 1, ProIdentityBaseURL: proIdentityAuthServiceURL, ProIdentityTimeoutSeconds: defaultProIdentityTimeoutSeconds}, nil
	}
	if err != nil {
		return domain.AdminMFASettings{}, err
	}
	settings.ProIdentityAPIKey = apiKey.String
	settings.ProIdentityBaseURL = proIdentityAuthServiceURL
	return settings, nil
}

func (s SQLStore) SaveAdminMFASettings(ctx context.Context, settings domain.AdminMFASettings) (domain.AdminMFASettings, error) {
	if settings.ProIdentityTimeoutSeconds <= 0 {
		settings.ProIdentityTimeoutSeconds = defaultProIdentityTimeoutSeconds
	}
	settings.ProIdentityBaseURL = proIdentityAuthServiceURL
	_, err := execOne(ctx, s.db, `
		INSERT INTO admin_mfa_settings(
			id, local_totp_enabled, local_totp_secret, local_totp_pending_secret,
			proidentity_enabled, proidentity_base_url, proidentity_api_key, proidentity_user_email,
			proidentity_timeout_seconds, proidentity_totp_enabled, native_webauthn_enabled
		)
		VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			local_totp_enabled = VALUES(local_totp_enabled),
			local_totp_secret = VALUES(local_totp_secret),
			local_totp_pending_secret = VALUES(local_totp_pending_secret),
			proidentity_enabled = VALUES(proidentity_enabled),
			proidentity_base_url = VALUES(proidentity_base_url),
			proidentity_api_key = VALUES(proidentity_api_key),
			proidentity_user_email = VALUES(proidentity_user_email),
			proidentity_timeout_seconds = VALUES(proidentity_timeout_seconds),
			proidentity_totp_enabled = VALUES(proidentity_totp_enabled),
			native_webauthn_enabled = VALUES(native_webauthn_enabled)`,
		settings.LocalTOTPEnabled,
		strings.TrimSpace(settings.LocalTOTPSecret),
		strings.TrimSpace(settings.LocalTOTPPendingSecret),
		settings.ProIdentityEnabled,
		strings.TrimSpace(settings.ProIdentityBaseURL),
		strings.TrimSpace(settings.ProIdentityAPIKey),
		strings.ToLower(strings.TrimSpace(settings.ProIdentityUserEmail)),
		settings.ProIdentityTimeoutSeconds,
		settings.ProIdentityTOTPEnabled,
		settings.NativeWebAuthnEnabled,
	)
	if err != nil {
		return domain.AdminMFASettings{}, err
	}
	return s.GetAdminMFASettings(ctx)
}

func (s SQLStore) CreateAdminMFAChallenge(ctx context.Context, challenge domain.AdminMFAChallenge) error {
	_, _ = s.db.ExecContext(ctx, `DELETE FROM admin_mfa_challenges WHERE expires_at < UTC_TIMESTAMP()`)
	_, err := execOne(ctx, s.db, `
		INSERT INTO admin_mfa_challenges(token, username, provider, request_id, expires_at)
		VALUES (?, ?, ?, ?, ?)`,
		challenge.Token,
		challenge.Username,
		challenge.Provider,
		challenge.RequestID,
		challenge.ExpiresAt.UTC(),
	)
	return err
}

func (s SQLStore) GetAdminMFAChallenge(ctx context.Context, token string) (domain.AdminMFAChallenge, error) {
	var challenge domain.AdminMFAChallenge
	err := s.db.QueryRowContext(ctx, `
		SELECT token, username, provider, request_id, expires_at, created_at
		FROM admin_mfa_challenges
		WHERE token = ?`, strings.TrimSpace(token)).Scan(
		&challenge.Token,
		&challenge.Username,
		&challenge.Provider,
		&challenge.RequestID,
		&challenge.ExpiresAt,
		&challenge.CreatedAt,
	)
	return challenge, err
}

func (s SQLStore) DeleteAdminMFAChallenge(ctx context.Context, token string) error {
	_, err := execOne(ctx, s.db, `DELETE FROM admin_mfa_challenges WHERE token = ?`, strings.TrimSpace(token))
	return err
}

func (s SQLStore) ListAdminWebAuthnCredentials(ctx context.Context) ([]domain.AdminWebAuthnCredential, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, credential_id, name, credential_json, created_at, last_used_at
		FROM admin_webauthn_credentials
		ORDER BY created_at ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var credentials []domain.AdminWebAuthnCredential
	for rows.Next() {
		var credential domain.AdminWebAuthnCredential
		var lastUsedAt sql.NullTime
		if err := rows.Scan(&credential.ID, &credential.CredentialID, &credential.Name, &credential.CredentialJSON, &credential.CreatedAt, &lastUsedAt); err != nil {
			return nil, err
		}
		if lastUsedAt.Valid {
			value := lastUsedAt.Time
			credential.LastUsedAt = &value
		}
		credentials = append(credentials, credential)
	}
	return credentials, rows.Err()
}

func (s SQLStore) CreateAdminWebAuthnCredential(ctx context.Context, credential domain.AdminWebAuthnCredential) (domain.AdminWebAuthnCredential, error) {
	name := strings.TrimSpace(credential.Name)
	if name == "" {
		name = "Hardware key"
	}
	result, err := execOne(ctx, s.db, `
		INSERT INTO admin_webauthn_credentials(credential_id, name, credential_json)
		VALUES (?, ?, ?)`,
		credential.CredentialID,
		name,
		credential.CredentialJSON,
	)
	if err != nil {
		return domain.AdminWebAuthnCredential{}, err
	}
	_, _ = execOne(ctx, s.db, `UPDATE admin_mfa_settings SET native_webauthn_enabled = TRUE WHERE id = 1`)
	id, _ := result.LastInsertId()
	credential.ID = uint64(id)
	credential.Name = name
	return credential, nil
}

func (s SQLStore) UpdateAdminWebAuthnCredential(ctx context.Context, credential domain.AdminWebAuthnCredential) error {
	_, err := execOne(ctx, s.db, `
		UPDATE admin_webauthn_credentials
		SET credential_json = ?, last_used_at = UTC_TIMESTAMP()
		WHERE id = ?`,
		credential.CredentialJSON,
		credential.ID,
	)
	return err
}

func (s SQLStore) CreateAdminWebAuthnSession(ctx context.Context, session domain.AdminWebAuthnSession) error {
	_, _ = execOne(ctx, s.db, `DELETE FROM admin_webauthn_sessions WHERE expires_at < UTC_TIMESTAMP()`)
	_, err := execOne(ctx, s.db, `
		INSERT INTO admin_webauthn_sessions(token, ceremony, session_json, expires_at)
		VALUES (?, ?, ?, ?)`,
		strings.TrimSpace(session.Token),
		strings.TrimSpace(session.Ceremony),
		session.SessionJSON,
		session.ExpiresAt.UTC(),
	)
	return err
}

func (s SQLStore) GetAdminWebAuthnSession(ctx context.Context, token string) (domain.AdminWebAuthnSession, error) {
	var session domain.AdminWebAuthnSession
	err := s.db.QueryRowContext(ctx, `
		SELECT token, ceremony, session_json, expires_at, created_at
		FROM admin_webauthn_sessions
		WHERE token = ?`, strings.TrimSpace(token)).Scan(
		&session.Token,
		&session.Ceremony,
		&session.SessionJSON,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	return session, err
}

func (s SQLStore) DeleteAdminWebAuthnSession(ctx context.Context, token string) error {
	_, err := execOne(ctx, s.db, `DELETE FROM admin_webauthn_sessions WHERE token = ?`, strings.TrimSpace(token))
	return err
}

func parseLoginLimiterKey(key string) (string, string) {
	parts := strings.Split(strings.TrimSpace(key), "|")
	if len(parts) < 3 {
		return "unknown", key
	}
	switch parts[1] {
	case "ip":
		return "ip", parts[2]
	case "account":
		return "account", parts[2]
	case "domain":
		return "domain", parts[2]
	case "pair":
		if len(parts) >= 4 {
			return "ip + account", parts[2] + " from " + parts[3]
		}
	}
	return "unknown", key
}

func (s SQLStore) CreateAlias(ctx context.Context, alias domain.Alias) (domain.Alias, error) {
	result, err := s.db.ExecContext(ctx, `INSERT INTO aliases(tenant_id, domain_id, source_local_part, destination) VALUES (?, ?, ?, ?)`,
		alias.TenantID, alias.DomainID, strings.ToLower(strings.TrimSpace(alias.SourceLocalPart)), strings.ToLower(strings.TrimSpace(alias.Destination)))
	if err != nil {
		return domain.Alias{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Alias{}, err
	}
	alias.ID = uint64(id)
	return alias, nil
}

func (s SQLStore) ListAliases(ctx context.Context) ([]domain.Alias, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant_id, domain_id, source_local_part, destination, created_at
		FROM aliases
		ORDER BY created_at DESC, id DESC
		LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var aliases []domain.Alias
	for rows.Next() {
		var alias domain.Alias
		if err := rows.Scan(&alias.ID, &alias.TenantID, &alias.DomainID, &alias.SourceLocalPart, &alias.Destination, &alias.CreatedAt); err != nil {
			return nil, err
		}
		aliases = append(aliases, alias)
	}
	return aliases, rows.Err()
}

func (s SQLStore) UpdateAlias(ctx context.Context, alias domain.Alias) (domain.Alias, error) {
	if _, err := execOne(ctx, s.db, `UPDATE aliases SET tenant_id = ?, domain_id = ?, source_local_part = ?, destination = ? WHERE id = ?`,
		alias.TenantID, alias.DomainID, strings.ToLower(strings.TrimSpace(alias.SourceLocalPart)), strings.ToLower(strings.TrimSpace(alias.Destination)), alias.ID); err != nil {
		return domain.Alias{}, err
	}
	return scanAlias(ctx, s.db, alias.ID)
}

func (s SQLStore) DeleteAlias(ctx context.Context, aliasID uint64) error {
	_, err := execOne(ctx, s.db, `DELETE FROM aliases WHERE id = ?`, aliasID)
	return err
}

func (s SQLStore) CreateCatchAllRoute(ctx context.Context, route domain.CatchAllRoute) (domain.CatchAllRoute, error) {
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO catch_all_routes(tenant_id, domain_id, destination, status) VALUES (?, ?, ?, 'active')
		ON DUPLICATE KEY UPDATE destination = VALUES(destination), tenant_id = VALUES(tenant_id), status = 'active'`,
		route.TenantID, route.DomainID, strings.ToLower(strings.TrimSpace(route.Destination)))
	if err != nil {
		return domain.CatchAllRoute{}, err
	}
	id, _ := result.LastInsertId()
	if id > 0 {
		route.ID = uint64(id)
	}
	route.Status = "active"
	return route, nil
}

func (s SQLStore) ListCatchAllRoutes(ctx context.Context) ([]domain.CatchAllRoute, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant_id, domain_id, destination, status, created_at, updated_at
		FROM catch_all_routes
		ORDER BY created_at DESC, id DESC
		LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var routes []domain.CatchAllRoute
	for rows.Next() {
		var route domain.CatchAllRoute
		if err := rows.Scan(&route.ID, &route.TenantID, &route.DomainID, &route.Destination, &route.Status, &route.CreatedAt, &route.UpdatedAt); err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}
	return routes, rows.Err()
}

func (s SQLStore) UpdateCatchAllRoute(ctx context.Context, route domain.CatchAllRoute) (domain.CatchAllRoute, error) {
	if _, err := execOne(ctx, s.db, `
		UPDATE catch_all_routes
		SET tenant_id = ?, domain_id = ?, destination = ?, status = ?
		WHERE id = ?`,
		route.TenantID, route.DomainID, strings.ToLower(strings.TrimSpace(route.Destination)), route.Status, route.ID); err != nil {
		return domain.CatchAllRoute{}, err
	}
	return scanCatchAllRoute(ctx, s.db, route.ID)
}

func (s SQLStore) DeleteCatchAllRoute(ctx context.Context, routeID uint64) error {
	_, err := execOne(ctx, s.db, `DELETE FROM catch_all_routes WHERE id = ?`, routeID)
	return err
}

func (s SQLStore) CreateSharedMailboxPermission(ctx context.Context, permission domain.SharedMailboxPermission) (domain.SharedMailboxPermission, error) {
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO shared_mailbox_permissions(tenant_id, shared_mailbox_id, user_id, can_read, can_send_as, can_send_on_behalf, can_manage)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE can_read = VALUES(can_read), can_send_as = VALUES(can_send_as), can_send_on_behalf = VALUES(can_send_on_behalf), can_manage = VALUES(can_manage)`,
		permission.TenantID, permission.SharedMailboxID, permission.UserID, permission.CanRead, permission.CanSendAs, permission.CanSendOnBehalf, permission.CanManage)
	if err != nil {
		return domain.SharedMailboxPermission{}, err
	}
	id, _ := result.LastInsertId()
	if id > 0 {
		permission.ID = uint64(id)
	}
	return permission, nil
}

func (s SQLStore) ListSharedMailboxPermissions(ctx context.Context) ([]domain.SharedMailboxPermission, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant_id, shared_mailbox_id, user_id, can_read, can_send_as, can_send_on_behalf, can_manage, created_at, updated_at
		FROM shared_mailbox_permissions
		ORDER BY created_at DESC, id DESC
		LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var permissions []domain.SharedMailboxPermission
	for rows.Next() {
		var permission domain.SharedMailboxPermission
		if err := rows.Scan(&permission.ID, &permission.TenantID, &permission.SharedMailboxID, &permission.UserID, &permission.CanRead, &permission.CanSendAs, &permission.CanSendOnBehalf, &permission.CanManage, &permission.CreatedAt, &permission.UpdatedAt); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, rows.Err()
}

func (s SQLStore) UpdateSharedMailboxPermission(ctx context.Context, permission domain.SharedMailboxPermission) (domain.SharedMailboxPermission, error) {
	if _, err := execOne(ctx, s.db, `
		UPDATE shared_mailbox_permissions
		SET tenant_id = ?, shared_mailbox_id = ?, user_id = ?, can_read = ?, can_send_as = ?, can_send_on_behalf = ?, can_manage = ?
		WHERE id = ?`,
		permission.TenantID, permission.SharedMailboxID, permission.UserID, permission.CanRead, permission.CanSendAs, permission.CanSendOnBehalf, permission.CanManage, permission.ID); err != nil {
		return domain.SharedMailboxPermission{}, err
	}
	return scanSharedMailboxPermission(ctx, s.db, permission.ID)
}

func (s SQLStore) DeleteSharedMailboxPermission(ctx context.Context, permissionID uint64) error {
	_, err := execOne(ctx, s.db, `DELETE FROM shared_mailbox_permissions WHERE id = ?`, permissionID)
	return err
}

func (s SQLStore) ListQuarantineEvents(ctx context.Context) ([]domain.QuarantineEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant_id, user_id, domain_id, message_id, sender, recipient, verdict, action, scanner, CAST(symbols_json AS CHAR), storage_path, size_bytes, sha256, status, resolved_at, resolution_note, created_at
		FROM quarantine_events
		ORDER BY created_at DESC, id DESC
		LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]domain.QuarantineEvent, 0)
	for rows.Next() {
		var event domain.QuarantineEvent
		var userID sql.NullInt64
		var domainID sql.NullInt64
		var messageID sql.NullString
		var sender sql.NullString
		var storagePath sql.NullString
		var sha256Value sql.NullString
		var resolvedAt sql.NullTime
		var resolutionNote sql.NullString
		if err := rows.Scan(&event.ID, &event.TenantID, &userID, &domainID, &messageID, &sender, &event.Recipient, &event.Verdict, &event.Action, &event.Scanner, &event.SymbolsJSON, &storagePath, &event.SizeBytes, &sha256Value, &event.Status, &resolvedAt, &resolutionNote, &event.CreatedAt); err != nil {
			return nil, err
		}
		if userID.Valid {
			id := uint64(userID.Int64)
			event.UserID = &id
		}
		if domainID.Valid {
			id := uint64(domainID.Int64)
			event.DomainID = &id
		}
		if messageID.Valid {
			event.MessageID = messageID.String
		}
		if sender.Valid {
			event.Sender = sender.String
		}
		if storagePath.Valid {
			event.StoragePath = storagePath.String
		}
		if sha256Value.Valid {
			event.SHA256 = sha256Value.String
		}
		if resolvedAt.Valid {
			event.ResolvedAt = &resolvedAt.Time
		}
		if resolutionNote.Valid {
			event.ResolutionNote = resolutionNote.String
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

type QuarantineMessageInput struct {
	Recipient   string
	Sender      string
	MessageID   string
	Verdict     string
	Action      string
	Scanner     string
	SymbolsJSON string
	Reader      io.Reader
}

func (s SQLStore) StoreQuarantineMessage(ctx context.Context, input QuarantineMessageInput) (domain.QuarantineEvent, error) {
	action := strings.TrimSpace(input.Action)
	if action == "" {
		action = "quarantine"
	}
	symbolsJSON := strings.TrimSpace(input.SymbolsJSON)
	if symbolsJSON == "" {
		symbolsJSON = `{}`
	}
	var tenantID uint64
	var userID sql.NullInt64
	var domainID sql.NullInt64
	err := s.db.QueryRowContext(ctx, `
		SELECT d.tenant_id, u.id, d.id
		FROM domains d
		LEFT JOIN users u ON u.primary_domain_id = d.id AND CONCAT(u.local_part, '@', d.name) = ?
		WHERE d.name = SUBSTRING_INDEX(?, '@', -1)
		  AND d.status IN ('pending', 'active')
		LIMIT 1`, strings.ToLower(input.Recipient), strings.ToLower(input.Recipient)).Scan(&tenantID, &userID, &domainID)
	if err != nil {
		return domain.QuarantineEvent{}, err
	}
	stored, err := s.quarantine.StoreMessage(ctx, quarantine.StoreRequest{TenantID: tenantID, Recipient: input.Recipient, MessageID: input.MessageID, Reader: input.Reader})
	if err != nil {
		return domain.QuarantineEvent{}, err
	}
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO quarantine_events(tenant_id, user_id, domain_id, message_id, sender, recipient, verdict, action, scanner, symbols_json, storage_path, size_bytes, sha256)
		VALUES (?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?)`,
		tenantID,
		nullInt64Arg(userID),
		nullInt64Arg(domainID),
		input.MessageID,
		input.Sender,
		strings.ToLower(input.Recipient),
		input.Verdict,
		action,
		input.Scanner,
		symbolsJSON,
		stored.StoragePath,
		stored.SizeBytes,
		stored.SHA256,
	)
	if err != nil {
		_ = s.quarantine.Delete(ctx, stored.StoragePath)
		return domain.QuarantineEvent{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.QuarantineEvent{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO audit_events(tenant_id, actor_type, action, target_type, target_id, metadata_json)
		VALUES (?, 'system', 'quarantine.created', 'quarantine_event', ?, JSON_OBJECT('recipient', ?, 'sender', ?, 'verdict', ?, 'status', 'held', 'scanner', ?, 'size_bytes', ?))`,
		tenantID,
		strconv.FormatInt(id, 10),
		strings.ToLower(input.Recipient),
		input.Sender,
		input.Verdict,
		input.Scanner,
		stored.SizeBytes,
	); err != nil {
		return domain.QuarantineEvent{}, err
	}
	return scanQuarantineEvent(ctx, s.db, uint64(id))
}

func (s SQLStore) ResolveQuarantineEvent(ctx context.Context, eventID uint64, status, note string) (domain.QuarantineEvent, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.QuarantineEvent{}, err
	}
	defer tx.Rollback()

	event, err := scanQuarantineEvent(ctx, tx, eventID)
	if err != nil {
		return domain.QuarantineEvent{}, err
	}
	if event.Status != "held" {
		return domain.QuarantineEvent{}, sql.ErrNoRows
	}
	if status == "released" && event.StoragePath != "" {
		if _, err := s.quarantine.Release(ctx, quarantine.ReleaseRequest{Recipient: event.Recipient, MessageID: event.MessageID, StoragePath: event.StoragePath, QuarantineID: event.ID}); err != nil {
			return domain.QuarantineEvent{}, err
		}
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE quarantine_events
		SET status = ?, resolved_at = CURRENT_TIMESTAMP, resolution_note = NULLIF(?, '')
		WHERE id = ?
		  AND status = 'held'`, status, note, eventID)
	if err != nil {
		return domain.QuarantineEvent{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return domain.QuarantineEvent{}, err
	}
	if affected == 0 {
		return domain.QuarantineEvent{}, sql.ErrNoRows
	}

	event, err = scanQuarantineEvent(ctx, tx, eventID)
	if err != nil {
		return domain.QuarantineEvent{}, err
	}
	if status == "deleted" && event.StoragePath != "" {
		if err := s.quarantine.Delete(ctx, event.StoragePath); err != nil {
			return domain.QuarantineEvent{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO audit_events(tenant_id, actor_type, action, target_type, target_id, metadata_json)
		VALUES (?, 'admin', ?, 'quarantine_event', ?, JSON_OBJECT('status', ?, 'recipient', ?, 'note', ?))`,
		event.TenantID,
		"quarantine."+status,
		strconv.FormatUint(eventID, 10),
		status,
		event.Recipient,
		note,
	); err != nil {
		return domain.QuarantineEvent{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.QuarantineEvent{}, err
	}
	return event, nil
}

type quarantineQuerier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func scanQuarantineEvent(ctx context.Context, q quarantineQuerier, eventID uint64) (domain.QuarantineEvent, error) {
	var event domain.QuarantineEvent
	var userID sql.NullInt64
	var domainID sql.NullInt64
	var messageID sql.NullString
	var sender sql.NullString
	var storagePath sql.NullString
	var sha256Value sql.NullString
	var resolvedAt sql.NullTime
	var resolutionNote sql.NullString
	err := q.QueryRowContext(ctx, `
		SELECT id, tenant_id, user_id, domain_id, message_id, sender, recipient, verdict, action, scanner, CAST(symbols_json AS CHAR), storage_path, size_bytes, sha256, status, resolved_at, resolution_note, created_at
		FROM quarantine_events
		WHERE id = ?`, eventID).Scan(&event.ID, &event.TenantID, &userID, &domainID, &messageID, &sender, &event.Recipient, &event.Verdict, &event.Action, &event.Scanner, &event.SymbolsJSON, &storagePath, &event.SizeBytes, &sha256Value, &event.Status, &resolvedAt, &resolutionNote, &event.CreatedAt)
	if err != nil {
		return domain.QuarantineEvent{}, err
	}
	if userID.Valid {
		id := uint64(userID.Int64)
		event.UserID = &id
	}
	if domainID.Valid {
		id := uint64(domainID.Int64)
		event.DomainID = &id
	}
	if messageID.Valid {
		event.MessageID = messageID.String
	}
	if sender.Valid {
		event.Sender = sender.String
	}
	if storagePath.Valid {
		event.StoragePath = storagePath.String
	}
	if sha256Value.Valid {
		event.SHA256 = sha256Value.String
	}
	if resolvedAt.Valid {
		event.ResolvedAt = &resolvedAt.Time
	}
	if resolutionNote.Valid {
		event.ResolutionNote = resolutionNote.String
	}
	return event, nil
}

func nullInt64Arg(value sql.NullInt64) any {
	if value.Valid {
		return value.Int64
	}
	return nil
}

func cleanUintPtr(value *uint64) *uint64 {
	if value == nil || *value == 0 {
		return nil
	}
	return value
}

func uintPtrArg(value *uint64) any {
	if value == nil || *value == 0 {
		return nil
	}
	return *value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

type execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type rowQuerier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func execOne(ctx context.Context, e execer, query string, args ...any) (sql.Result, error) {
	return e.ExecContext(ctx, query, args...)
}

func scanTenant(ctx context.Context, q rowQuerier, tenantID uint64) (domain.Tenant, error) {
	var tenant domain.Tenant
	err := q.QueryRowContext(ctx, `
		SELECT id, name, slug, status, created_at, updated_at
		FROM tenants
		WHERE id = ?`, tenantID).Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt)
	return tenant, err
}

func scanDomain(ctx context.Context, q rowQuerier, domainID uint64) (domain.Domain, error) {
	var mailDomain domain.Domain
	err := q.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, status, dkim_selector, created_at, updated_at
		FROM domains
		WHERE id = ?`, domainID).Scan(&mailDomain.ID, &mailDomain.TenantID, &mailDomain.Name, &mailDomain.Status, &mailDomain.DKIMSelector, &mailDomain.CreatedAt, &mailDomain.UpdatedAt)
	return mailDomain, err
}

func scanUser(ctx context.Context, q rowQuerier, userID uint64) (domain.User, error) {
	var user domain.User
	err := q.QueryRowContext(ctx, `
		SELECT id, tenant_id, primary_domain_id, local_part, display_name, mailbox_type, status, quota_bytes, created_at, updated_at
		FROM users
		WHERE id = ?`, userID).Scan(&user.ID, &user.TenantID, &user.PrimaryDomainID, &user.LocalPart, &user.DisplayName, &user.MailboxType, &user.Status, &user.QuotaBytes, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

func scanTenantAdmin(ctx context.Context, q rowQuerier, adminID uint64) (domain.TenantAdmin, error) {
	var admin domain.TenantAdmin
	err := q.QueryRowContext(ctx, `
		SELECT id, tenant_id, user_id, role, status, created_at, updated_at
		FROM tenant_admins
		WHERE id = ?`, adminID).Scan(&admin.ID, &admin.TenantID, &admin.UserID, &admin.Role, &admin.Status, &admin.CreatedAt, &admin.UpdatedAt)
	return admin, err
}

func normalizeTenantAdminRole(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "read_only":
		return "read_only"
	default:
		return "tenant_admin"
	}
}

func normalizeTenantAdminStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "disabled":
		return "disabled"
	default:
		return "active"
	}
}

func scanAlias(ctx context.Context, q rowQuerier, aliasID uint64) (domain.Alias, error) {
	var alias domain.Alias
	err := q.QueryRowContext(ctx, `
		SELECT id, tenant_id, domain_id, source_local_part, destination, created_at
		FROM aliases
		WHERE id = ?`, aliasID).Scan(&alias.ID, &alias.TenantID, &alias.DomainID, &alias.SourceLocalPart, &alias.Destination, &alias.CreatedAt)
	return alias, err
}

func scanCatchAllRoute(ctx context.Context, q rowQuerier, routeID uint64) (domain.CatchAllRoute, error) {
	var route domain.CatchAllRoute
	err := q.QueryRowContext(ctx, `
		SELECT id, tenant_id, domain_id, destination, status, created_at, updated_at
		FROM catch_all_routes
		WHERE id = ?`, routeID).Scan(&route.ID, &route.TenantID, &route.DomainID, &route.Destination, &route.Status, &route.CreatedAt, &route.UpdatedAt)
	return route, err
}

func scanSharedMailboxPermission(ctx context.Context, q rowQuerier, permissionID uint64) (domain.SharedMailboxPermission, error) {
	var permission domain.SharedMailboxPermission
	err := q.QueryRowContext(ctx, `
		SELECT id, tenant_id, shared_mailbox_id, user_id, can_read, can_send_as, can_send_on_behalf, can_manage, created_at, updated_at
		FROM shared_mailbox_permissions
		WHERE id = ?`, permissionID).Scan(&permission.ID, &permission.TenantID, &permission.SharedMailboxID, &permission.UserID, &permission.CanRead, &permission.CanSendAs, &permission.CanSendOnBehalf, &permission.CanManage, &permission.CreatedAt, &permission.UpdatedAt)
	return permission, err
}

func (s SQLStore) ListAuditEvents(ctx context.Context) ([]domain.AuditEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant_id, actor_type, actor_id, action, target_type, target_id, CAST(metadata_json AS CHAR), created_at
		FROM audit_events
		ORDER BY created_at DESC, id DESC
		LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]domain.AuditEvent, 0)
	for rows.Next() {
		var event domain.AuditEvent
		var tenantID sql.NullInt64
		var actorID sql.NullInt64
		if err := rows.Scan(&event.ID, &tenantID, &event.ActorType, &actorID, &event.Action, &event.TargetType, &event.TargetID, &event.MetadataJSON, &event.CreatedAt); err != nil {
			return nil, err
		}
		if tenantID.Valid {
			id := uint64(tenantID.Int64)
			event.TenantID = &id
		}
		if actorID.Valid {
			id := uint64(actorID.Int64)
			event.ActorID = &id
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s SQLStore) RecordAuditEvent(ctx context.Context, event domain.AuditEvent) error {
	var tenantID any
	if event.TenantID != nil {
		tenantID = *event.TenantID
	}
	var actorID any
	if event.ActorID != nil {
		actorID = *event.ActorID
	}
	metadata := event.MetadataJSON
	if metadata == "" {
		metadata = `{}`
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO audit_events(tenant_id, actor_type, actor_id, action, target_type, target_id, metadata_json)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		tenantID,
		event.ActorType,
		actorID,
		event.Action,
		event.TargetType,
		event.TargetID,
		metadata,
	)
	return err
}

func (s SQLStore) ListTenantPolicies(ctx context.Context) ([]domain.TenantPolicy, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT tenant_id, spam_action, malware_action, require_tls_for_auth, created_at, updated_at
		FROM tenant_policies
		ORDER BY tenant_id
		LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	policies := make([]domain.TenantPolicy, 0)
	for rows.Next() {
		var policy domain.TenantPolicy
		if err := rows.Scan(&policy.TenantID, &policy.SpamAction, &policy.MalwareAction, &policy.RequireTLSForAuth, &policy.CreatedAt, &policy.UpdatedAt); err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}
	return policies, rows.Err()
}

func (s SQLStore) UpdateTenantPolicy(ctx context.Context, policy domain.TenantPolicy) (domain.TenantPolicy, error) {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO tenant_policies(tenant_id, spam_action, malware_action, require_tls_for_auth)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE spam_action = VALUES(spam_action), malware_action = VALUES(malware_action), require_tls_for_auth = VALUES(require_tls_for_auth)`,
		policy.TenantID,
		policy.SpamAction,
		policy.MalwareAction,
		policy.RequireTLSForAuth,
	)
	if err != nil {
		return domain.TenantPolicy{}, err
	}
	return policy, nil
}

func (s SQLStore) GetMailServerSettings(ctx context.Context) (domain.MailServerSettings, error) {
	settings, err := s.scanMailServerSettings(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		settings = defaultMailServerSettings(s.dns)
	} else if err != nil {
		return domain.MailServerSettings{}, err
	}
	return s.withEffectiveMailHostname(ctx, settings)
}

func (s SQLStore) UpdateMailServerSettings(ctx context.Context, settings domain.MailServerSettings) (domain.MailServerSettings, error) {
	settings.HostnameMode = normalizeHostnameMode(settings.HostnameMode)
	settings.MailHostname = normalizeDNSName(settings.MailHostname)
	settings.PublicIPv4 = strings.TrimSpace(settings.PublicIPv4)
	settings.PublicIPv6 = strings.TrimSpace(settings.PublicIPv6)
	settings.HeadTenantID = cleanUintPtr(settings.HeadTenantID)
	settings.HeadDomainID = cleanUintPtr(settings.HeadDomainID)
	settings.TLSMode = normalizeProxyTLSMode(settings.TLSMode)
	settings.DefaultLanguage = normalizeLanguageOrDefault(settings.DefaultLanguage)
	if settings.HostnameMode == "shared" && settings.MailHostname == "" {
		settings.MailHostname = effectiveMailHost("", s.dns.MailHostname)
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO mail_server_settings(id, hostname_mode, mail_hostname, head_tenant_id, head_domain_id, public_ipv4, public_ipv6, sni_enabled, tls_mode, force_https, default_language, mailbox_mfa_enabled, force_mailbox_mfa, cloudflare_real_ip_enabled)
		VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE hostname_mode = VALUES(hostname_mode), mail_hostname = VALUES(mail_hostname), head_tenant_id = VALUES(head_tenant_id), head_domain_id = VALUES(head_domain_id), public_ipv4 = VALUES(public_ipv4), public_ipv6 = VALUES(public_ipv6), sni_enabled = VALUES(sni_enabled), tls_mode = VALUES(tls_mode), force_https = VALUES(force_https), default_language = VALUES(default_language), mailbox_mfa_enabled = VALUES(mailbox_mfa_enabled), force_mailbox_mfa = VALUES(force_mailbox_mfa), cloudflare_real_ip_enabled = VALUES(cloudflare_real_ip_enabled)`,
		settings.HostnameMode,
		settings.MailHostname,
		uintPtrArg(settings.HeadTenantID),
		uintPtrArg(settings.HeadDomainID),
		settings.PublicIPv4,
		settings.PublicIPv6,
		settings.SNIEnabled,
		settings.TLSMode,
		settings.ForceHTTPS,
		settings.DefaultLanguage,
		settings.MailboxMFAEnabled,
		settings.ForceMailboxMFA,
		settings.CloudflareRealIPEnabled,
	)
	if err != nil {
		return domain.MailServerSettings{}, err
	}
	return s.GetMailServerSettings(ctx)
}

func (s SQLStore) scanMailServerSettings(ctx context.Context) (domain.MailServerSettings, error) {
	var settings domain.MailServerSettings
	var headTenant sql.NullInt64
	var headDomain sql.NullInt64
	err := s.db.QueryRowContext(ctx, `
		SELECT hostname_mode, mail_hostname, head_tenant_id, head_domain_id, public_ipv4, public_ipv6, sni_enabled, COALESCE(tls_mode, 'system'), COALESCE(force_https, 1), COALESCE(default_language, 'en'), COALESCE(mailbox_mfa_enabled, 1), COALESCE(force_mailbox_mfa, 0), COALESCE(cloudflare_real_ip_enabled, 0)
		FROM mail_server_settings
		WHERE id = 1`).Scan(&settings.HostnameMode, &settings.MailHostname, &headTenant, &headDomain, &settings.PublicIPv4, &settings.PublicIPv6, &settings.SNIEnabled, &settings.TLSMode, &settings.ForceHTTPS, &settings.DefaultLanguage, &settings.MailboxMFAEnabled, &settings.ForceMailboxMFA, &settings.CloudflareRealIPEnabled)
	if err != nil {
		return domain.MailServerSettings{}, err
	}
	if headTenant.Valid {
		id := uint64(headTenant.Int64)
		settings.HeadTenantID = &id
	}
	if headDomain.Valid {
		id := uint64(headDomain.Int64)
		settings.HeadDomainID = &id
	}
	settings.HostnameMode = normalizeHostnameMode(settings.HostnameMode)
	settings.MailHostname = normalizeDNSName(settings.MailHostname)
	settings.TLSMode = normalizeProxyTLSMode(settings.TLSMode)
	settings.DefaultLanguage = normalizeLanguageOrDefault(settings.DefaultLanguage)
	return settings, nil
}

func defaultMailServerSettings(settings DNSSettings) domain.MailServerSettings {
	mode := normalizeHostnameMode(settings.HostnameMode)
	mailHost := effectiveMailHost("", settings.MailHostname)
	out := domain.MailServerSettings{
		HostnameMode:      mode,
		MailHostname:      mailHost,
		PublicIPv4:        strings.TrimSpace(settings.PublicIPv4),
		PublicIPv6:        strings.TrimSpace(settings.PublicIPv6),
		SNIEnabled:        settings.SNIEnabled,
		TLSMode:           normalizeProxyTLSMode(settings.TLSMode),
		ForceHTTPS:        settings.ForceHTTPS,
		DefaultLanguage:   normalizeLanguageOrDefault("en"),
		MailboxMFAEnabled: true,
	}
	if settings.HeadTenantID != 0 {
		out.HeadTenantID = &settings.HeadTenantID
	}
	if settings.HeadDomainID != 0 {
		out.HeadDomainID = &settings.HeadDomainID
	}
	return out
}

func (s SQLStore) withEffectiveMailHostname(ctx context.Context, settings domain.MailServerSettings) (domain.MailServerSettings, error) {
	settings.HostnameMode = normalizeHostnameMode(settings.HostnameMode)
	if settings.PublicIPv4 == "" {
		settings.PublicIPv4 = strings.TrimSpace(s.dns.PublicIPv4)
	}
	if settings.PublicIPv6 == "" {
		settings.PublicIPv6 = strings.TrimSpace(s.dns.PublicIPv6)
	}
	if settings.HostnameMode == "head-domain" && settings.HeadDomainID != nil && *settings.HeadDomainID != 0 {
		name, err := s.domainName(ctx, *settings.HeadDomainID)
		if err != nil {
			return domain.MailServerSettings{}, err
		}
		settings.EffectiveHostname = "mail." + name
		settings.MailHostname = settings.EffectiveHostname
		return settings, nil
	}
	if settings.HostnameMode == "per-domain" {
		settings.EffectiveHostname = "mail.<domain>"
		return settings, nil
	}
	settings.MailHostname = effectiveMailHost("", firstNonEmpty(settings.MailHostname, s.dns.MailHostname))
	settings.EffectiveHostname = settings.MailHostname
	return settings, nil
}

func (s SQLStore) domainName(ctx context.Context, domainID uint64) (string, error) {
	var name string
	err := s.db.QueryRowContext(ctx, `SELECT name FROM domains WHERE id = ?`, domainID).Scan(&name)
	if err != nil {
		return "", err
	}
	return normalizeDNSName(name), nil
}

func (s SQLStore) GetDomainDNS(ctx context.Context, domainID uint64) (domain.DomainDNS, error) {
	var domainName string
	var selector sql.NullString
	var dkimTXT sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT d.name, k.selector, k.public_dns_txt
		FROM domains d
		LEFT JOIN dkim_keys k ON k.domain_id = d.id AND k.status = 'active'
		WHERE d.id = ?
		ORDER BY k.created_at DESC
		LIMIT 1
	`, domainID).Scan(&domainName, &selector, &dkimTXT)
	if err != nil {
		return domain.DomainDNS{}, err
	}
	selectorValue := ""
	if selector.Valid {
		selectorValue = selector.String
	}
	dkimValue := ""
	if dkimTXT.Valid {
		dkimValue = dkimTXT.String
	}
	settings, err := s.GetMailServerSettings(ctx)
	if err != nil {
		return domain.DomainDNS{}, err
	}
	dnsSettings := dnsSettingsFromMailServer(settings, s.dns)
	tlsSettings, err := s.GetDomainTLSSettings(ctx, domainID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return domain.DomainDNS{}, err
	}
	dnsSettings.DisableWebmailAlias = !tlsSettings.DNSWebmailAliasEnabled
	dnsSettings.DisableAdminAlias = !tlsSettings.DNSAdminAliasEnabled
	return buildDomainDNS(domainID, domainName, selectorValue, dkimValue, dnsSettings), nil
}

func dnsSettingsFromMailServer(settings domain.MailServerSettings, fallback DNSSettings) DNSSettings {
	return DNSSettings{
		MailHostname:    settings.EffectiveHostname,
		AdminHostname:   fallback.AdminHostname,
		WebmailHostname: fallback.WebmailHostname,
		PublicIPv4:      firstNonEmpty(settings.PublicIPv4, fallback.PublicIPv4),
		PublicIPv6:      firstNonEmpty(settings.PublicIPv6, fallback.PublicIPv6),
		HostnameMode:    normalizeHostnameMode(settings.HostnameMode),
		SNIEnabled:      settings.SNIEnabled,
	}
}

func buildDomainDNS(domainID uint64, domainName, selector, dkimTXT string, settings DNSSettings) domain.DomainDNS {
	domainName = normalizeDNSName(domainName)
	mailHost := mailHostForDomain(domainName, settings)
	domainMailAlias := "mail." + domainName
	warnings := make([]string, 0)
	priority := 10

	records := make([]domain.DNSRecord, 0, 20)
	if hostBelongsToDomain(mailHost, domainName) {
		if strings.TrimSpace(settings.PublicIPv4) != "" {
			records = append(records, domain.DNSRecord{Type: "A", Name: mailHost, Value: strings.TrimSpace(settings.PublicIPv4), Purpose: "Mail server IPv4 address", Required: true})
		}
		if strings.TrimSpace(settings.PublicIPv6) != "" {
			records = append(records, domain.DNSRecord{Type: "AAAA", Name: mailHost, Value: strings.TrimSpace(settings.PublicIPv6), Purpose: "Mail server IPv6 address", Required: false})
		}
		if strings.TrimSpace(settings.PublicIPv4) == "" && strings.TrimSpace(settings.PublicIPv6) == "" {
			warnings = append(warnings, "Mail host "+mailHost+" is inside this DNS zone, but no public IPv4/IPv6 is configured. Set PROIDENTITY_PUBLIC_IPV4/PROIDENTITY_PUBLIC_IPV6 or use a shared public mail hostname.")
		}
	} else if domainMailAlias != mailHost {
		records = append(records, domain.DNSRecord{Type: "CNAME", Name: domainMailAlias, Value: mailHost, Purpose: "Domain mail client alias. Keep DNS-only; do not proxy mail records.", Required: false})
	}
	if !settings.DisableWebmailAlias {
		records = appendWebServiceAlias(records, "webmail."+domainName, settings.WebmailHostname, "Webmail browser alias", settings)
	}
	if !settings.DisableAdminAlias {
		records = appendWebServiceAlias(records, "madmin."+domainName, settings.AdminHostname, "Admin panel browser alias", settings)
	}

	records = append(records, []domain.DNSRecord{
		{Type: "MX", Name: domainName, Value: mailHost, Priority: &priority, Purpose: "Inbound mail delivery", Required: true},
		{Type: "TXT", Name: domainName, Value: "v=spf1 mx -all", Purpose: "SPF sender authorization", Required: true},
		{Type: "TXT", Name: "_dmarc." + domainName, Value: "v=DMARC1; p=quarantine; rua=mailto:dmarc@" + domainName, Purpose: "DMARC reporting and policy", Required: true},
		{Type: "TXT", Name: "_mta-sts." + domainName, Value: "v=STSv1; id=2026050601", Purpose: "MTA-STS policy advertisement", Required: false},
		{Type: "TXT", Name: "_smtp._tls." + domainName, Value: "v=TLSRPTv1; rua=mailto:tlsrpt@" + domainName, Purpose: "SMTP TLS reporting", Required: false},
		{Type: "CNAME", Name: "autoconfig." + domainName, Value: mailHost, Purpose: "Thunderbird automatic account setup", Required: false},
		{Type: "CNAME", Name: "autodiscover." + domainName, Value: mailHost, Purpose: "Outlook automatic account setup", Required: false},
		{Type: "SRV", Name: "_submission._tcp." + domainName, Value: "0 1 587 " + mailHost, Purpose: "SMTP submission discovery", Required: false},
		{Type: "SRV", Name: "_imaps._tcp." + domainName, Value: "0 1 993 " + mailHost, Purpose: "IMAP over TLS discovery", Required: false},
		{Type: "SRV", Name: "_pop3s._tcp." + domainName, Value: "0 1 995 " + mailHost, Purpose: "POP3 over TLS discovery", Required: false},
		{Type: "SRV", Name: "_sieve._tcp." + domainName, Value: "0 1 4190 " + mailHost, Purpose: "ManageSieve discovery", Required: false},
		{Type: "SRV", Name: "_caldavs._tcp." + domainName, Value: "0 1 443 " + mailHost, Purpose: "CalDAV discovery", Required: false},
		{Type: "SRV", Name: "_carddavs._tcp." + domainName, Value: "0 1 443 " + mailHost, Purpose: "CardDAV discovery", Required: false},
	}...)
	if selector != "" && dkimTXT != "" {
		records = append(records, domain.DNSRecord{
			Type:     "TXT",
			Name:     fmt.Sprintf("%s._domainkey.%s", selector, domainName),
			Value:    normalizeDKIMTXT(dkimTXT),
			Purpose:  "DKIM message signing",
			Required: true,
		})
	}
	return domain.DomainDNS{DomainID: domainID, Domain: domainName, MailHost: mailHost, Records: records, ClientSetup: clientSetup(domainName, mailHost), Warnings: warnings, Provisionable: len(warnings) == 0}
}

func effectiveMailHost(domainName, configured string) string {
	configured = normalizeDNSName(configured)
	if configured == "" || strings.HasSuffix(configured, ".local") || !strings.Contains(configured, ".") {
		if domainName == "" {
			return "mail.local"
		}
		return "mail." + domainName
	}
	return configured
}

func mailHostForDomain(domainName string, settings DNSSettings) string {
	mode := normalizeHostnameMode(settings.HostnameMode)
	if mode == "per-domain" {
		return "mail." + domainName
	}
	return effectiveMailHost(domainName, settings.MailHostname)
}

func appendWebServiceAlias(records []domain.DNSRecord, aliasName, targetHost, purpose string, settings DNSSettings) []domain.DNSRecord {
	aliasName = normalizeDNSName(aliasName)
	targetHost = normalizeDNSName(targetHost)
	if aliasName == "" || targetHost == "" || strings.HasSuffix(targetHost, ".local") || !strings.Contains(targetHost, ".") {
		return records
	}
	if aliasName == targetHost {
		if strings.TrimSpace(settings.PublicIPv4) != "" {
			records = append(records, domain.DNSRecord{Type: "A", Name: aliasName, Value: strings.TrimSpace(settings.PublicIPv4), Purpose: purpose, Required: false, Proxied: boolPtr(true)})
		}
		if strings.TrimSpace(settings.PublicIPv6) != "" {
			records = append(records, domain.DNSRecord{Type: "AAAA", Name: aliasName, Value: strings.TrimSpace(settings.PublicIPv6), Purpose: purpose, Required: false, Proxied: boolPtr(true)})
		}
		return records
	}
	return append(records, domain.DNSRecord{Type: "CNAME", Name: aliasName, Value: targetHost, Purpose: purpose, Required: false, Proxied: boolPtr(true)})
}

func (s SQLStore) GetDomainTLS(ctx context.Context, domainID uint64) (domain.DomainTLS, error) {
	mailDomain, err := scanDomain(ctx, s.db, domainID)
	if err != nil {
		return domain.DomainTLS{}, err
	}
	settings, err := s.GetDomainTLSSettings(ctx, domainID)
	if err != nil {
		return domain.DomainTLS{}, err
	}
	settings.DesiredHostnames = desiredTLSHostnames(mailDomain.Name, settings)
	certificates, err := s.ListTLSCertificates(ctx, domainID)
	if err != nil {
		return domain.DomainTLS{}, err
	}
	if discovered, ok := discoverFilesystemCertificate(domainID, mailDomain.Name, settings); ok && !certificateListHasPath(certificates, discovered.CertPath) {
		certificates = append([]domain.TLSCertificate{discovered}, certificates...)
	}
	jobs, err := s.ListTLSCertificateJobs(ctx, domainID)
	if err != nil {
		return domain.DomainTLS{}, err
	}
	return domain.DomainTLS{
		DomainID:     domainID,
		Domain:       normalizeDNSName(mailDomain.Name),
		Settings:     settings,
		Certificates: certificates,
		Jobs:         jobs,
	}, nil
}

func (s SQLStore) GetDomainTLSSettings(ctx context.Context, domainID uint64) (domain.DomainTLSSettings, error) {
	settings := defaultDomainTLSSettings(domainID)
	err := s.db.QueryRowContext(ctx, `
		SELECT domain_id, dns_webmail_alias_enabled, dns_admin_alias_enabled, tls_mode, challenge_type, use_for_https, use_for_mail_sni,
		       include_mail_hostname, include_webmail_hostname, include_admin_hostname, certificate_name, custom_cert_path, custom_key_path, custom_chain_path, updated_at
		FROM domain_tls_settings
		WHERE domain_id = ?`, domainID).Scan(
		&settings.DomainID,
		&settings.DNSWebmailAliasEnabled,
		&settings.DNSAdminAliasEnabled,
		&settings.TLSMode,
		&settings.ChallengeType,
		&settings.UseForHTTPS,
		&settings.UseForMailSNI,
		&settings.IncludeMailHostname,
		&settings.IncludeWebmailHostname,
		&settings.IncludeAdminHostname,
		&settings.CertificateName,
		&settings.CustomCertPath,
		&settings.CustomKeyPath,
		&settings.CustomChainPath,
		&settings.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		if _, domainErr := scanDomain(ctx, s.db, domainID); domainErr != nil {
			return domain.DomainTLSSettings{}, domainErr
		}
		return settings, nil
	}
	if err != nil {
		return domain.DomainTLSSettings{}, err
	}
	return normalizeDomainTLSSettings(settings), nil
}

func (s SQLStore) UpdateDomainTLSSettings(ctx context.Context, settings domain.DomainTLSSettings) (domain.DomainTLSSettings, error) {
	if _, err := scanDomain(ctx, s.db, settings.DomainID); err != nil {
		return domain.DomainTLSSettings{}, err
	}
	settings = normalizeDomainTLSSettings(settings)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO domain_tls_settings(domain_id, dns_webmail_alias_enabled, dns_admin_alias_enabled, tls_mode, challenge_type, use_for_https, use_for_mail_sni,
		                                include_mail_hostname, include_webmail_hostname, include_admin_hostname, certificate_name, custom_cert_path, custom_key_path, custom_chain_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		  dns_webmail_alias_enabled = VALUES(dns_webmail_alias_enabled),
		  dns_admin_alias_enabled = VALUES(dns_admin_alias_enabled),
		  tls_mode = VALUES(tls_mode),
		  challenge_type = VALUES(challenge_type),
		  use_for_https = VALUES(use_for_https),
		  use_for_mail_sni = VALUES(use_for_mail_sni),
		  include_mail_hostname = VALUES(include_mail_hostname),
		  include_webmail_hostname = VALUES(include_webmail_hostname),
		  include_admin_hostname = VALUES(include_admin_hostname),
		  certificate_name = VALUES(certificate_name),
		  custom_cert_path = VALUES(custom_cert_path),
		  custom_key_path = VALUES(custom_key_path),
		  custom_chain_path = VALUES(custom_chain_path)`,
		settings.DomainID,
		settings.DNSWebmailAliasEnabled,
		settings.DNSAdminAliasEnabled,
		settings.TLSMode,
		settings.ChallengeType,
		settings.UseForHTTPS,
		settings.UseForMailSNI,
		settings.IncludeMailHostname,
		settings.IncludeWebmailHostname,
		settings.IncludeAdminHostname,
		settings.CertificateName,
		settings.CustomCertPath,
		settings.CustomKeyPath,
		settings.CustomChainPath,
	)
	if err != nil {
		return domain.DomainTLSSettings{}, err
	}
	updated, err := s.GetDomainTLSSettings(ctx, settings.DomainID)
	if err != nil {
		return domain.DomainTLSSettings{}, err
	}
	domainName, err := s.domainName(ctx, settings.DomainID)
	if err != nil {
		return domain.DomainTLSSettings{}, err
	}
	updated.DesiredHostnames = desiredTLSHostnames(domainName, updated)
	return updated, nil
}

func (s SQLStore) ListTLSCertificates(ctx context.Context, domainID uint64) ([]domain.TLSCertificate, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, domain_id, source, status, common_name, CAST(sans_json AS CHAR), cert_path, key_path, chain_path, issuer, serial, fingerprint_sha256,
		       not_before, not_after, used_for_https, used_for_mail_sni, last_error, created_at, updated_at
		FROM tls_certificates
		WHERE domain_id = ?
		ORDER BY updated_at DESC, id DESC
		LIMIT 50`, domainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var certificates []domain.TLSCertificate
	for rows.Next() {
		cert, err := scanTLSCertificateRows(rows)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, cert)
	}
	return certificates, rows.Err()
}

func (s SQLStore) ListTLSCertificateJobs(ctx context.Context, domainID uint64) ([]domain.TLSCertificateJob, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, domain_id, certificate_id, job_type, challenge_type, status, step, progress, message, error, CAST(hostnames_json AS CHAR), requested_by, created_at, updated_at, started_at, finished_at
		FROM tls_certificate_jobs
		WHERE domain_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT 50`, domainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []domain.TLSCertificateJob
	for rows.Next() {
		job, err := scanTLSJobRows(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

func (s SQLStore) CreateTLSCertificateJob(ctx context.Context, job domain.TLSCertificateJob) (domain.TLSCertificateJob, error) {
	mailDomain, err := scanDomain(ctx, s.db, job.DomainID)
	if err != nil {
		return domain.TLSCertificateJob{}, err
	}
	settings, err := s.GetDomainTLSSettings(ctx, job.DomainID)
	if err != nil {
		return domain.TLSCertificateJob{}, err
	}
	job.JobType = normalizeTLSJobType(job.JobType)
	job.ChallengeType = normalizeTLSChallengeType(firstNonEmpty(job.ChallengeType, settings.ChallengeType))
	job.Status = "queued"
	job.Step = "queued"
	job.Progress = 0
	job.Message = "Waiting for the TLS worker to request and deploy the certificate."
	job.Hostnames = uniqueDNSNames(job.Hostnames)
	if len(job.Hostnames) == 0 {
		job.Hostnames = desiredTLSHostnames(mailDomain.Name, settings)
	}
	hostnamesJSON, err := json.Marshal(job.Hostnames)
	if err != nil {
		return domain.TLSCertificateJob{}, err
	}
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO tls_certificate_jobs(domain_id, job_type, challenge_type, status, step, progress, message, hostnames_json, log_json, requested_by)
		VALUES (?, ?, ?, 'queued', 'queued', 0, ?, ?, JSON_ARRAY(), ?)`,
		job.DomainID, job.JobType, job.ChallengeType, job.Message, string(hostnamesJSON), strings.TrimSpace(job.RequestedBy))
	if err != nil {
		return domain.TLSCertificateJob{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.TLSCertificateJob{}, err
	}
	return scanTLSJob(ctx, s.db, uint64(id))
}

func defaultDomainTLSSettings(domainID uint64) domain.DomainTLSSettings {
	return domain.DomainTLSSettings{
		DomainID:               domainID,
		DNSWebmailAliasEnabled: true,
		DNSAdminAliasEnabled:   true,
		TLSMode:                "inherit",
		ChallengeType:          "dns-cloudflare",
		UseForHTTPS:            true,
		UseForMailSNI:          true,
		IncludeMailHostname:    true,
		IncludeWebmailHostname: true,
		IncludeAdminHostname:   true,
		UpdatedAt:              time.Now().UTC(),
	}
}

func normalizeDomainTLSSettings(settings domain.DomainTLSSettings) domain.DomainTLSSettings {
	if settings.DomainID == 0 {
		return settings
	}
	settings.TLSMode = normalizeTLSMode(settings.TLSMode)
	settings.ChallengeType = normalizeTLSChallengeType(settings.ChallengeType)
	settings.CertificateName = normalizeDNSName(settings.CertificateName)
	settings.CustomCertPath = strings.TrimSpace(settings.CustomCertPath)
	settings.CustomKeyPath = strings.TrimSpace(settings.CustomKeyPath)
	settings.CustomChainPath = strings.TrimSpace(settings.CustomChainPath)
	return settings
}

func normalizeTLSMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "letsencrypt-dns-cloudflare", "letsencrypt-http", "custom", "disabled":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "inherit"
	}
}

func normalizeTLSChallengeType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "http-01", "manual-dns", "custom-import", "none":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "dns-cloudflare"
	}
}

func normalizeTLSJobType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "renew", "import", "deploy", "check":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "issue"
	}
}

func desiredTLSHostnames(domainName string, settings domain.DomainTLSSettings) []string {
	domainName = normalizeDNSName(domainName)
	candidates := make([]string, 0, 3)
	if settings.IncludeMailHostname {
		candidates = append(candidates, "mail."+domainName)
	}
	if settings.DNSWebmailAliasEnabled && settings.IncludeWebmailHostname {
		candidates = append(candidates, "webmail."+domainName)
	}
	if settings.DNSAdminAliasEnabled && settings.IncludeAdminHostname {
		candidates = append(candidates, "madmin."+domainName)
	}
	return uniqueDNSNames(candidates)
}

func uniqueDNSNames(values []string) []string {
	seen := make(map[string]bool)
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = normalizeDNSName(value)
		if value == "" || strings.HasSuffix(value, ".local") || !strings.Contains(value, ".") || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func discoverFilesystemCertificate(domainID uint64, domainName string, settings domain.DomainTLSSettings) (domain.TLSCertificate, bool) {
	names := certificateCandidateNames(domainName, settings)
	for _, name := range names {
		certPath := strings.TrimSpace(settings.CustomCertPath)
		keyPath := strings.TrimSpace(settings.CustomKeyPath)
		chainPath := strings.TrimSpace(settings.CustomChainPath)
		source := "letsencrypt"
		if certPath == "" {
			certPath = filepath.Join("/etc/letsencrypt/live", name, "fullchain.pem")
			keyPath = filepath.Join("/etc/letsencrypt/live", name, "privkey.pem")
			chainPath = filepath.Join("/etc/letsencrypt/live", name, "chain.pem")
		} else {
			source = "custom"
		}
		cert, ok := parseCertificateFile(domainID, source, certPath, keyPath, chainPath)
		if ok {
			cert.UsedForHTTPS = settings.UseForHTTPS
			cert.UsedForMailSNI = settings.UseForMailSNI
			return cert, true
		}
		if strings.TrimSpace(settings.CustomCertPath) != "" {
			break
		}
	}
	return domain.TLSCertificate{}, false
}

func certificateCandidateNames(domainName string, settings domain.DomainTLSSettings) []string {
	candidates := []string{settings.CertificateName}
	candidates = append(candidates, settings.DesiredHostnames...)
	domainName = normalizeDNSName(domainName)
	candidates = append(candidates, "madmin."+domainName, "webmail."+domainName, "mail."+domainName)
	return uniqueDNSNames(candidates)
}

func parseCertificateFile(domainID uint64, source, certPath, keyPath, chainPath string) (domain.TLSCertificate, bool) {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return domain.TLSCertificate{}, false
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return domain.TLSCertificate{}, false
	}
	parsed, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return domain.TLSCertificate{}, false
	}
	fingerprint := sha256.Sum256(parsed.Raw)
	status := "active"
	now := time.Now()
	if now.After(parsed.NotAfter) {
		status = "expired"
	} else if time.Until(parsed.NotAfter) < 30*24*time.Hour {
		status = "expiring"
	}
	notBefore := parsed.NotBefore
	notAfter := parsed.NotAfter
	return domain.TLSCertificate{
		DomainID:          domainID,
		Source:            source,
		Status:            status,
		CommonName:        parsed.Subject.CommonName,
		SANs:              parsed.DNSNames,
		CertPath:          certPath,
		KeyPath:           keyPath,
		ChainPath:         chainPath,
		Issuer:            parsed.Issuer.String(),
		Serial:            parsed.SerialNumber.String(),
		FingerprintSHA256: strings.ToUpper(hex.EncodeToString(fingerprint[:])),
		NotBefore:         &notBefore,
		NotAfter:          &notAfter,
		DaysRemaining:     int(time.Until(parsed.NotAfter).Hours() / 24),
	}, true
}

func certificateListHasPath(certificates []domain.TLSCertificate, certPath string) bool {
	for _, cert := range certificates {
		if strings.TrimSpace(cert.CertPath) == strings.TrimSpace(certPath) && strings.TrimSpace(certPath) != "" {
			return true
		}
	}
	return false
}

type tlsCertificateScanner interface {
	Scan(dest ...any) error
}

func scanTLSCertificateRows(scanner tlsCertificateScanner) (domain.TLSCertificate, error) {
	var cert domain.TLSCertificate
	var sansJSON string
	var notBefore sql.NullTime
	var notAfter sql.NullTime
	err := scanner.Scan(
		&cert.ID,
		&cert.DomainID,
		&cert.Source,
		&cert.Status,
		&cert.CommonName,
		&sansJSON,
		&cert.CertPath,
		&cert.KeyPath,
		&cert.ChainPath,
		&cert.Issuer,
		&cert.Serial,
		&cert.FingerprintSHA256,
		&notBefore,
		&notAfter,
		&cert.UsedForHTTPS,
		&cert.UsedForMailSNI,
		&cert.LastError,
		&cert.CreatedAt,
		&cert.UpdatedAt,
	)
	if err != nil {
		return domain.TLSCertificate{}, err
	}
	_ = json.Unmarshal([]byte(sansJSON), &cert.SANs)
	if notBefore.Valid {
		cert.NotBefore = &notBefore.Time
	}
	if notAfter.Valid {
		cert.NotAfter = &notAfter.Time
		cert.DaysRemaining = int(time.Until(notAfter.Time).Hours() / 24)
	}
	return cert, nil
}

func scanTLSJob(ctx context.Context, q rowQuerier, jobID uint64) (domain.TLSCertificateJob, error) {
	return scanTLSJobRows(q.QueryRowContext(ctx, `
		SELECT id, domain_id, certificate_id, job_type, challenge_type, status, step, progress, message, error, CAST(hostnames_json AS CHAR), requested_by, created_at, updated_at, started_at, finished_at
		FROM tls_certificate_jobs
		WHERE id = ?`, jobID))
}

func scanTLSJobRows(scanner tlsCertificateScanner) (domain.TLSCertificateJob, error) {
	var job domain.TLSCertificateJob
	var certificateID sql.NullInt64
	var hostnamesJSON string
	var startedAt sql.NullTime
	var finishedAt sql.NullTime
	err := scanner.Scan(
		&job.ID,
		&job.DomainID,
		&certificateID,
		&job.JobType,
		&job.ChallengeType,
		&job.Status,
		&job.Step,
		&job.Progress,
		&job.Message,
		&job.Error,
		&hostnamesJSON,
		&job.RequestedBy,
		&job.CreatedAt,
		&job.UpdatedAt,
		&startedAt,
		&finishedAt,
	)
	if err != nil {
		return domain.TLSCertificateJob{}, err
	}
	if certificateID.Valid {
		id := uint64(certificateID.Int64)
		job.CertificateID = &id
	}
	_ = json.Unmarshal([]byte(hostnamesJSON), &job.Hostnames)
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		job.FinishedAt = &finishedAt.Time
	}
	return job, nil
}

func normalizeHostnameMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "head-domain", "per-domain":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "shared"
	}
}

func normalizeProxyTLSMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "none", "behind-proxy", "letsencrypt-http", "letsencrypt-dns-cloudflare", "custom-cert":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "system"
	}
}

func normalizeLanguageOrDefault(value string) string {
	if normalized := i18n.NormalizeLanguage(value); normalized != "" {
		return normalized
	}
	return "en"
}

func hostBelongsToDomain(host, domainName string) bool {
	host = normalizeDNSName(host)
	domainName = normalizeDNSName(domainName)
	return host == domainName || strings.HasSuffix(host, "."+domainName)
}

func normalizeDNSName(value string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
}

func clientSetup(domainName, mailHost string) []domain.ClientSetup {
	sampleEmail := "user@" + domainName
	return []domain.ClientSetup{
		{
			Client: "Thunderbird",
			Method: "Autoconfig XML",
			Status: "supported",
			URLs: []domain.SetupURL{
				{Label: "Autoconfig URL", Value: "https://" + mailHost + "/.well-known/autoconfig/mail/config-v1.1.xml?emailaddress=" + sampleEmail},
				{Label: "Legacy URL", Value: "https://autoconfig." + domainName + "/mail/config-v1.1.xml"},
			},
			DNS:   []domain.DNSRecord{{Type: "CNAME", Name: "autoconfig." + domainName, Value: mailHost}},
			Notes: []string{"Use full email address as username."},
		},
		{
			Client: "Outlook",
			Method: "Autodiscover XML",
			Status: "supported",
			URLs: []domain.SetupURL{
				{Label: "Autodiscover URL", Value: "https://autodiscover." + domainName + "/autodiscover/autodiscover.xml"},
			},
			DNS:   []domain.DNSRecord{{Type: "CNAME", Name: "autodiscover." + domainName, Value: mailHost}},
			Notes: []string{"Exchange/MAPI is not implemented; Outlook receives IMAP, POP3, SMTP, and DAV settings."},
		},
		{
			Client: "iPhone, iPad, macOS",
			Method: "IMAP/SMTP plus CalDAV/CardDAV",
			Status: "supported",
			URLs: []domain.SetupURL{
				{Label: "CalDAV", Value: "https://" + mailHost + "/dav/calendars/" + sampleEmail + "/default/"},
				{Label: "CardDAV", Value: "https://" + mailHost + "/dav/addressbooks/" + sampleEmail + "/default/"},
			},
			Notes: []string{"Mail uses IMAP 993 and SMTP submission 587; calendar and contacts use DAV."},
		},
		{
			Client: "Gmail app",
			Method: "Manual IMAP/SMTP",
			Status: "manual",
			URLs: []domain.SetupURL{
				{Label: "IMAP", Value: mailHost + ":993 SSL/TLS"},
				{Label: "SMTP", Value: mailHost + ":587 STARTTLS"},
			},
			Notes: []string{"Google's Gmail client does not consume domain autoconfig for arbitrary IMAP providers; use full email address as username."},
		},
	}
}

func (s SQLStore) GetCloudflareConfig(ctx context.Context, domainID uint64) (domain.CloudflareConfig, error) {
	config, err := scanCloudflareConfig(ctx, s.db, domainID)
	if errors.Is(err, sql.ErrNoRows) {
		if _, domainErr := scanDomain(ctx, s.db, domainID); domainErr != nil {
			return domain.CloudflareConfig{}, domainErr
		}
		return domain.CloudflareConfig{DomainID: domainID, Status: "not_configured"}, nil
	}
	return config, err
}

func (s SQLStore) SaveCloudflareConfig(ctx context.Context, domainID uint64, zoneID, apiToken string) (domain.CloudflareConfig, error) {
	if _, err := scanDomain(ctx, s.db, domainID); err != nil {
		return domain.CloudflareConfig{}, err
	}
	existing, err := scanCloudflareConfig(ctx, s.db, domainID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return domain.CloudflareConfig{}, err
	}
	if strings.TrimSpace(apiToken) == "" && existing.TokenConfigured {
		apiToken = existingTokenPlaceholder
	}
	if strings.TrimSpace(apiToken) == "" && !existing.TokenConfigured {
		return domain.CloudflareConfig{}, errCloudflareTokenRequired
	}
	if apiToken == existingTokenPlaceholder {
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO cloudflare_domain_configs(domain_id, zone_id, api_token, status)
			VALUES (?, ?, '', 'configured')
			ON DUPLICATE KEY UPDATE zone_id = VALUES(zone_id), status = 'configured', last_error = NULL`, domainID, zoneID)
	} else {
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO cloudflare_domain_configs(domain_id, zone_id, api_token, status)
			VALUES (?, ?, ?, 'configured')
			ON DUPLICATE KEY UPDATE zone_id = VALUES(zone_id), api_token = VALUES(api_token), status = 'configured', last_error = NULL`, domainID, zoneID, apiToken)
	}
	if err != nil {
		return domain.CloudflareConfig{}, err
	}
	return scanCloudflareConfig(ctx, s.db, domainID)
}

const existingTokenPlaceholder = "__keep_existing_token__"

func scanCloudflareConfig(ctx context.Context, q rowQuerier, domainID uint64) (domain.CloudflareConfig, error) {
	var config domain.CloudflareConfig
	var token string
	var lastChecked sql.NullTime
	var lastError sql.NullString
	err := q.QueryRowContext(ctx, `
		SELECT domain_id, zone_id, zone_name, api_token, status, last_checked_at, last_error, updated_at
		FROM cloudflare_domain_configs
		WHERE domain_id = ?`, domainID).Scan(&config.DomainID, &config.ZoneID, &config.ZoneName, &token, &config.Status, &lastChecked, &lastError, &config.UpdatedAt)
	if err != nil {
		return domain.CloudflareConfig{}, err
	}
	config.TokenConfigured = strings.TrimSpace(token) != ""
	if lastError.Valid {
		config.LastError = lastError.String
	}
	if lastChecked.Valid {
		config.LastCheckedAt = &lastChecked.Time
	}
	return config, nil
}

func scanCloudflareToken(ctx context.Context, q rowQuerier, domainID uint64) (string, domain.CloudflareConfig, error) {
	var config domain.CloudflareConfig
	var token string
	var lastChecked sql.NullTime
	var lastError sql.NullString
	err := q.QueryRowContext(ctx, `
		SELECT domain_id, zone_id, zone_name, api_token, status, last_checked_at, last_error, updated_at
		FROM cloudflare_domain_configs
		WHERE domain_id = ?`, domainID).Scan(&config.DomainID, &config.ZoneID, &config.ZoneName, &token, &config.Status, &lastChecked, &lastError, &config.UpdatedAt)
	if err != nil {
		return "", domain.CloudflareConfig{}, err
	}
	config.TokenConfigured = strings.TrimSpace(token) != ""
	if lastError.Valid {
		config.LastError = lastError.String
	}
	if lastChecked.Valid {
		config.LastCheckedAt = &lastChecked.Time
	}
	return token, config, nil
}

func (s SQLStore) CheckCloudflareDNS(ctx context.Context, domainID uint64) (domain.DNSProvisionPlan, error) {
	plan, _, err := s.cloudflarePlan(ctx, domainID)
	return plan, err
}

func (s SQLStore) ApplyCloudflareDNS(ctx context.Context, domainID uint64, replace bool) (domain.DNSProvisionResult, error) {
	plan, ctxData, err := s.cloudflarePlan(ctx, domainID)
	if err != nil {
		return domain.DNSProvisionResult{}, err
	}
	if plan.Status == "blocked" {
		return domain.DNSProvisionResult{}, errDomainDNSNotReady
	}
	hasBlocking := false
	for _, action := range plan.Actions {
		if action.Action == "conflict" || action.Action == "blocked" {
			hasBlocking = true
		}
	}
	if hasBlocking && !replace {
		return domain.DNSProvisionResult{}, fmt.Errorf("%w; run with replace enabled after review", errCloudflareDNSConflicts)
	}
	backupRecords := make([]cloudflareDNSRecord, 0)
	changed := 0
	for _, planned := range ctxData.planned {
		switch planned.action.Action {
		case "ok":
			continue
		case "create":
			if err := ctxData.client.createRecord(ctx, ctxData.zoneID, planned.desired); err != nil {
				return domain.DNSProvisionResult{}, err
			}
			changed++
		case "conflict", "blocked":
			if !replace {
				continue
			}
			for _, existing := range planned.touch {
				backupRecords = append(backupRecords, existing)
			}
			if len(planned.touch) > 0 && planned.touch[0].Type == planned.desired.Type {
				if err := ctxData.client.updateRecord(ctx, ctxData.zoneID, planned.touch[0].ID, planned.desired); err != nil {
					return domain.DNSProvisionResult{}, err
				}
				changed++
				for _, extra := range planned.touch[1:] {
					if err := ctxData.client.deleteRecord(ctx, ctxData.zoneID, extra.ID); err != nil {
						return domain.DNSProvisionResult{}, err
					}
					changed++
				}
				continue
			}
			for _, existing := range planned.touch {
				if err := ctxData.client.deleteRecord(ctx, ctxData.zoneID, existing.ID); err != nil {
					return domain.DNSProvisionResult{}, err
				}
				changed++
			}
			if err := ctxData.client.createRecord(ctx, ctxData.zoneID, planned.desired); err != nil {
				return domain.DNSProvisionResult{}, err
			}
			changed++
		}
	}
	backupJSON, _ := json.Marshal(backupRecords)
	planJSON, _ := json.Marshal(plan)
	resultJSON, _ := json.Marshal(map[string]any{"changed": changed, "replace": replace})
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO dns_provision_backups(domain_id, provider, zone_id, mode, replace_existing, plan_json, backup_json, result_json)
		VALUES (?, 'cloudflare', ?, 'apply', ?, ?, ?, ?)`, domainID, ctxData.zoneID, replace, string(planJSON), string(backupJSON), string(resultJSON))
	if err != nil {
		return domain.DNSProvisionResult{}, err
	}
	backupID, _ := result.LastInsertId()
	_ = s.markCloudflareChecked(ctx, domainID, ctxData.zoneID, plan.ZoneName, "checked", "")
	return domain.DNSProvisionResult{Plan: plan, BackupID: uint64(backupID), Applied: true, Changed: changed, BackupJSON: string(backupJSON)}, nil
}

type cloudflarePlanContext struct {
	client  cloudflareClient
	zoneID  string
	planned []plannedCloudflareAction
}

type plannedCloudflareAction struct {
	action  domain.DNSProvisionAction
	desired cloudflareDNSRecord
	touch   []cloudflareDNSRecord
}

func (s SQLStore) cloudflarePlan(ctx context.Context, domainID uint64) (domain.DNSProvisionPlan, cloudflarePlanContext, error) {
	token, config, err := scanCloudflareToken(ctx, s.db, domainID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.DNSProvisionPlan{}, cloudflarePlanContext{}, errCloudflareTokenRequired
		}
		return domain.DNSProvisionPlan{}, cloudflarePlanContext{}, err
	}
	if strings.TrimSpace(token) == "" {
		return domain.DNSProvisionPlan{}, cloudflarePlanContext{}, errCloudflareTokenRequired
	}
	dns, err := s.GetDomainDNS(ctx, domainID)
	if err != nil {
		return domain.DNSProvisionPlan{}, cloudflarePlanContext{}, err
	}
	if !dns.Provisionable {
		return domain.DNSProvisionPlan{
			DomainID:       domainID,
			Domain:         dns.Domain,
			Provider:       "cloudflare",
			Status:         "blocked",
			ReplaceAllowed: false,
			Actions: []domain.DNSProvisionAction{{
				Action: "blocked",
				Type:   "DNS",
				Name:   dns.Domain,
				Reason: strings.Join(dns.Warnings, " "),
			}},
			Summary: strings.Join(dns.Warnings, " "),
		}, cloudflarePlanContext{}, nil
	}
	client := cloudflareClient{token: token, httpClient: http.DefaultClient, baseURL: cloudflareAPIBaseURL}
	zoneID := strings.TrimSpace(config.ZoneID)
	zoneName := strings.TrimSpace(config.ZoneName)
	if zoneID == "" {
		zone, err := client.findZone(ctx, dns.Domain)
		if err != nil {
			_ = s.markCloudflareChecked(ctx, domainID, "", "", "error", err.Error())
			return domain.DNSProvisionPlan{}, cloudflarePlanContext{}, err
		}
		zoneID = zone.ID
		zoneName = zone.Name
	} else if zoneName == "" {
		zoneName = dns.Domain
	}
	existing, err := client.listRecords(ctx, zoneID)
	if err != nil {
		_ = s.markCloudflareChecked(ctx, domainID, zoneID, zoneName, "error", err.Error())
		return domain.DNSProvisionPlan{}, cloudflarePlanContext{}, err
	}
	desired := desiredCloudflareRecords(dns.Records)
	planned := planCloudflareActions(desired, existing)
	actions := make([]domain.DNSProvisionAction, 0, len(planned))
	conflicts := 0
	creates := 0
	for _, item := range planned {
		actions = append(actions, item.action)
		if item.action.Action == "conflict" || item.action.Action == "blocked" {
			conflicts++
		}
		if item.action.Action == "create" {
			creates++
		}
	}
	status := "ok"
	summary := "Cloudflare DNS already matches desired mail records."
	if conflicts > 0 {
		status = "conflicts"
		summary = fmt.Sprintf("%d conflicting records need review before replacement.", conflicts)
	} else if creates > 0 {
		status = "changes"
		summary = fmt.Sprintf("%d records can be created safely.", creates)
	}
	_ = s.markCloudflareChecked(ctx, domainID, zoneID, zoneName, "checked", "")
	return domain.DNSProvisionPlan{
		DomainID:       domainID,
		Domain:         dns.Domain,
		Provider:       "cloudflare",
		ZoneID:         zoneID,
		ZoneName:       zoneName,
		Status:         status,
		ReplaceAllowed: conflicts > 0,
		Actions:        actions,
		Summary:        summary,
	}, cloudflarePlanContext{client: client, zoneID: zoneID, planned: planned}, nil
}

func (s SQLStore) markCloudflareChecked(ctx context.Context, domainID uint64, zoneID, zoneName, status, message string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE cloudflare_domain_configs
		SET zone_id = COALESCE(NULLIF(?, ''), zone_id), zone_name = COALESCE(NULLIF(?, ''), zone_name), status = ?, last_checked_at = CURRENT_TIMESTAMP, last_error = NULLIF(?, '')
		WHERE domain_id = ?`, zoneID, zoneName, status, message, domainID)
	return err
}

var cloudflareAPIBaseURL = "https://api.cloudflare.com/client/v4"

type cloudflareClient struct {
	token      string
	baseURL    string
	httpClient *http.Client
}

type cloudflareZone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type cloudflareDNSRecord struct {
	ID       string         `json:"id,omitempty"`
	Type     string         `json:"type"`
	Name     string         `json:"name"`
	Content  string         `json:"content,omitempty"`
	Priority *int           `json:"priority,omitempty"`
	TTL      int            `json:"ttl,omitempty"`
	Proxied  *bool          `json:"proxied,omitempty"`
	Comment  string         `json:"comment,omitempty"`
	Data     map[string]any `json:"data,omitempty"`
}

type cloudflareEnvelope[T any] struct {
	Success    bool              `json:"success"`
	Result     T                 `json:"result"`
	Errors     []cloudflareError `json:"errors"`
	ResultInfo struct {
		Page       int `json:"page"`
		PerPage    int `json:"per_page"`
		TotalPages int `json:"total_pages"`
		Count      int `json:"count"`
		TotalCount int `json:"total_count"`
	} `json:"result_info"`
}

type cloudflareError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type cloudflareEnvelopeStatus interface {
	cloudflareOK() bool
	cloudflareErrors() []cloudflareError
}

func (e *cloudflareEnvelope[T]) cloudflareOK() bool {
	return e.Success
}

func (e *cloudflareEnvelope[T]) cloudflareErrors() []cloudflareError {
	return e.Errors
}

func (c cloudflareClient) findZone(ctx context.Context, zoneName string) (cloudflareZone, error) {
	var envelope cloudflareEnvelope[[]cloudflareZone]
	if err := c.request(ctx, http.MethodGet, "/zones?name="+url.QueryEscape(zoneName), nil, &envelope); err != nil {
		return cloudflareZone{}, err
	}
	if len(envelope.Result) == 0 {
		return cloudflareZone{}, fmt.Errorf("cloudflare zone %s not found or token has no access", zoneName)
	}
	return envelope.Result[0], nil
}

func (c cloudflareClient) listRecords(ctx context.Context, zoneID string) ([]cloudflareDNSRecord, error) {
	all := make([]cloudflareDNSRecord, 0)
	for page := 1; ; page++ {
		var envelope cloudflareEnvelope[[]cloudflareDNSRecord]
		path := fmt.Sprintf("/zones/%s/dns_records?per_page=500&page=%d", url.PathEscape(zoneID), page)
		if err := c.request(ctx, http.MethodGet, path, nil, &envelope); err != nil {
			return nil, err
		}
		all = append(all, envelope.Result...)
		if envelope.ResultInfo.TotalPages == 0 || page >= envelope.ResultInfo.TotalPages {
			break
		}
	}
	return all, nil
}

func (c cloudflareClient) createRecord(ctx context.Context, zoneID string, record cloudflareDNSRecord) error {
	var envelope cloudflareEnvelope[cloudflareDNSRecord]
	return c.request(ctx, http.MethodPost, "/zones/"+url.PathEscape(zoneID)+"/dns_records", record, &envelope)
}

func (c cloudflareClient) updateRecord(ctx context.Context, zoneID, recordID string, record cloudflareDNSRecord) error {
	var envelope cloudflareEnvelope[cloudflareDNSRecord]
	return c.request(ctx, http.MethodPut, "/zones/"+url.PathEscape(zoneID)+"/dns_records/"+url.PathEscape(recordID), record, &envelope)
}

func (c cloudflareClient) deleteRecord(ctx context.Context, zoneID, recordID string) error {
	var envelope cloudflareEnvelope[map[string]any]
	return c.request(ctx, http.MethodDelete, "/zones/"+url.PathEscape(zoneID)+"/dns_records/"+url.PathEscape(recordID), nil, &envelope)
}

func (c cloudflareClient) request(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.baseURL, "/")+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode cloudflare response: %w", err)
	}
	if envelope, ok := out.(cloudflareEnvelopeStatus); ok && !envelope.cloudflareOK() {
		return cloudflareAPIError(envelope.cloudflareErrors())
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("cloudflare api returned http %d", resp.StatusCode)
	}
	return nil
}

func cloudflareAPIError(errors []cloudflareError) error {
	if len(errors) == 0 {
		return fmt.Errorf("cloudflare api request failed")
	}
	parts := make([]string, 0, len(errors))
	for _, item := range errors {
		parts = append(parts, fmt.Sprintf("%d %s", item.Code, item.Message))
	}
	return fmt.Errorf("cloudflare api request failed: %s", strings.Join(parts, "; "))
}

func desiredCloudflareRecords(records []domain.DNSRecord) []cloudflareDNSRecord {
	out := make([]cloudflareDNSRecord, 0, len(records))
	for _, record := range records {
		item := cloudflareDNSRecord{Type: record.Type, Name: record.Name, Content: record.Value, Priority: record.Priority, TTL: 1, Comment: "Managed by ProIdentity Mail"}
		if record.Type == "A" || record.Type == "AAAA" || record.Type == "CNAME" {
			if record.Proxied != nil {
				item.Proxied = boolPtr(*record.Proxied)
			} else {
				item.Proxied = boolPtr(false)
			}
		}
		if record.Type == "SRV" {
			parts := strings.Fields(record.Value)
			if len(parts) == 4 {
				priority, _ := strconv.Atoi(parts[0])
				weight, _ := strconv.Atoi(parts[1])
				port, _ := strconv.Atoi(parts[2])
				item.Content = ""
				item.Data = map[string]any{"priority": priority, "weight": weight, "port": port, "target": normalizeDNSName(parts[3])}
			}
		}
		out = append(out, item)
	}
	return out
}

func planCloudflareActions(desired, existing []cloudflareDNSRecord) []plannedCloudflareAction {
	planned := make([]plannedCloudflareAction, 0, len(desired))
	for _, want := range desired {
		touch := relevantExisting(want, existing)
		action := domain.DNSProvisionAction{Action: "create", Type: want.Type, Name: want.Name, Value: recordComparable(want), Priority: want.Priority}
		if len(touch) == 0 {
			action.Reason = "record is missing"
			planned = append(planned, plannedCloudflareAction{action: action, desired: want})
			continue
		}
		if len(touch) == 1 && strings.EqualFold(recordComparable(touch[0]), recordComparable(want)) && sameRecordPriority(touch[0], want) && strings.EqualFold(touch[0].Type, want.Type) && sameProxied(touch[0], want) {
			action.Action = "ok"
			action.Reason = "record already matches"
			planned = append(planned, plannedCloudflareAction{action: action, desired: want, touch: touch})
			continue
		}
		action.Action = "conflict"
		action.Reason = "existing record differs"
		if len(touch) == 1 && strings.EqualFold(recordComparable(touch[0]), recordComparable(want)) && sameRecordPriority(touch[0], want) && strings.EqualFold(touch[0].Type, want.Type) && !sameProxied(touch[0], want) {
			if proxiedValue(want.Proxied) {
				action.Reason = "existing record is DNS-only; browser aliases should be proxied"
			} else {
				action.Reason = "existing record is proxied; mail service records must be DNS-only"
			}
		}
		if want.Type == "CNAME" && len(touch) > 0 && !allType(touch, "CNAME") {
			action.Action = "blocked"
			action.Reason = "CNAME cannot coexist with other record types at the same name"
		}
		action.Existing = make([]domain.DNSRecord, 0, len(touch))
		for _, current := range touch {
			action.Existing = append(action.Existing, domain.DNSRecord{Type: current.Type, Name: current.Name, Value: recordComparable(current), Priority: current.Priority})
		}
		planned = append(planned, plannedCloudflareAction{action: action, desired: want, touch: touch})
	}
	return planned
}

func relevantExisting(want cloudflareDNSRecord, existing []cloudflareDNSRecord) []cloudflareDNSRecord {
	out := make([]cloudflareDNSRecord, 0)
	for _, current := range existing {
		if !strings.EqualFold(current.Name, want.Name) {
			continue
		}
		if want.Type == "CNAME" {
			out = append(out, current)
			continue
		}
		if !strings.EqualFold(current.Type, want.Type) {
			continue
		}
		if want.Type == "TXT" && strings.EqualFold(want.Name, strings.TrimSuffix(want.Name, ".")) {
			if strings.HasPrefix(strings.ToLower(want.Content), "v=spf1") && !strings.HasPrefix(strings.ToLower(current.Content), "v=spf1") {
				continue
			}
		}
		out = append(out, current)
	}
	return out
}

func recordComparable(record cloudflareDNSRecord) string {
	if strings.EqualFold(record.Type, "SRV") {
		if srv, ok := srvComparable(record); ok {
			return srv
		}
	}
	return strings.TrimSpace(record.Content)
}

func srvComparable(record cloudflareDNSRecord) (string, bool) {
	if len(record.Data) > 0 {
		priority, okPriority := numberFromCloudflareValue(record.Data["priority"])
		weight, okWeight := numberFromCloudflareValue(record.Data["weight"])
		port, okPort := numberFromCloudflareValue(record.Data["port"])
		target, okTarget := record.Data["target"].(string)
		if okPriority && okWeight && okPort && okTarget && strings.TrimSpace(target) != "" {
			return fmt.Sprintf("%d %d %d %s", priority, weight, port, normalizeDNSName(target)), true
		}
	}
	parts := strings.Fields(record.Content)
	if len(parts) == 4 {
		priority, errPriority := strconv.Atoi(parts[0])
		weight, errWeight := strconv.Atoi(parts[1])
		port, errPort := strconv.Atoi(parts[2])
		if errPriority == nil && errWeight == nil && errPort == nil {
			return fmt.Sprintf("%d %d %d %s", priority, weight, port, normalizeDNSName(parts[3])), true
		}
	}
	if len(parts) == 3 && record.Priority != nil {
		weight, errWeight := strconv.Atoi(parts[0])
		port, errPort := strconv.Atoi(parts[1])
		if errWeight == nil && errPort == nil {
			return fmt.Sprintf("%d %d %d %s", *record.Priority, weight, port, normalizeDNSName(parts[2])), true
		}
	}
	return strings.TrimSpace(record.Content), false
}

func numberFromCloudflareValue(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	case json.Number:
		n, err := typed.Int64()
		return int(n), err == nil
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(typed))
		return n, err == nil
	default:
		return 0, false
	}
}

func sameRecordPriority(a, b cloudflareDNSRecord) bool {
	if strings.EqualFold(a.Type, "SRV") && strings.EqualFold(b.Type, "SRV") {
		return true
	}
	return samePriority(a.Priority, b.Priority)
}

func samePriority(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func sameProxied(a, b cloudflareDNSRecord) bool {
	return proxiedValue(a.Proxied) == proxiedValue(b.Proxied)
}

func proxiedValue(value *bool) bool {
	return value != nil && *value
}

func boolPtr(value bool) *bool {
	return &value
}

func allType(records []cloudflareDNSRecord, recordType string) bool {
	for _, record := range records {
		if !strings.EqualFold(record.Type, recordType) {
			return false
		}
	}
	return true
}

func normalizeDKIMTXT(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", " ")
	if start := strings.Index(value, "("); start >= 0 {
		if end := strings.LastIndex(value, ")"); end > start {
			value = value[start+1 : end]
		}
	}
	value = strings.ReplaceAll(value, `"`, "")
	value = strings.Join(strings.Fields(value), " ")
	return value
}
