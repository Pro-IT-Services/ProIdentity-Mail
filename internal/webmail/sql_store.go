package webmail

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"

	"proidentity-mail/internal/domain"
	"proidentity-mail/internal/i18n"
	"proidentity-mail/internal/security"
)

type SQLAuthStore struct {
	db *sql.DB
}

func NewSQLAuthStore(db *sql.DB) SQLAuthStore {
	return SQLAuthStore{db: db}
}

func (s SQLAuthStore) VerifyUserPassword(ctx context.Context, email, password string) (bool, error) {
	var hash string
	err := s.db.QueryRowContext(ctx, `
		SELECT u.password_hash
		FROM users u
		JOIN domains d ON d.id = u.primary_domain_id
		WHERE CONCAT(u.local_part, '@', d.name) = ?
		  AND u.status = 'active'
		  AND d.status IN ('pending', 'active')
		LIMIT 1`, email).Scan(&hash)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return security.VerifyPassword(hash, password), nil
}

func (s SQLAuthStore) ListMailboxes(ctx context.Context, email string) ([]MailboxAccount, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var userID uint64
	var displayName string
	var localPart string
	var domainName string
	err := s.db.QueryRowContext(ctx, `
		SELECT u.id, COALESCE(NULLIF(u.display_name, ''), u.local_part), u.local_part, d.name
		FROM users u
		JOIN domains d ON d.id = u.primary_domain_id
		WHERE CONCAT(u.local_part, '@', d.name) = ?
		  AND u.mailbox_type = 'user'
		  AND u.status = 'active'
		  AND d.status IN ('pending', 'active')
		LIMIT 1`, email).Scan(&userID, &displayName, &localPart, &domainName)
	if err != nil {
		return nil, err
	}
	personalAddress := localPart + "@" + domainName
	mailboxes := []MailboxAccount{{
		ID:              personalAddress,
		Name:            displayName,
		Address:         personalAddress,
		Kind:            "personal",
		CanRead:         true,
		CanSendAs:       true,
		CanSendOnBehalf: true,
		CanManage:       true,
	}}
	rows, err := s.db.QueryContext(ctx, `
		SELECT CONCAT(shared.local_part, '@', shared_domain.name),
		       COALESCE(NULLIF(shared.display_name, ''), shared.local_part),
		       p.can_read, p.can_send_as, p.can_send_on_behalf, p.can_manage
		FROM shared_mailbox_permissions p
		JOIN users shared ON shared.id = p.shared_mailbox_id
		JOIN domains shared_domain ON shared_domain.id = shared.primary_domain_id
		WHERE p.user_id = ?
		  AND p.can_read = 1
		  AND shared.mailbox_type = 'shared'
		  AND shared.status = 'active'
		  AND shared_domain.status IN ('pending', 'active')
		ORDER BY shared.display_name, shared.local_part`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var mailbox MailboxAccount
		if err := rows.Scan(&mailbox.Address, &mailbox.Name, &mailbox.CanRead, &mailbox.CanSendAs, &mailbox.CanSendOnBehalf, &mailbox.CanManage); err != nil {
			return nil, err
		}
		mailbox.ID = mailbox.Address
		mailbox.Kind = "shared"
		mailboxes = append(mailboxes, mailbox)
	}
	return mailboxes, rows.Err()
}

func (s SQLAuthStore) ReportMessage(ctx context.Context, email, id, verdict string) error {
	var userID uint64
	var tenantID uint64
	err := s.db.QueryRowContext(ctx, `
		SELECT u.id, u.tenant_id
		FROM users u
		JOIN domains d ON d.id = u.primary_domain_id
		WHERE CONCAT(u.local_part, '@', d.name) = ?
		  AND u.status = 'active'
		  AND d.status IN ('pending', 'active')
		LIMIT 1`, email).Scan(&userID, &tenantID)
	if err != nil {
		return err
	}
	action := "message.report_ham"
	if verdict == "spam" {
		action = "message.report_spam"
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO audit_events(tenant_id, actor_type, actor_id, action, target_type, target_id, metadata_json)
		VALUES (?, 'user', ?, ?, 'message', ?, JSON_OBJECT('email', ?, 'verdict', ?))`,
		tenantID,
		userID,
		action,
		id,
		email,
		verdict,
	)
	return err
}

func (s SQLAuthStore) RecordUserAudit(ctx context.Context, email, action, targetType, targetID string, metadata map[string]any) error {
	email = strings.ToLower(strings.TrimSpace(email))
	var tenantID sql.NullInt64
	var userID sql.NullInt64
	err := s.db.QueryRowContext(ctx, `
		SELECT u.tenant_id, u.id
		FROM users u
		JOIN domains d ON d.id = u.primary_domain_id
		WHERE CONCAT(u.local_part, '@', d.name) = ?
		LIMIT 1`, email).Scan(&tenantID, &userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadata["email"] = email
	body, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	if targetID == "" {
		targetID = email
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO audit_events(tenant_id, actor_type, actor_id, action, target_type, target_id, metadata_json)
		VALUES (?, 'user', ?, ?, ?, ?, ?)`,
		nullInt64AuditArg(tenantID),
		nullInt64AuditArg(userID),
		action,
		targetType,
		targetID,
		string(body),
	)
	return err
}

func nullInt64AuditArg(value sql.NullInt64) any {
	if value.Valid {
		return value.Int64
	}
	return nil
}

