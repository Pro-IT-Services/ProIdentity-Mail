package admin

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"proidentity-mail/internal/domain"
)

type SQLStore struct {
	db *sql.DB
}

func NewSQLStore(db *sql.DB) SQLStore {
	return SQLStore{db: db}
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

func (s SQLStore) CreateDomain(ctx context.Context, mailDomain domain.Domain) (domain.Domain, error) {
	result, err := s.db.ExecContext(ctx, `INSERT INTO domains(tenant_id, name, status, dkim_selector) VALUES (?, ?, 'pending', 'mail')`, mailDomain.TenantID, mailDomain.Name)
	if err != nil {
		return domain.Domain{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Domain{}, err
	}
	mailDomain.ID = uint64(id)
	mailDomain.Status = "pending"
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

func (s SQLStore) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	result, err := s.db.ExecContext(ctx, `INSERT INTO users(tenant_id, primary_domain_id, local_part, display_name, password_hash, status, quota_bytes) VALUES (?, ?, ?, ?, ?, 'active', ?)`,
		user.TenantID,
		user.PrimaryDomainID,
		user.LocalPart,
		user.DisplayName,
		user.PasswordHash,
		uint64(10737418240),
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
	user.QuotaBytes = 10737418240
	return user, nil
}

func (s SQLStore) ListUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant_id, primary_domain_id, local_part, display_name, status, quota_bytes, created_at, updated_at
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
		if err := rows.Scan(&user.ID, &user.TenantID, &user.PrimaryDomainID, &user.LocalPart, &user.DisplayName, &user.Status, &user.QuotaBytes, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
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
	priority := 10
	records := []domain.DNSRecord{
		{Type: "MX", Name: domainName, Value: "mail." + domainName, Priority: &priority},
		{Type: "TXT", Name: domainName, Value: "v=spf1 mx -all"},
		{Type: "TXT", Name: "_dmarc." + domainName, Value: "v=DMARC1; p=quarantine; rua=mailto:dmarc@" + domainName},
		{Type: "TXT", Name: "_mta-sts." + domainName, Value: "v=STSv1; id=2026050601"},
		{Type: "TXT", Name: "_smtp._tls." + domainName, Value: "v=TLSRPTv1; rua=mailto:tlsrpt@" + domainName},
	}
	if selector.Valid && dkimTXT.Valid && dkimTXT.String != "" {
		records = append(records, domain.DNSRecord{
			Type:  "TXT",
			Name:  fmt.Sprintf("%s._domainkey.%s", selector.String, domainName),
			Value: normalizeDKIMTXT(dkimTXT.String),
		})
	}
	return domain.DomainDNS{DomainID: domainID, Domain: domainName, Records: records}, nil
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
