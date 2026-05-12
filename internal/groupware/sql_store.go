package groupware

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"strings"

	"proidentity-mail/internal/security"
)

type SQLStore struct {
	db *sql.DB
}

func NewSQLStore(db *sql.DB) SQLStore {
	return SQLStore{db: db}
}

func (s SQLStore) VerifyUserPassword(ctx context.Context, email, password string) (bool, error) {
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

func (s SQLStore) VerifyProtocolPassword(ctx context.Context, email, password, protocol string) (bool, error) {
	protocol = normalizeProtocol(protocol)
	if protocol == "" {
		return false, nil
	}
	email = strings.ToLower(strings.TrimSpace(email))
	var userID uint64
	var hash string
	var forceMFA bool
	var totpEnabled bool
	err := s.db.QueryRowContext(ctx, `
		SELECT u.id, u.password_hash, COALESCE(ms.force_mailbox_mfa, 0), COALESCE(mfa.totp_enabled, 0)
		FROM users u
		JOIN domains d ON d.id = u.primary_domain_id
		LEFT JOIN mail_server_settings ms ON ms.id = 1
		LEFT JOIN user_mfa_settings mfa ON mfa.user_id = u.id
		WHERE CONCAT(u.local_part, '@', d.name) = ?
		  AND u.mailbox_type = 'user'
		  AND u.status = 'active'
		  AND d.status IN ('pending', 'active')
		LIMIT 1`, email).Scan(&userID, &hash, &forceMFA, &totpEnabled)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !forceMFA && !totpEnabled && security.VerifyPassword(hash, password) {
		return true, nil
	}
	fingerprint := sha256.Sum256([]byte(password))
	var id uint64
	var protocols string
	err = s.db.QueryRowContext(ctx, `
		SELECT id, protocols
		FROM user_app_passwords
		WHERE user_id = ? AND secret_sha256 = ? AND status = 'active'
		LIMIT 1`, userID, hex.EncodeToString(fingerprint[:])).Scan(&id, &protocols)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !protocolListAllows(protocols, protocol) {
		return false, nil
	}
	_, _ = s.db.ExecContext(ctx, `UPDATE user_app_passwords SET last_used_at = UTC_TIMESTAMP(), last_used_protocol = ? WHERE id = ?`, protocol, id)
	return true, nil
}

func (s SQLStore) PutContact(ctx context.Context, email, href string, body []byte) (DAVObject, error) {
	bookID, err := s.ensureAddressBook(ctx, email)
	if err != nil {
		return DAVObject{}, err
	}
	object := DAVObject{Href: href, ETag: etag(body), Body: append([]byte(nil), body...)}
	uid := valueFromLine(string(body), "UID")
	if uid == "" {
		uid = strings.TrimSuffix(href, ".vcf")
	}
	fullName := valueFromLine(string(body), "FN")
	mail := valueFromLine(string(body), "EMAIL")
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO contact_objects(address_book_id, uid, href, etag, vcard, full_name, email)
		VALUES (?, ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''))
		ON DUPLICATE KEY UPDATE
		  uid = VALUES(uid),
		  etag = VALUES(etag),
		  vcard = VALUES(vcard),
		  full_name = VALUES(full_name),
		  email = VALUES(email)`, bookID, uid, href, object.ETag, string(body), fullName, mail)
	if err != nil {
		return DAVObject{}, err
	}
	return object, nil
}

func (s SQLStore) GetContact(ctx context.Context, email, href string) (DAVObject, error) {
	bookID, err := s.ensureAddressBook(ctx, email)
	if err != nil {
		return DAVObject{}, err
	}
	var object DAVObject
	var body string
	err = s.db.QueryRowContext(ctx, `SELECT href, etag, vcard FROM contact_objects WHERE address_book_id = ? AND href = ?`, bookID, href).Scan(&object.Href, &object.ETag, &body)
	if err == sql.ErrNoRows {
		return DAVObject{}, ErrNotFound
	}
	if err != nil {
		return DAVObject{}, err
	}
	object.Body = []byte(body)
	return object, nil
}

func (s SQLStore) ListContacts(ctx context.Context, email string) ([]DAVObject, error) {
	bookID, err := s.ensureAddressBook(ctx, email)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT href, etag, vcard FROM contact_objects WHERE address_book_id = ? ORDER BY href`, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var objects []DAVObject
	for rows.Next() {
		var object DAVObject
		var body string
		if err := rows.Scan(&object.Href, &object.ETag, &body); err != nil {
			return nil, err
		}
		object.Body = []byte(body)
		objects = append(objects, object)
	}
	return objects, rows.Err()
}

