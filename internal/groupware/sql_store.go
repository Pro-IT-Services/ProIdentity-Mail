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
