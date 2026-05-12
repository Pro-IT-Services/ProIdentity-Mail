package domain

import "time"

type Tenant struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Domain struct {
	ID           uint64    `json:"id"`
	TenantID     uint64    `json:"tenant_id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	DKIMSelector string    `json:"dkim_selector"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type User struct {
	ID              uint64    `json:"id"`
	TenantID        uint64    `json:"tenant_id"`
	PrimaryDomainID uint64    `json:"primary_domain_id"`
	LocalPart       string    `json:"local_part"`
	DisplayName     string    `json:"display_name"`
	MailboxType     string    `json:"mailbox_type"`
	PasswordHash    string    `json:"password_hash,omitempty"`
	Status          string    `json:"status"`
	QuotaBytes      uint64    `json:"quota_bytes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type TenantAdmin struct {
	ID        uint64    `json:"id"`
	TenantID  uint64    `json:"tenant_id"`
	UserID    uint64    `json:"user_id"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginRateLimit struct {
	ID            uint64     `json:"id"`
	Service       string     `json:"service"`
	LimiterKey    string     `json:"limiter_key"`
	Scope         string     `json:"scope"`
	Subject       string     `json:"subject"`
	FailureCount  uint       `json:"failure_count"`
	FirstFailedAt *time.Time `json:"first_failed_at,omitempty"`
	LastFailedAt  *time.Time `json:"last_failed_at,omitempty"`
	LockedUntil   *time.Time `json:"locked_until,omitempty"`
	Locked        bool       `json:"locked"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type AdminMFASettings struct {
	ID                            uint64    `json:"id"`
	LocalTOTPEnabled              bool      `json:"local_totp_enabled"`
	LocalTOTPSecret               string    `json:"-"`
	LocalTOTPPendingSecret        string    `json:"-"`
	ProIdentityEnabled            bool      `json:"proidentity_enabled"`
	ProIdentityBaseURL            string    `json:"proidentity_base_url"`
	ProIdentityAPIKey             string    `json:"-"`
	ProIdentityAPIKeyConfigured   bool      `json:"proidentity_api_key_configured"`
	ProIdentityUserEmail          string    `json:"proidentity_user_email"`
	ProIdentityTimeoutSeconds     int       `json:"proidentity_timeout_seconds"`
	ProIdentityTOTPEnabled        bool      `json:"proidentity_totp_enabled"`
	EffectiveProvider             string    `json:"effective_provider"`
	NativeWebAuthnEnabled         bool      `json:"native_webauthn_enabled"`
	NativeWebAuthnCredentialCount int       `json:"native_webauthn_credential_count"`
	UpdatedAt                     time.Time `json:"updated_at"`
}

type AdminMFAChallenge struct {
	Token     string    `json:"-"`
	Username  string    `json:"username"`
	Provider  string    `json:"provider"`
	RequestID string    `json:"request_id,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type AdminWebAuthnCredential struct {
	ID             uint64     `json:"id"`
	CredentialID   []byte     `json:"-"`
	Name           string     `json:"name"`
	CredentialJSON []byte     `json:"-"`
	CreatedAt      time.Time  `json:"created_at"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
}

type AdminWebAuthnSession struct {
	Token       string    `json:"-"`
	Ceremony    string    `json:"ceremony"`
	SessionJSON []byte    `json:"-"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type Alias struct {
	ID              uint64    `json:"id"`
	TenantID        uint64    `json:"tenant_id"`
	DomainID        uint64    `json:"domain_id"`
	SourceLocalPart string    `json:"source_local_part"`
	Destination     string    `json:"destination"`
	CreatedAt       time.Time `json:"created_at"`
}

type CatchAllRoute struct {
	ID          uint64    `json:"id"`
	TenantID    uint64    `json:"tenant_id"`
	DomainID    uint64    `json:"domain_id"`
	Destination string    `json:"destination"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SharedMailboxPermission struct {
	ID              uint64    `json:"id"`
	TenantID        uint64    `json:"tenant_id"`
	SharedMailboxID uint64    `json:"shared_mailbox_id"`
	UserID          uint64    `json:"user_id"`
	CanRead         bool      `json:"can_read"`
	CanSendAs       bool      `json:"can_send_as"`
	CanSendOnBehalf bool      `json:"can_send_on_behalf"`
	CanManage       bool      `json:"can_manage"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type DNSRecord struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Priority *int   `json:"priority,omitempty"`
	Purpose  string `json:"purpose,omitempty"`
	Required bool   `json:"required,omitempty"`
	Proxied  *bool  `json:"proxied,omitempty"`
}

type DomainDNS struct {
	DomainID      uint64        `json:"domain_id"`
	Domain        string        `json:"domain"`
	MailHost      string        `json:"mail_host"`
	Records       []DNSRecord   `json:"records"`
	ClientSetup   []ClientSetup `json:"client_setup,omitempty"`
	Warnings      []string      `json:"warnings,omitempty"`
	Provisionable bool          `json:"provisionable"`
}

type DomainTLS struct {
	DomainID     uint64              `json:"domain_id"`
	Domain       string              `json:"domain"`
	Settings     DomainTLSSettings   `json:"settings"`
	Certificates []TLSCertificate    `json:"certificates"`
	Jobs         []TLSCertificateJob `json:"jobs"`
}

type DomainTLSSettings struct {
	DomainID               uint64    `json:"domain_id"`
	DNSWebmailAliasEnabled bool      `json:"dns_webmail_alias_enabled"`
	DNSAdminAliasEnabled   bool      `json:"dns_admin_alias_enabled"`
	TLSMode                string    `json:"tls_mode"`
	ChallengeType          string    `json:"challenge_type"`
	UseForHTTPS            bool      `json:"use_for_https"`
	UseForMailSNI          bool      `json:"use_for_mail_sni"`
	IncludeMailHostname    bool      `json:"include_mail_hostname"`
	IncludeWebmailHostname bool      `json:"include_webmail_hostname"`
	IncludeAdminHostname   bool      `json:"include_admin_hostname"`
	CertificateName        string    `json:"certificate_name,omitempty"`
	CustomCertPath         string    `json:"custom_cert_path,omitempty"`
	CustomKeyPath          string    `json:"custom_key_path,omitempty"`
	CustomChainPath        string    `json:"custom_chain_path,omitempty"`
	DesiredHostnames       []string  `json:"desired_hostnames,omitempty"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type TLSCertificate struct {
	ID                uint64     `json:"id"`
	DomainID          uint64     `json:"domain_id"`
	DomainName        string     `json:"domain_name,omitempty"`
	Source            string     `json:"source"`
	Status            string     `json:"status"`
	CommonName        string     `json:"common_name,omitempty"`
	SANs              []string   `json:"sans,omitempty"`
	CertPath          string     `json:"cert_path,omitempty"`
	KeyPath           string     `json:"key_path,omitempty"`
	ChainPath         string     `json:"chain_path,omitempty"`
	Issuer            string     `json:"issuer,omitempty"`
	Serial            string     `json:"serial,omitempty"`
	FingerprintSHA256 string     `json:"fingerprint_sha256,omitempty"`
	NotBefore         *time.Time `json:"not_before,omitempty"`
	NotAfter          *time.Time `json:"not_after,omitempty"`
	DaysRemaining     int        `json:"days_remaining"`
	UsedForHTTPS      bool       `json:"used_for_https"`
	UsedForMailSNI    bool       `json:"used_for_mail_sni"`
	LastError         string     `json:"last_error,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type TLSCertificateJob struct {
	ID            uint64     `json:"id"`
	DomainID      uint64     `json:"domain_id"`
	CertificateID *uint64    `json:"certificate_id,omitempty"`
	JobType       string     `json:"job_type"`
	ChallengeType string     `json:"challenge_type"`
	Status        string     `json:"status"`
	Step          string     `json:"step"`
	Progress      int        `json:"progress"`
	Message       string     `json:"message,omitempty"`
	Error         string     `json:"error,omitempty"`
	Hostnames     []string   `json:"hostnames,omitempty"`
	RequestedBy   string     `json:"requested_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
}

type MailServerSettings struct {
	HostnameMode            string  `json:"hostname_mode"`
	MailHostname            string  `json:"mail_hostname"`
	HeadTenantID            *uint64 `json:"head_tenant_id,omitempty"`
	HeadDomainID            *uint64 `json:"head_domain_id,omitempty"`
	PublicIPv4              string  `json:"public_ipv4,omitempty"`
	PublicIPv6              string  `json:"public_ipv6,omitempty"`
	SNIEnabled              bool    `json:"sni_enabled"`
	TLSMode                 string  `json:"tls_mode"`
	ForceHTTPS              bool    `json:"force_https"`
	HTTPSCertificateID      *uint64 `json:"https_certificate_id,omitempty"`
	HTTPSCertificateName    string  `json:"https_certificate_name,omitempty"`
	HTTPSCertPath           string  `json:"https_cert_path,omitempty"`
	HTTPSKeyPath            string  `json:"https_key_path,omitempty"`
	HTTPSChainPath          string  `json:"https_chain_path,omitempty"`
	DefaultLanguage         string  `json:"default_language"`
	MailboxMFAEnabled       bool    `json:"mailbox_mfa_enabled"`
	ForceMailboxMFA         bool    `json:"force_mailbox_mfa"`
	CloudflareRealIPEnabled bool    `json:"cloudflare_real_ip_enabled"`
	EffectiveHostname       string  `json:"effective_hostname,omitempty"`
	ConfigApplyQueued       bool    `json:"config_apply_queued,omitempty"`
	ConfigApplyError        string  `json:"config_apply_error,omitempty"`
}

type SetupURL struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type ClientSetup struct {
	Client string      `json:"client"`
	Method string      `json:"method"`
	Status string      `json:"status"`
	URLs   []SetupURL  `json:"urls,omitempty"`
	DNS    []DNSRecord `json:"dns,omitempty"`
	Notes  []string    `json:"notes,omitempty"`
}

type CloudflareConfig struct {
	DomainID        uint64     `json:"domain_id"`
	ZoneID          string     `json:"zone_id,omitempty"`
	ZoneName        string     `json:"zone_name,omitempty"`
	Status          string     `json:"status"`
	TokenConfigured bool       `json:"token_configured"`
	LastError       string     `json:"last_error,omitempty"`
	LastCheckedAt   *time.Time `json:"last_checked_at,omitempty"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type DNSProvisionAction struct {
	Action   string      `json:"action"`
	Type     string      `json:"type"`
	Name     string      `json:"name"`
	Value    string      `json:"value"`
	Priority *int        `json:"priority,omitempty"`
	Reason   string      `json:"reason,omitempty"`
	Existing []DNSRecord `json:"existing,omitempty"`
}

type DNSProvisionPlan struct {
	DomainID       uint64               `json:"domain_id"`
	Domain         string               `json:"domain"`
	Provider       string               `json:"provider"`
	ZoneID         string               `json:"zone_id,omitempty"`
	ZoneName       string               `json:"zone_name,omitempty"`
	Status         string               `json:"status"`
	ReplaceAllowed bool                 `json:"replace_allowed"`
	Actions        []DNSProvisionAction `json:"actions"`
	Summary        string               `json:"summary"`
}

type DNSProvisionResult struct {
	Plan       DNSProvisionPlan `json:"plan"`
	BackupID   uint64           `json:"backup_id,omitempty"`
	Applied    bool             `json:"applied"`
	Changed    int              `json:"changed"`
	BackupJSON string           `json:"backup_json,omitempty"`
}

type QuarantineEvent struct {
	ID             uint64     `json:"id"`
	TenantID       uint64     `json:"tenant_id"`
	UserID         *uint64    `json:"user_id,omitempty"`
	DomainID       *uint64    `json:"domain_id,omitempty"`
	MessageID      string     `json:"message_id,omitempty"`
	Sender         string     `json:"sender,omitempty"`
	Recipient      string     `json:"recipient"`
	Verdict        string     `json:"verdict"`
	Action         string     `json:"action"`
	Scanner        string     `json:"scanner"`
	SymbolsJSON    string     `json:"symbols_json"`
	StoragePath    string     `json:"storage_path,omitempty"`
	SizeBytes      int64      `json:"size_bytes,omitempty"`
	SHA256         string     `json:"sha256,omitempty"`
	Status         string     `json:"status"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	ResolutionNote string     `json:"resolution_note,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type AuditEvent struct {
	ID           uint64    `json:"id"`
	TenantID     *uint64   `json:"tenant_id,omitempty"`
	ActorType    string    `json:"actor_type"`
	ActorID      *uint64   `json:"actor_id,omitempty"`
	Action       string    `json:"action"`
	TargetType   string    `json:"target_type"`
	TargetID     string    `json:"target_id"`
	MetadataJSON string    `json:"metadata_json"`
	Category     string    `json:"category,omitempty"`
	Severity     string    `json:"severity,omitempty"`
	Title        string    `json:"title,omitempty"`
	Summary      string    `json:"summary,omitempty"`
	ActorLabel   string    `json:"actor_label,omitempty"`
	TargetLabel  string    `json:"target_label,omitempty"`
	Details      []Detail  `json:"details,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type Detail struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type TenantPolicy struct {
	TenantID          uint64    `json:"tenant_id"`
	SpamAction        string    `json:"spam_action"`
	MalwareAction     string    `json:"malware_action"`
	RequireTLSForAuth bool      `json:"require_tls_for_auth"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
