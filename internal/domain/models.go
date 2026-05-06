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
	PasswordHash    string    `json:"password_hash,omitempty"`
	Status          string    `json:"status"`
	QuotaBytes      uint64    `json:"quota_bytes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Alias struct {
	ID              uint64    `json:"id"`
	TenantID        uint64    `json:"tenant_id"`
	DomainID        uint64    `json:"domain_id"`
	SourceLocalPart string    `json:"source_local_part"`
	Destination     string    `json:"destination"`
	CreatedAt       time.Time `json:"created_at"`
}

type DNSRecord struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Priority *int   `json:"priority,omitempty"`
}

type DomainDNS struct {
	DomainID uint64      `json:"domain_id"`
	Domain   string      `json:"domain"`
	Records  []DNSRecord `json:"records"`
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
	CreatedAt    time.Time `json:"created_at"`
}

type TenantPolicy struct {
	TenantID          uint64    `json:"tenant_id"`
	SpamAction        string    `json:"spam_action"`
	MalwareAction     string    `json:"malware_action"`
	RequireTLSForAuth bool      `json:"require_tls_for_auth"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