func (s SQLStore) PutCalendarObject(ctx context.Context, email, href string, body []byte) (DAVObject, error) {
	calendarID, err := s.ensureCalendar(ctx, email)
	if err != nil {
		return DAVObject{}, err
	}
	object := DAVObject{Href: href, ETag: etag(body), Body: append([]byte(nil), body...)}
	uid := valueFromLine(string(body), "UID")
	if uid == "" {
		uid = strings.TrimSuffix(href, ".ics")
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO calendar_objects(calendar_id, uid, href, etag, icalendar)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		  uid = VALUES(uid),
		  etag = VALUES(etag),
		  icalendar = VALUES(icalendar)`, calendarID, uid, href, object.ETag, string(body))
	if err != nil {
		return DAVObject{}, err
	}
	return object, nil
}

func (s SQLStore) GetCalendarObject(ctx context.Context, email, href string) (DAVObject, error) {
	calendarID, err := s.ensureCalendar(ctx, email)
	if err != nil {
		return DAVObject{}, err
	}
	var object DAVObject
	var body string
	err = s.db.QueryRowContext(ctx, `SELECT href, etag, icalendar FROM calendar_objects WHERE calendar_id = ? AND href = ?`, calendarID, href).Scan(&object.Href, &object.ETag, &body)
	if err == sql.ErrNoRows {
		return DAVObject{}, ErrNotFound
	}
	if err != nil {
		return DAVObject{}, err
	}
	object.Body = []byte(body)
	return object, nil
}

func (s SQLStore) ListCalendarObjects(ctx context.Context, email string) ([]DAVObject, error) {
	calendarID, err := s.ensureCalendar(ctx, email)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT href, etag, icalendar FROM calendar_objects WHERE calendar_id = ? ORDER BY href`, calendarID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var objects []DAVObject
	for rows.Next() {
		var object DAVObject
		var body string
		if err := rows.Scan(&object.Href, &object.ETag, &body); err != nil {
			return nil, err
		}
		object.Body = []byte(body)
		objects = append(objects, object)
	}
	return objects, rows.Err()
}

func (s SQLStore) ensureAddressBook(ctx context.Context, email string) (uint64, error) {
	tenantID, userID, err := s.userIDs(ctx, email)
	if err != nil {
		return 0, err
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT IGNORE INTO address_books(tenant_id, user_id, slug, display_name)
		VALUES (?, ?, 'default', 'Default Address Book')`, tenantID, userID); err != nil {
		return 0, err
	}
	var id uint64
	err = s.db.QueryRowContext(ctx, `SELECT id FROM address_books WHERE user_id = ? AND slug = 'default'`, userID).Scan(&id)
	return id, err
}

func (s SQLStore) ensureCalendar(ctx context.Context, email string) (uint64, error) {
	tenantID, userID, err := s.userIDs(ctx, email)
	if err != nil {
		return 0, err
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT IGNORE INTO calendars(tenant_id, user_id, slug, display_name)
		VALUES (?, ?, 'default', 'Default Calendar')`, tenantID, userID); err != nil {
		return 0, err
	}
	var id uint64
	err = s.db.QueryRowContext(ctx, `SELECT id FROM calendars WHERE user_id = ? AND slug = 'default'`, userID).Scan(&id)
	return id, err
}

func (s SQLStore) userIDs(ctx context.Context, email string) (uint64, uint64, error) {
	var tenantID, userID uint64
	err := s.db.QueryRowContext(ctx, `
		SELECT u.tenant_id, u.id
		FROM users u
		JOIN domains d ON d.id = u.primary_domain_id
		WHERE CONCAT(u.local_part, '@', d.name) = ?
		  AND u.status = 'active'
		  AND d.status IN ('pending', 'active')
		LIMIT 1`, email).Scan(&tenantID, &userID)
	if err == sql.ErrNoRows {
		return 0, 0, ErrNotFound
	}
	return tenantID, userID, err
}

func etag(body []byte) string {
	sum := sha256.Sum256(body)
	return `"` + hex.EncodeToString(sum[:]) + `"`
}

func valueFromLine(text, name string) string {
	prefix := strings.ToUpper(name) + ":"
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if strings.HasPrefix(strings.ToUpper(line), prefix) {
			return strings.TrimSpace(line[len(prefix):])
		}
	}
	return ""
}

func normalizeProtocol(protocol string) string {
	protocol = strings.ToLower(strings.TrimSpace(protocol))
	switch protocol {
	case "carddav", "caldav", "groupware":
		return "dav"
	default:
		return protocol
	}
}

func protocolListAllows(protocols, protocol string) bool {
	protocol = normalizeProtocol(protocol)
	for _, candidate := range strings.Split(protocols, ",") {
		if normalizeProtocol(candidate) == protocol {
			return true
		}
	}
	return false
}
