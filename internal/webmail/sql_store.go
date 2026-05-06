package webmail

import (
	"context"
	"database/sql"
	"log"

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
