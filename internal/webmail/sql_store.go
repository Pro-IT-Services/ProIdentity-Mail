package webmail

import (
	"context"
	"database/sql"

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

type CompositeStore struct {
	Auth    SQLAuthStore
	Mailbox MaildirStore
}

func (s CompositeStore) VerifyUserPassword(ctx context.Context, email, password string) (bool, error) {
	return s.Auth.VerifyUserPassword(ctx, email, password)
}

func (s CompositeStore) ListRecentMessages(ctx context.Context, email string, limit int) ([]MessageSummary, error) {
	return s.Mailbox.ListRecentMessages(ctx, email, limit)
}

func (s CompositeStore) GetMessage(ctx context.Context, email, id string) (MessageDetail, error) {
	return s.Mailbox.GetMessage(ctx, email, id)
}
