package groupware

import (
	"context"
	"database/sql"

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