func (s SQLAuthStore) ListContacts(ctx context.Context, email string) ([]Contact, error) {
	bookID, err := s.ensureAddressBook(ctx, email)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT uid, COALESCE(full_name, ''), COALESCE(email, '')
		FROM contact_objects
		WHERE address_book_id = ?
		ORDER BY full_name, email`, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	contacts := make([]Contact, 0)
	for rows.Next() {
		var contact Contact
		if err := rows.Scan(&contact.ID, &contact.Name, &contact.Email); err != nil {
			return nil, err
		}
		contacts = append(contacts, contact)
	}
	return contacts, rows.Err()
}

func (s SQLAuthStore) CreateContact(ctx context.Context, email string, contact Contact) (Contact, error) {
	bookID, err := s.ensureAddressBook(ctx, email)
	if err != nil {
		return Contact{}, err
	}
	uid := fmt.Sprintf("contact-%d", time.Now().UnixNano())
	href := uid + ".vcf"
	body := fmt.Sprintf("BEGIN:VCARD\r\nVERSION:3.0\r\nUID:%s\r\nFN:%s\r\nEMAIL:%s\r\nEND:VCARD\r\n", uid, contact.Name, contact.Email)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO contact_objects(address_book_id, uid, href, etag, vcard, full_name, email)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, bookID, uid, href, objectETag(body), body, contact.Name, contact.Email)
	if err != nil {
		return Contact{}, err
	}
	contact.ID = uid
	return contact, nil
}

func (s SQLAuthStore) UpdateContact(ctx context.Context, email, id string, contact Contact) (Contact, error) {
	bookID, err := s.ensureAddressBook(ctx, email)
	if err != nil {
		return Contact{}, err
	}
	body := fmt.Sprintf("BEGIN:VCARD\r\nVERSION:3.0\r\nUID:%s\r\nFN:%s\r\nEMAIL:%s\r\nEND:VCARD\r\n", id, contact.Name, contact.Email)
	result, err := s.db.ExecContext(ctx, `
		UPDATE contact_objects
		SET etag = ?, vcard = ?, full_name = ?, email = ?, updated_at = CURRENT_TIMESTAMP
		WHERE address_book_id = ? AND uid = ?`, objectETag(body), body, contact.Name, contact.Email, bookID, id)
	if err != nil {
		return Contact{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Contact{}, err
	}
	if affected == 0 {
		return Contact{}, sql.ErrNoRows
	}
	contact.ID = id
	return contact, nil
}

func (s SQLAuthStore) DeleteContact(ctx context.Context, email, id string) error {
	bookID, err := s.ensureAddressBook(ctx, email)
	if err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM contact_objects WHERE address_book_id = ? AND uid = ?`, bookID, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s SQLAuthStore) ListFilters(ctx context.Context, email string) ([]MailFilter, error) {
	_, userID, err := s.userIDs(ctx, email)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, field, operator, value, action, COALESCE(folder, ''), enabled, created_at, updated_at
		FROM mail_filters
		WHERE user_id = ?
		ORDER BY enabled DESC, name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	filters := make([]MailFilter, 0)
	for rows.Next() {
		var filter MailFilter
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(&filter.ID, &filter.Name, &filter.Field, &filter.Operator, &filter.Value, &filter.Action, &filter.Folder, &filter.Enabled, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		filter.CreatedAt = createdAt.Format(time.RFC3339)
		filter.UpdatedAt = updatedAt.Format(time.RFC3339)
		filters = append(filters, filter)
	}
	return filters, rows.Err()
}

func (s SQLAuthStore) ListFilterMailboxes(ctx context.Context) ([]MailboxFilterSet, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT CONCAT(u.local_part, '@', d.name) AS email,
		       f.id, f.name, f.field, f.operator, f.value, f.action, COALESCE(f.folder, ''), f.enabled, f.created_at, f.updated_at
		FROM mail_filters f
		JOIN users u ON u.id = f.user_id
		JOIN domains d ON d.id = u.primary_domain_id
		WHERE u.status = 'active'
		  AND u.mailbox_type = 'user'
		  AND d.status IN ('pending', 'active')
		ORDER BY email, f.enabled DESC, f.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sets := make([]MailboxFilterSet, 0)
	index := map[string]int{}
	for rows.Next() {
		var email string
		var filter MailFilter
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(&email, &filter.ID, &filter.Name, &filter.Field, &filter.Operator, &filter.Value, &filter.Action, &filter.Folder, &filter.Enabled, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		filter.CreatedAt = createdAt.Format(time.RFC3339)
		filter.UpdatedAt = updatedAt.Format(time.RFC3339)
		pos, ok := index[email]
		if !ok {
			pos = len(sets)
			index[email] = pos
			sets = append(sets, MailboxFilterSet{Email: email})
		}
		sets[pos].Filters = append(sets[pos].Filters, filter)
	}
	return sets, rows.Err()
}

func (s SQLAuthStore) CreateFilter(ctx context.Context, email string, filter MailFilter) (MailFilter, error) {
	if err := validateFilter(filter); err != nil {
		return MailFilter{}, err
	}
	tenantID, userID, err := s.userIDs(ctx, email)
	if err != nil {
		return MailFilter{}, err
	}
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO mail_filters(tenant_id, user_id, name, field, operator, value, action, folder, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?)`,
		tenantID, userID, filter.Name, filter.Field, filter.Operator, filter.Value, filter.Action, filter.Folder, filter.Enabled)
	if err != nil {
		return MailFilter{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return MailFilter{}, err
	}
	filter.ID = fmt.Sprintf("%d", id)
	return filter, nil
}

func (s SQLAuthStore) UpdateFilter(ctx context.Context, email, id string, filter MailFilter) (MailFilter, error) {
	if err := validateFilter(filter); err != nil {
		return MailFilter{}, err
	}
	_, userID, err := s.userIDs(ctx, email)
	if err != nil {
		return MailFilter{}, err
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE mail_filters
		SET name = ?, field = ?, operator = ?, value = ?, action = ?, folder = NULLIF(?, ''), enabled = ?
		WHERE id = ? AND user_id = ?`,
		filter.Name, filter.Field, filter.Operator, filter.Value, filter.Action, filter.Folder, filter.Enabled, id, userID)
	if err != nil {
		return MailFilter{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return MailFilter{}, err
	}
	if affected == 0 {
		return MailFilter{}, sql.ErrNoRows
	}
	filter.ID = id
	return filter, nil
}

func (s SQLAuthStore) DeleteFilter(ctx context.Context, email, id string) error {
	_, userID, err := s.userIDs(ctx, email)
	if err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM mail_filters WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func validateFilter(filter MailFilter) error {
	if strings.TrimSpace(filter.Name) == "" || strings.TrimSpace(filter.Value) == "" {
		return errors.New("filter name and value are required")
	}
	switch filter.Field {
	case "from", "to", "subject", "body":
	default:
		return errors.New("filter field must be from, to, subject, or body")
	}
	switch filter.Operator {
	case "contains", "equals", "starts_with", "ends_with":
	default:
		return errors.New("filter operator must be contains, equals, starts_with, or ends_with")
	}
	switch filter.Action {
	case "move", "mark_spam", "delete", "keep":
	default:
		return errors.New("filter action must be move, mark_spam, delete, or keep")
	}
	if filter.Action == "move" && strings.TrimSpace(filter.Folder) == "" {
		return errors.New("move filters require a destination folder")
	}
	return nil
}

func (s SQLAuthStore) ListCalendarEvents(ctx context.Context, email string) ([]CalendarEvent, error) {
	calendarID, err := s.ensureCalendar(ctx, email)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT uid, icalendar, starts_at, ends_at
		FROM calendar_objects
		WHERE calendar_id = ?
		ORDER BY COALESCE(starts_at, created_at), uid`, calendarID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]CalendarEvent, 0)
	for rows.Next() {
		var event CalendarEvent
		var body string
		var startsAt sql.NullTime
		var endsAt sql.NullTime
		if err := rows.Scan(&event.ID, &body, &startsAt, &endsAt); err != nil {
			return nil, err
		}
		event.Title = icalValue(body, "SUMMARY")
		if startsAt.Valid {
			event.StartsAt = startsAt.Time
		}
		if endsAt.Valid {
			event.EndsAt = endsAt.Time
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s SQLAuthStore) CreateCalendarEvent(ctx context.Context, email string, event CalendarEvent) (CalendarEvent, error) {
	calendarID, err := s.ensureCalendar(ctx, email)
	if err != nil {
		return CalendarEvent{}, err
	}
	uid := fmt.Sprintf("event-%d", time.Now().UnixNano())
	href := uid + ".ics"
	body := fmt.Sprintf("BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//ProIdentity//Mail//EN\r\nBEGIN:VEVENT\r\nUID:%s\r\nSUMMARY:%s\r\nDTSTART:%s\r\nDTEND:%s\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n", uid, event.Title, event.StartsAt.UTC().Format("20060102T150405Z"), event.EndsAt.UTC().Format("20060102T150405Z"))
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO calendar_objects(calendar_id, uid, href, etag, icalendar, starts_at, ends_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, calendarID, uid, href, objectETag(body), body, event.StartsAt.UTC(), event.EndsAt.UTC())
	if err != nil {
		return CalendarEvent{}, err
	}
	event.ID = uid
	return event, nil
}

func (s SQLAuthStore) UpdateCalendarEvent(ctx context.Context, email, id string, event CalendarEvent) (CalendarEvent, error) {
	calendarID, err := s.ensureCalendar(ctx, email)
	if err != nil {
		return CalendarEvent{}, err
	}
	body := fmt.Sprintf("BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//ProIdentity//Mail//EN\r\nBEGIN:VEVENT\r\nUID:%s\r\nSUMMARY:%s\r\nDTSTART:%s\r\nDTEND:%s\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n", id, event.Title, event.StartsAt.UTC().Format("20060102T150405Z"), event.EndsAt.UTC().Format("20060102T150405Z"))
	result, err := s.db.ExecContext(ctx, `
		UPDATE calendar_objects
		SET etag = ?, icalendar = ?, starts_at = ?, ends_at = ?, updated_at = CURRENT_TIMESTAMP
		WHERE calendar_id = ? AND uid = ?`, objectETag(body), body, event.StartsAt.UTC(), event.EndsAt.UTC(), calendarID, id)
	if err != nil {
		return CalendarEvent{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return CalendarEvent{}, err
	}
	if affected == 0 {
		return CalendarEvent{}, sql.ErrNoRows
	}
	event.ID = id
	return event, nil
}

func (s SQLAuthStore) DeleteCalendarEvent(ctx context.Context, email, id string) error {
	calendarID, err := s.ensureCalendar(ctx, email)
	if err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM calendar_objects WHERE calendar_id = ? AND uid = ?`, calendarID, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s SQLAuthStore) ChangePassword(ctx context.Context, email, newPassword string) error {
	hash, err := security.HashPassword(newPassword)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE users u
		JOIN domains d ON d.id = u.primary_domain_id
		SET u.password_hash = ?
		WHERE CONCAT(u.local_part, '@', d.name) = ?
		  AND u.status = 'active'
		  AND d.status IN ('pending', 'active')`, hash, email)
	return err
}

func (s SQLAuthStore) GetProfile(ctx context.Context, email string) (UserProfile, error) {
	var profile UserProfile
	err := s.db.QueryRowContext(ctx, `
		SELECT CONCAT(u.local_part, '@', d.name),
		       COALESCE(NULLIF(settings.first_name, ''), ''),
		       COALESCE(NULLIF(settings.last_name, ''), ''),
		       COALESCE(NULLIF(u.display_name, ''), u.local_part),
		       COALESCE(settings.signature_html, ''),
		       COALESCE(settings.signature_auto_add, 0),
		       COALESCE(NULLIF(settings.language, ''), NULLIF(server.default_language, ''), 'en')
		FROM users u
		JOIN domains d ON d.id = u.primary_domain_id
		LEFT JOIN webmail_user_settings settings ON settings.user_id = u.id
		LEFT JOIN mail_server_settings server ON server.id = 1
		WHERE CONCAT(u.local_part, '@', d.name) = ?
		  AND u.mailbox_type = 'user'
		  AND u.status = 'active'
		  AND d.status IN ('pending', 'active')
		LIMIT 1`, email).Scan(&profile.Email, &profile.FirstName, &profile.LastName, &profile.DisplayName, &profile.SignatureHTML, &profile.SignatureAutoAdd, &profile.Language)
	profile.Language = normalizeProfileLanguage(profile.Language)
	return profile, err
}

func (s SQLAuthStore) UpdateProfile(ctx context.Context, email string, profile UserProfile) (UserProfile, error) {
	tenantID, userID, err := s.userIDs(ctx, email)
	if err != nil {
		return UserProfile{}, err
	}
	profile.Language = normalizeProfileLanguage(profile.Language)
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO webmail_user_settings(tenant_id, user_id, signature_html, signature_auto_add, language)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		  signature_html = VALUES(signature_html),
		  signature_auto_add = VALUES(signature_auto_add),
		  language = VALUES(language)`,
		tenantID, userID, profile.SignatureHTML, profile.SignatureAutoAdd, profile.Language); err != nil {
		return UserProfile{}, err
	}
	return s.GetProfile(ctx, email)
}

func (s SQLAuthStore) ListContentTrust(ctx context.Context, email string) ([]ContentTrustEntry, error) {
	_, userID, err := s.userIDs(ctx, email)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, scope, value, DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ')
		FROM webmail_content_trust
		WHERE user_id = ?
		ORDER BY scope, value`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	entries := make([]ContentTrustEntry, 0)
	for rows.Next() {
		var id uint64
		var entry ContentTrustEntry
		if err := rows.Scan(&id, &entry.Scope, &entry.Value, &entry.CreatedAt); err != nil {
			return nil, err
		}
		entry.ID = strconv.FormatUint(id, 10)
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (s SQLAuthStore) AddContentTrust(ctx context.Context, email string, entry ContentTrustEntry) (ContentTrustEntry, error) {
	tenantID, userID, err := s.userIDs(ctx, email)
	if err != nil {
		return ContentTrustEntry{}, err
	}
	normalized, err := normalizeContentTrustEntry(entry)
	if err != nil {
		return ContentTrustEntry{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO webmail_content_trust(tenant_id, user_id, scope, value)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE updated_at = current_timestamp()`,
		tenantID, userID, normalized.Scope, normalized.Value); err != nil {
		return ContentTrustEntry{}, err
	}
	var id uint64
	if err := s.db.QueryRowContext(ctx, `
		SELECT id, DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ')
		FROM webmail_content_trust
		WHERE user_id = ? AND scope = ? AND value = ?
		LIMIT 1`, userID, normalized.Scope, normalized.Value).Scan(&id, &normalized.CreatedAt); err != nil {
		return ContentTrustEntry{}, err
	}
	normalized.ID = strconv.FormatUint(id, 10)
	return normalized, nil
}

func (s SQLAuthStore) GetMailboxSecurity(ctx context.Context, email string) (MailboxSecurity, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var state MailboxSecurity
	var mailboxMFAEnabled bool
	err := s.db.QueryRowContext(ctx, `
		SELECT CONCAT(u.local_part, '@', d.name),
		       COALESCE(ms.mailbox_mfa_enabled, 1),
		       COALESCE(ms.force_mailbox_mfa, 0),
		       COALESCE(mfa.totp_enabled, 0)
		FROM users u
		JOIN domains d ON d.id = u.primary_domain_id
		LEFT JOIN mail_server_settings ms ON ms.id = 1
		LEFT JOIN user_mfa_settings mfa ON mfa.user_id = u.id
		WHERE CONCAT(u.local_part, '@', d.name) = ?
		  AND u.mailbox_type = 'user'
		  AND u.status = 'active'
		  AND d.status IN ('pending', 'active')
		LIMIT 1`, email).Scan(&state.Email, &mailboxMFAEnabled, &state.ForceMFA, &state.TOTPEnabled)
	if err != nil {
		return MailboxSecurity{}, err
	}
	state.MFAAvailable = mailboxMFAEnabled
	state.MFAEnabled = state.TOTPEnabled
	state.SetupNeeded = state.MFAAvailable && state.ForceMFA && !state.MFAEnabled
	return state, nil
}

func (s SQLAuthStore) GetAdminMFASettings(ctx context.Context) (domain.AdminMFASettings, error) {
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
		return domain.AdminMFASettings{ID: 1, ProIdentityBaseURL: proIdentityAuthServiceURL, ProIdentityTimeoutSeconds: 90}, nil
	}
	if err != nil {
		return domain.AdminMFASettings{}, err
	}
	settings.ProIdentityAPIKey = apiKey.String
	settings.ProIdentityBaseURL = proIdentityAuthServiceURL
	if settings.ProIdentityTimeoutSeconds <= 0 {
		settings.ProIdentityTimeoutSeconds = 90
	}
	return settings, nil
}

