package webmail

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

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

func (s CompositeStore) ListRecentMessages(ctx context.Context, email string, limit int) ([]MessageSummary, error) {
	return s.Mailbox.ListRecentMessages(ctx, email, limit)
}

func (s CompositeStore) ListMessages(ctx context.Context, email, folder string, limit int) ([]MessageSummary, error) {
	return s.Mailbox.ListMessages(ctx, email, folder, limit)
}

func (s CompositeStore) GetMessage(ctx context.Context, email, id string) (MessageDetail, error) {
	return s.Mailbox.GetMessage(ctx, email, id)
}

func (s CompositeStore) SendMessage(ctx context.Context, message OutboundMessage) error {
	return s.Sender.Send(ctx, message)
}

func (s CompositeStore) ReportMessage(ctx context.Context, email, id, verdict string) error {
	path, err := s.Mailbox.MessagePath(ctx, email, id)
	if err != nil {
		return err
	}
	if s.Learner != nil {
		if err := s.Learner.Learn(ctx, path, verdict); err != nil {
			log.Printf("rspamd learning failed verdict=%q: %v", verdict, err)
		}
	}
	if verdict == "spam" {
		if err := s.Mailbox.MoveMessage(ctx, email, id, "spam"); err != nil {
			log.Printf("spam folder move failed: %v", err)
		}
	} else if verdict == "ham" {
		if err := s.Mailbox.MoveMessage(ctx, email, id, "inbox"); err != nil {
			log.Printf("inbox folder move failed: %v", err)
		}
	}
	return s.Auth.ReportMessage(ctx, email, id, verdict)
}

func (s CompositeStore) MoveMessage(ctx context.Context, email, id, folder string) error {
	return s.Mailbox.MoveMessage(ctx, email, id, folder)
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
		return store.CreateFilter(ctx, email, filter)
	}
	return MailFilter{}, sql.ErrNoRows
}

func (s CompositeStore) UpdateFilter(ctx context.Context, email, id string, filter MailFilter) (MailFilter, error) {
	if store, ok := s.Auth.(interface {
		UpdateFilter(context.Context, string, string, MailFilter) (MailFilter, error)
	}); ok {
		return store.UpdateFilter(ctx, email, id, filter)
	}
	return MailFilter{}, sql.ErrNoRows
}

func (s CompositeStore) DeleteFilter(ctx context.Context, email, id string) error {
	if store, ok := s.Auth.(interface {
		DeleteFilter(context.Context, string, string) error
	}); ok {
		return store.DeleteFilter(ctx, email, id)
	}
	return sql.ErrNoRows
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

func (s CompositeStore) ChangePassword(ctx context.Context, email, newPassword string) error {
	if store, ok := s.Auth.(interface {
		ChangePassword(context.Context, string, string) error
	}); ok {
		return store.ChangePassword(ctx, email, newPassword)
	}
	return sql.ErrNoRows
}

func objectETag(body string) string {
	sum := sha256.Sum256([]byte(body))
	return `"` + hex.EncodeToString(sum[:]) + `"`
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
