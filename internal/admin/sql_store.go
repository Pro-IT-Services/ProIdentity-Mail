package admin

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
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

func (s SQLStore) ListQuarantineEvents(ctx context.Context) ([]domain.QuarantineEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant_id, user_id, domain_id, message_id, sender, recipient, verdict, action, scanner, CAST(symbols_json AS CHAR), status, resolved_at, resolution_note, created_at
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
		var resolvedAt sql.NullTime
		var resolutionNote sql.NullString
		if err := rows.Scan(&event.ID, &event.TenantID, &userID, &domainID, &messageID, &sender, &event.Recipient, &event.Verdict, &event.Action, &event.Scanner, &event.SymbolsJSON, &event.Status, &resolvedAt, &resolutionNote, &event.CreatedAt); err != nil {
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

func (s SQLStore) ResolveQuarantineEvent(ctx context.Context, eventID uint64, status, note string) (domain.QuarantineEvent, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.QuarantineEvent{}, err
	}
	defer tx.Rollback()

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

	event, err := scanQuarantineEvent(ctx, tx, eventID)
	if err != nil {
		return domain.QuarantineEvent{}, err
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
	var resolvedAt sql.NullTime
	var resolutionNote sql.NullString
	err := q.QueryRowContext(ctx, `
		SELECT id, tenant_id, user_id, domain_id, message_id, sender, recipient, verdict, action, scanner, CAST(symbols_json AS CHAR), status, resolved_at, resolution_note, created_at
		FROM quarantine_events
		WHERE id = ?`, eventID).Scan(&event.ID, &event.TenantID, &userID, &domainID, &messageID, &sender, &event.Recipient, &event.Verdict, &event.Action, &event.Scanner, &event.SymbolsJSON, &event.Status, &resolvedAt, &resolutionNote, &event.CreatedAt)
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
	if resolvedAt.Valid {
		event.ResolvedAt = &resolvedAt.Time
	}
	if resolutionNote.Valid {
		event.ResolutionNote = resolutionNote.String
	}
	return event, nil
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