func (s SQLAuthStore) CreateMailboxMFAChallenge(ctx context.Context, challenge MailboxMFAChallenge) error {
	tenantID, userID, err := s.userIDs(ctx, challenge.Email)
	if err != nil {
		return err
	}
	_, _ = s.db.ExecContext(ctx, `DELETE FROM user_mfa_challenges WHERE expires_at < UTC_TIMESTAMP()`)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO user_mfa_challenges(token, tenant_id, user_id, purpose, provider, request_id, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		strings.TrimSpace(challenge.Token),
		tenantID,
		userID,
		normalizeMFAPurpose(challenge.Purpose),
		normalizeMFAProvider(challenge.Provider),
		strings.TrimSpace(challenge.RequestID),
		challenge.ExpiresAt.UTC(),
	)
	return err
}

func (s SQLAuthStore) GetMailboxMFAChallenge(ctx context.Context, token string) (MailboxMFAChallenge, error) {
	var challenge MailboxMFAChallenge
	err := s.db.QueryRowContext(ctx, `
		SELECT c.token, CONCAT(u.local_part, '@', d.name), c.purpose, COALESCE(NULLIF(c.provider, ''), 'totp'), COALESCE(c.request_id, ''), c.expires_at
		FROM user_mfa_challenges c
		JOIN users u ON u.id = c.user_id
		JOIN domains d ON d.id = u.primary_domain_id
		WHERE c.token = ?
		LIMIT 1`, strings.TrimSpace(token)).Scan(&challenge.Token, &challenge.Email, &challenge.Purpose, &challenge.Provider, &challenge.RequestID, &challenge.ExpiresAt)
	return challenge, err
}

func (s SQLAuthStore) DeleteMailboxMFAChallenge(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM user_mfa_challenges WHERE token = ?`, strings.TrimSpace(token))
	return err
}

func (s SQLAuthStore) BeginMailboxTOTPEnrollment(ctx context.Context, email string) (MailboxTOTPEnrollment, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	tenantID, userID, err := s.userIDs(ctx, email)
	if err != nil {
		return MailboxTOTPEnrollment{}, err
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "ProIdentity Mail",
		AccountName: email,
		SecretSize:  20,
		Period:      30,
		Digits:      6,
	})
	if err != nil {
		return MailboxTOTPEnrollment{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO user_mfa_settings(tenant_id, user_id, pending_totp_secret)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE pending_totp_secret = VALUES(pending_totp_secret)`,
		tenantID, userID, key.Secret()); err != nil {
		return MailboxTOTPEnrollment{}, err
	}
	qr, err := totpQRCodeDataURL(key.URL())
	if err != nil {
		return MailboxTOTPEnrollment{}, err
	}
	return MailboxTOTPEnrollment{Email: email, OTPAuthURL: key.URL(), QRDataURL: qr}, nil
}

func (s SQLAuthStore) VerifyMailboxTOTPEnrollment(ctx context.Context, email, code string) (MailboxSecurity, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	tenantID, userID, err := s.userIDs(ctx, email)
	if err != nil {
		return MailboxSecurity{}, err
	}
	var secret string
	err = s.db.QueryRowContext(ctx, `SELECT pending_totp_secret FROM user_mfa_settings WHERE user_id = ?`, userID).Scan(&secret)
	if err != nil {
		return MailboxSecurity{}, err
	}
	if !totp.Validate(strings.TrimSpace(code), secret) {
		return MailboxSecurity{}, errors.New("invalid authenticator code")
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO user_mfa_settings(tenant_id, user_id, totp_enabled, totp_secret, pending_totp_secret)
		VALUES (?, ?, TRUE, ?, '')
		ON DUPLICATE KEY UPDATE
		  totp_enabled = TRUE,
		  totp_secret = VALUES(totp_secret),
		  pending_totp_secret = ''`,
		tenantID, userID, secret)
	if err != nil {
		return MailboxSecurity{}, err
	}
	return s.GetMailboxSecurity(ctx, email)
}

func (s SQLAuthStore) VerifyMailboxTOTPCode(ctx context.Context, email, code string) (bool, error) {
	_, userID, err := s.userIDs(ctx, email)
	if err != nil {
		return false, err
	}
	var secret string
	err = s.db.QueryRowContext(ctx, `
		SELECT totp_secret
		FROM user_mfa_settings
		WHERE user_id = ? AND totp_enabled = TRUE
		LIMIT 1`, userID).Scan(&secret)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return totp.Validate(strings.TrimSpace(code), secret), nil
}

func (s SQLAuthStore) ListAppPasswords(ctx context.Context, email string) ([]AppPassword, error) {
	_, userID, err := s.userIDs(ctx, email)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, protocols, status, last_used_at, last_used_protocol, created_at
		FROM user_app_passwords
		WHERE user_id = ?
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	passwords := make([]AppPassword, 0)
	for rows.Next() {
		var password AppPassword
		var id uint64
		var protocols string
		var lastUsedAt sql.NullTime
		if err := rows.Scan(&id, &password.Name, &protocols, &password.Status, &lastUsedAt, &password.LastUsedProtocol, &password.CreatedAt); err != nil {
			return nil, err
		}
		password.ID = strconv.FormatUint(id, 10)
		password.Protocols = normalizeAppPasswordProtocols(strings.Split(protocols, ","))
		if lastUsedAt.Valid {
			password.LastUsedAt = &lastUsedAt.Time
		}
		passwords = append(passwords, password)
	}
	return passwords, rows.Err()
}

func (s SQLAuthStore) CreateAppPassword(ctx context.Context, email string, req AppPassword) (AppPassword, error) {
	tenantID, userID, err := s.userIDs(ctx, email)
	if err != nil {
		return AppPassword{}, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return AppPassword{}, errors.New("name is required")
	}
	if len(name) > 120 {
		name = name[:120]
	}
	protocols := normalizeAppPasswordProtocols(req.Protocols)
	secret, err := newAppPasswordSecret()
	if err != nil {
		return AppPassword{}, err
	}
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO user_app_passwords(tenant_id, user_id, name, secret_sha256, protocols, status)
		VALUES (?, ?, ?, ?, ?, 'active')`,
		tenantID,
		userID,
		name,
		appPasswordFingerprint(secret),
		strings.Join(protocols, ","),
	)
	if err != nil {
		return AppPassword{}, err
	}
	id, _ := result.LastInsertId()
	return AppPassword{
		ID:        strconv.FormatInt(id, 10),
		Name:      name,
		Protocols: protocols,
		Secret:    secret,
		Status:    "active",
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (s SQLAuthStore) RevokeAppPassword(ctx context.Context, email, id string) error {
	_, userID, err := s.userIDs(ctx, email)
	if err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE user_app_passwords
		SET status = 'revoked', revoked_at = UTC_TIMESTAMP()
		WHERE user_id = ? AND id = ? AND status = 'active'`, userID, strings.TrimSpace(id))
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s SQLAuthStore) VerifyProtocolPassword(ctx context.Context, email, password, protocol string) (bool, error) {
	protocol = normalizeProtocol(protocol)
	if protocol == "" || protocol == "webmail" {
		return false, nil
	}
	email = strings.ToLower(strings.TrimSpace(email))
	state, err := s.GetMailboxSecurity(ctx, email)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	if err == nil && !state.ForceMFA && !state.MFAEnabled {
		ok, err := s.VerifyUserPassword(ctx, email, password)
		if err != nil || ok {
			return ok, err
		}
	}
	_, userID, err := s.userIDs(ctx, email)
	if err != nil {
		return false, err
	}
	fingerprint := appPasswordFingerprint(password)
	var id uint64
	var protocols string
	err = s.db.QueryRowContext(ctx, `
		SELECT id, protocols
		FROM user_app_passwords
		WHERE user_id = ? AND secret_sha256 = ? AND status = 'active'
		LIMIT 1`, userID, fingerprint).Scan(&id, &protocols)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !protocolAllowed(strings.Split(protocols, ","), protocol) {
		return false, nil
	}
	_, _ = s.db.ExecContext(ctx, `
		UPDATE user_app_passwords
		SET last_used_at = UTC_TIMESTAMP(), last_used_protocol = ?
		WHERE id = ?`, protocol, id)
	return true, nil
}

func (s SQLAuthStore) ensureAddressBook(ctx context.Context, email string) (uint64, error) {
	tenantID, userID, err := s.userIDs(ctx, email)
	if err != nil {
		return 0, err
	}
	if _, err := s.db.ExecContext(ctx, `INSERT IGNORE INTO address_books(tenant_id, user_id, slug, display_name) VALUES (?, ?, 'default', 'Default Address Book')`, tenantID, userID); err != nil {
		return 0, err
	}
	var id uint64
	err = s.db.QueryRowContext(ctx, `SELECT id FROM address_books WHERE user_id = ? AND slug = 'default'`, userID).Scan(&id)
	return id, err
}

func (s SQLAuthStore) ensureCalendar(ctx context.Context, email string) (uint64, error) {
	tenantID, userID, err := s.userIDs(ctx, email)
	if err != nil {
		return 0, err
	}
	if _, err := s.db.ExecContext(ctx, `INSERT IGNORE INTO calendars(tenant_id, user_id, slug, display_name) VALUES (?, ?, 'default', 'Default Calendar')`, tenantID, userID); err != nil {
		return 0, err
	}
	var id uint64
	err = s.db.QueryRowContext(ctx, `SELECT id FROM calendars WHERE user_id = ? AND slug = 'default'`, userID).Scan(&id)
	return id, err
}

func (s SQLAuthStore) userIDs(ctx context.Context, email string) (uint64, uint64, error) {
	var tenantID, userID uint64
	err := s.db.QueryRowContext(ctx, `
		SELECT u.tenant_id, u.id
		FROM users u
		JOIN domains d ON d.id = u.primary_domain_id
		WHERE CONCAT(u.local_part, '@', d.name) = ?
		  AND u.status = 'active'
		  AND d.status IN ('pending', 'active')
		LIMIT 1`, email).Scan(&tenantID, &userID)
	return tenantID, userID, err
}

type CompositeStore struct {
	Auth    AuthStore
	Mailbox MaildirStore
	Sender  SMTPSender
	Learner SpamLearner
}

type AuthStore interface {
	VerifyUserPassword(ctx context.Context, email, password string) (bool, error)
	ReportMessage(ctx context.Context, email, id, verdict string) error
}

type SpamLearner interface {
	Learn(ctx context.Context, path, verdict string) error
}

func (s CompositeStore) VerifyUserPassword(ctx context.Context, email, password string) (bool, error) {
	return s.Auth.VerifyUserPassword(ctx, email, password)
}

func (s CompositeStore) VerifyProtocolPassword(ctx context.Context, email, password, protocol string) (bool, error) {
	if store, ok := s.Auth.(interface {
		VerifyProtocolPassword(context.Context, string, string, string) (bool, error)
	}); ok {
		return store.VerifyProtocolPassword(ctx, email, password, protocol)
	}
	if normalizeProtocol(protocol) == "webmail" {
		return false, nil
	}
	return s.Auth.VerifyUserPassword(ctx, email, password)
}

func (s CompositeStore) ListMailboxes(ctx context.Context, email string) ([]MailboxAccount, error) {
	if store, ok := s.Auth.(interface {
		ListMailboxes(context.Context, string) ([]MailboxAccount, error)
	}); ok {
		return store.ListMailboxes(ctx, email)
	}
	return []MailboxAccount{{
		ID:              email,
		Name:            strings.Split(email, "@")[0],
		Address:         email,
		Kind:            "personal",
		CanRead:         true,
		CanSendAs:       true,
		CanSendOnBehalf: true,
		CanManage:       true,
	}}, nil
}

func (s CompositeStore) ListRecentMessages(ctx context.Context, email string, limit int) ([]MessageSummary, error) {
	return s.Mailbox.ListRecentMessages(ctx, email, limit)
}

func (s CompositeStore) ListMessages(ctx context.Context, email, folder string, limit int) ([]MessageSummary, error) {
	return s.Mailbox.ListMessages(ctx, email, folder, limit)
}

func (s CompositeStore) GetMessage(ctx context.Context, email, id string) (MessageDetail, error) {
	return s.Mailbox.GetMessage(ctx, email, id)
}

func (s CompositeStore) MarkMessageRead(ctx context.Context, email, id string) (MessageDetail, error) {
	return s.Mailbox.MarkMessageRead(ctx, email, id)
}

func (s CompositeStore) SendMessage(ctx context.Context, message OutboundMessage) error {
	if err := s.Sender.Send(ctx, message); err != nil {
		return err
	}
	return s.Mailbox.SaveSentMessage(ctx, message)
}

func (s CompositeStore) SaveSentMessage(ctx context.Context, message OutboundMessage) error {
	return s.Mailbox.SaveSentMessage(ctx, message)
}

func (s CompositeStore) SaveDraftMessage(ctx context.Context, message OutboundMessage) (string, error) {
	return s.Mailbox.SaveDraftMessage(ctx, message)
}

func (s CompositeStore) ReportMessage(ctx context.Context, email, id, verdict string) error {
	path, err := s.Mailbox.MessagePath(ctx, email, id)
	if err != nil {
		return err
	}
	if s.Learner != nil {
		if err := s.Learner.Learn(ctx, path, verdict); err != nil {
			return fmt.Errorf("rspamd learning failed verdict=%q: %w", verdict, err)
		}
	}
	if verdict == "spam" {
		if err := s.Mailbox.MoveMessage(ctx, email, id, "spam"); err != nil {
			return fmt.Errorf("move message to spam: %w", err)
		}
	} else if verdict == "ham" {
		if err := s.Mailbox.MoveMessage(ctx, email, id, "inbox"); err != nil {
			return fmt.Errorf("move message to inbox: %w", err)
		}
	}
	return s.Auth.ReportMessage(ctx, email, id, verdict)
}

func (s CompositeStore) MoveMessage(ctx context.Context, email, id, folder string) error {
	return s.Mailbox.MoveMessage(ctx, email, id, folder)
}

func (s CompositeStore) DeleteMessage(ctx context.Context, email, id string) error {
	return s.Mailbox.DeleteMessage(ctx, email, id)
}

func (s CompositeStore) ListFolders(ctx context.Context, email string) ([]MailFolder, error) {
	return s.Mailbox.ListFolders(ctx, email)
}

func (s CompositeStore) CreateFolder(ctx context.Context, email, name string) (MailFolder, error) {
	return s.Mailbox.CreateFolder(ctx, email, name)
}

func (s CompositeStore) DeleteFolder(ctx context.Context, email, name string) error {
	return s.Mailbox.DeleteFolder(ctx, email, name)
}

func (s CompositeStore) ListFilters(ctx context.Context, email string) ([]MailFilter, error) {
	if store, ok := s.Auth.(interface {
		ListFilters(context.Context, string) ([]MailFilter, error)
	}); ok {
		return store.ListFilters(ctx, email)
	}
	return nil, sql.ErrNoRows
}

func (s CompositeStore) CreateFilter(ctx context.Context, email string, filter MailFilter) (MailFilter, error) {
	if store, ok := s.Auth.(interface {
		CreateFilter(context.Context, string, MailFilter) (MailFilter, error)
	}); ok {
		created, err := store.CreateFilter(ctx, email, filter)
		if err != nil {
			return MailFilter{}, err
		}
		if err := s.SyncFilters(ctx, email); err != nil {
			return MailFilter{}, err
		}
		return created, nil
	}
	return MailFilter{}, sql.ErrNoRows
}

func (s CompositeStore) UpdateFilter(ctx context.Context, email, id string, filter MailFilter) (MailFilter, error) {
	if store, ok := s.Auth.(interface {
		UpdateFilter(context.Context, string, string, MailFilter) (MailFilter, error)
	}); ok {
		updated, err := store.UpdateFilter(ctx, email, id, filter)
		if err != nil {
			return MailFilter{}, err
		}
		if err := s.SyncFilters(ctx, email); err != nil {
			return MailFilter{}, err
		}
		return updated, nil
	}
	return MailFilter{}, sql.ErrNoRows
}

func (s CompositeStore) DeleteFilter(ctx context.Context, email, id string) error {
	if store, ok := s.Auth.(interface {
		DeleteFilter(context.Context, string, string) error
	}); ok {
		if err := store.DeleteFilter(ctx, email, id); err != nil {
			return err
		}
		return s.SyncFilters(ctx, email)
	}
	return sql.ErrNoRows
}

func (s CompositeStore) SyncFilters(ctx context.Context, email string) error {
	filters, err := s.ListFilters(ctx, email)
	if err != nil {
		return err
	}
	return s.Mailbox.SyncFilters(ctx, email, filters)
}

func (s CompositeStore) SyncAllFilters(ctx context.Context) error {
	store, ok := s.Auth.(interface {
		ListFilterMailboxes(context.Context) ([]MailboxFilterSet, error)
	})
	if !ok {
		return nil
	}
	sets, err := store.ListFilterMailboxes(ctx)
	if err != nil {
		return err
	}
	for _, set := range sets {
		if err := s.Mailbox.SyncFilters(ctx, set.Email, set.Filters); err != nil {
			return fmt.Errorf("sync filters for %s: %w", set.Email, err)
		}
	}
	return nil
}

func (s CompositeStore) ListContacts(ctx context.Context, email string) ([]Contact, error) {
	if store, ok := s.Auth.(interface {
		ListContacts(context.Context, string) ([]Contact, error)
	}); ok {
		return store.ListContacts(ctx, email)
	}
	return nil, sql.ErrNoRows
}

func (s CompositeStore) CreateContact(ctx context.Context, email string, contact Contact) (Contact, error) {
	if store, ok := s.Auth.(interface {
		CreateContact(context.Context, string, Contact) (Contact, error)
	}); ok {
		return store.CreateContact(ctx, email, contact)
	}
	return Contact{}, sql.ErrNoRows
}

func (s CompositeStore) UpdateContact(ctx context.Context, email, id string, contact Contact) (Contact, error) {
	if store, ok := s.Auth.(interface {
		UpdateContact(context.Context, string, string, Contact) (Contact, error)
	}); ok {
		return store.UpdateContact(ctx, email, id, contact)
	}
	return Contact{}, sql.ErrNoRows
}

func (s CompositeStore) DeleteContact(ctx context.Context, email, id string) error {
	if store, ok := s.Auth.(interface {
		DeleteContact(context.Context, string, string) error
	}); ok {
		return store.DeleteContact(ctx, email, id)
	}
	return sql.ErrNoRows
}

func (s CompositeStore) ListCalendarEvents(ctx context.Context, email string) ([]CalendarEvent, error) {
	if store, ok := s.Auth.(interface {
		ListCalendarEvents(context.Context, string) ([]CalendarEvent, error)
	}); ok {
		return store.ListCalendarEvents(ctx, email)
	}
	return nil, sql.ErrNoRows
}

func (s CompositeStore) CreateCalendarEvent(ctx context.Context, email string, event CalendarEvent) (CalendarEvent, error) {
	if store, ok := s.Auth.(interface {
		CreateCalendarEvent(context.Context, string, CalendarEvent) (CalendarEvent, error)
	}); ok {
		return store.CreateCalendarEvent(ctx, email, event)
	}
	return CalendarEvent{}, sql.ErrNoRows
}

func (s CompositeStore) UpdateCalendarEvent(ctx context.Context, email, id string, event CalendarEvent) (CalendarEvent, error) {
	if store, ok := s.Auth.(interface {
		UpdateCalendarEvent(context.Context, string, string, CalendarEvent) (CalendarEvent, error)
	}); ok {
		return store.UpdateCalendarEvent(ctx, email, id, event)
	}
	return CalendarEvent{}, sql.ErrNoRows
}

func (s CompositeStore) DeleteCalendarEvent(ctx context.Context, email, id string) error {
	if store, ok := s.Auth.(interface {
		DeleteCalendarEvent(context.Context, string, string) error
	}); ok {
		return store.DeleteCalendarEvent(ctx, email, id)
	}
	return sql.ErrNoRows
}

func (s CompositeStore) GetProfile(ctx context.Context, email string) (UserProfile, error) {
	if store, ok := s.Auth.(interface {
		GetProfile(context.Context, string) (UserProfile, error)
	}); ok {
		return store.GetProfile(ctx, email)
	}
	return UserProfile{Email: email, DisplayName: strings.Split(email, "@")[0], Language: "en"}, nil
}

func (s CompositeStore) UpdateProfile(ctx context.Context, email string, profile UserProfile) (UserProfile, error) {
	if store, ok := s.Auth.(interface {
		UpdateProfile(context.Context, string, UserProfile) (UserProfile, error)
	}); ok {
		return store.UpdateProfile(ctx, email, profile)
	}
	profile.Email = email
	profile.Language = normalizeProfileLanguage(profile.Language)
	return profile, nil
}

func (s CompositeStore) ChangePassword(ctx context.Context, email, newPassword string) error {
	if store, ok := s.Auth.(interface {
		ChangePassword(context.Context, string, string) error
	}); ok {
		return store.ChangePassword(ctx, email, newPassword)
	}
	return sql.ErrNoRows
}

func (s CompositeStore) ListContentTrust(ctx context.Context, email string) ([]ContentTrustEntry, error) {
	if store, ok := s.Auth.(interface {
		ListContentTrust(context.Context, string) ([]ContentTrustEntry, error)
	}); ok {
		return store.ListContentTrust(ctx, email)
	}
	return nil, nil
}

func (s CompositeStore) AddContentTrust(ctx context.Context, email string, entry ContentTrustEntry) (ContentTrustEntry, error) {
	if store, ok := s.Auth.(interface {
		AddContentTrust(context.Context, string, ContentTrustEntry) (ContentTrustEntry, error)
	}); ok {
		return store.AddContentTrust(ctx, email, entry)
	}
	return normalizeContentTrustEntry(entry)
}

func (s CompositeStore) GetMailboxSecurity(ctx context.Context, email string) (MailboxSecurity, error) {
	if store, ok := s.Auth.(interface {
		GetMailboxSecurity(context.Context, string) (MailboxSecurity, error)
	}); ok {
		return store.GetMailboxSecurity(ctx, email)
	}
	return MailboxSecurity{Email: email}, nil
}

func (s CompositeStore) GetAdminMFASettings(ctx context.Context) (domain.AdminMFASettings, error) {
	if store, ok := s.Auth.(interface {
		GetAdminMFASettings(context.Context) (domain.AdminMFASettings, error)
	}); ok {
		return store.GetAdminMFASettings(ctx)
	}
	return domain.AdminMFASettings{ID: 1, ProIdentityBaseURL: proIdentityAuthServiceURL, ProIdentityTimeoutSeconds: 90}, nil
}

func (s CompositeStore) CreateMailboxMFAChallenge(ctx context.Context, challenge MailboxMFAChallenge) error {
	if store, ok := s.Auth.(interface {
		CreateMailboxMFAChallenge(context.Context, MailboxMFAChallenge) error
	}); ok {
		return store.CreateMailboxMFAChallenge(ctx, challenge)
	}
	return sql.ErrNoRows
}

func (s CompositeStore) GetMailboxMFAChallenge(ctx context.Context, token string) (MailboxMFAChallenge, error) {
	if store, ok := s.Auth.(interface {
		GetMailboxMFAChallenge(context.Context, string) (MailboxMFAChallenge, error)
	}); ok {
		return store.GetMailboxMFAChallenge(ctx, token)
	}
	return MailboxMFAChallenge{}, sql.ErrNoRows
}

func (s CompositeStore) DeleteMailboxMFAChallenge(ctx context.Context, token string) error {
	if store, ok := s.Auth.(interface {
		DeleteMailboxMFAChallenge(context.Context, string) error
	}); ok {
		return store.DeleteMailboxMFAChallenge(ctx, token)
	}
	return nil
}

func (s CompositeStore) BeginMailboxTOTPEnrollment(ctx context.Context, email string) (MailboxTOTPEnrollment, error) {
	if store, ok := s.Auth.(interface {
		BeginMailboxTOTPEnrollment(context.Context, string) (MailboxTOTPEnrollment, error)
	}); ok {
		return store.BeginMailboxTOTPEnrollment(ctx, email)
	}
	return MailboxTOTPEnrollment{}, sql.ErrNoRows
}

func (s CompositeStore) VerifyMailboxTOTPEnrollment(ctx context.Context, email, code string) (MailboxSecurity, error) {
	if store, ok := s.Auth.(interface {
		VerifyMailboxTOTPEnrollment(context.Context, string, string) (MailboxSecurity, error)
	}); ok {
		return store.VerifyMailboxTOTPEnrollment(ctx, email, code)
	}
	return MailboxSecurity{}, sql.ErrNoRows
}

func (s CompositeStore) VerifyMailboxTOTPCode(ctx context.Context, email, code string) (bool, error) {
	if store, ok := s.Auth.(interface {
		VerifyMailboxTOTPCode(context.Context, string, string) (bool, error)
	}); ok {
		return store.VerifyMailboxTOTPCode(ctx, email, code)
	}
	return false, nil
}

func (s CompositeStore) ListAppPasswords(ctx context.Context, email string) ([]AppPassword, error) {
	if store, ok := s.Auth.(interface {
		ListAppPasswords(context.Context, string) ([]AppPassword, error)
	}); ok {
		return store.ListAppPasswords(ctx, email)
	}
	return nil, nil
}

func (s CompositeStore) CreateAppPassword(ctx context.Context, email string, req AppPassword) (AppPassword, error) {
	if store, ok := s.Auth.(interface {
		CreateAppPassword(context.Context, string, AppPassword) (AppPassword, error)
	}); ok {
		return store.CreateAppPassword(ctx, email, req)
	}
	return AppPassword{}, sql.ErrNoRows
}

func (s CompositeStore) RevokeAppPassword(ctx context.Context, email, id string) error {
	if store, ok := s.Auth.(interface {
		RevokeAppPassword(context.Context, string, string) error
	}); ok {
		return store.RevokeAppPassword(ctx, email, id)
	}
	return sql.ErrNoRows
}

func (s CompositeStore) RecordUserAudit(ctx context.Context, email, action, targetType, targetID string, metadata map[string]any) error {
	if store, ok := s.Auth.(interface {
		RecordUserAudit(context.Context, string, string, string, string, map[string]any) error
	}); ok {
		return store.RecordUserAudit(ctx, email, action, targetType, targetID, metadata)
	}
	return nil
}

func objectETag(body string) string {
	sum := sha256.Sum256([]byte(body))
	return `"` + hex.EncodeToString(sum[:]) + `"`
}

func normalizeProfileLanguage(value string) string {
	if normalized := i18n.NormalizeLanguage(value); normalized != "" {
		return normalized
	}
	return "en"
}

func normalizeMFAPurpose(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "setup":
		return "setup"
	default:
		return "login"
	}
}

func icalValue(body, name string) string {
	prefix := strings.ToUpper(name) + ":"
	for _, line := range strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n") {
		if strings.HasPrefix(strings.ToUpper(line), prefix) {
			return strings.TrimSpace(line[len(prefix):])
		}
	}
	return ""
}
