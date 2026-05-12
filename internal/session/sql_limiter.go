package session

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"strings"
	"time"
)

const (
	authSprayAlertAction              = "security.alert.auth_spray"
	authSprayDistinctAccountThreshold = 5
	authSprayWindow                   = time.Minute
	authSprayAlertCooldown            = 5 * time.Minute
	authSprayMaxSampleAccounts        = 5
)

type SQLLoginLimiter struct {
	db                     *sql.DB
	service                string
	penalties              []Penalty
	window                 time.Duration
	accountLockoutFailures int
}

func NewSQLLoginLimiter(db *sql.DB, service string, options Options) *SQLLoginLimiter {
	window := options.Window
	if window == 0 {
		window = time.Hour
	}
	accountLockoutFailures := options.AccountLockoutFailures
	if accountLockoutFailures == 0 {
		accountLockoutFailures = 10
	}
	return &SQLLoginLimiter{
		db:                     db,
		service:                strings.TrimSpace(service),
		penalties:              normalizePenaltySchedule(options),
		window:                 window,
		accountLockoutFailures: accountLockoutFailures,
	}
}

func (l *SQLLoginLimiter) Locked(key string) bool {
	if l == nil || l.db == nil || strings.TrimSpace(key) == "" {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var lockedUntil sql.NullTime
	err := l.db.QueryRowContext(ctx, `
		SELECT locked_until
		FROM login_rate_limits
		WHERE service = ? AND limiter_key = ?`, l.service, key).Scan(&lockedUntil)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		log.Printf("login limiter locked check failed service=%q key=%q: %v", l.service, key, err)
		return true
	}
	if !lockedUntil.Valid {
		return false
	}
	now := time.Now().UTC()
	if now.Before(lockedUntil.Time.UTC()) {
		return true
	}
	if _, err := l.db.ExecContext(ctx, `UPDATE login_rate_limits SET locked_until = NULL, updated_at = UTC_TIMESTAMP() WHERE service = ? AND limiter_key = ?`, l.service, key); err != nil {
		log.Printf("login limiter expired lock reset failed service=%q key=%q: %v", l.service, key, err)
	}
	return false
}

func (l *SQLLoginLimiter) Fail(key string) {
	if l == nil || l.db == nil || strings.TrimSpace(key) == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("login limiter transaction failed service=%q key=%q: %v", l.service, key, err)
		return
	}
	defer tx.Rollback()

	var currentCount int
	var firstFailedAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
		SELECT failure_count, first_failed_at
		FROM login_rate_limits
		WHERE service = ? AND limiter_key = ?
		FOR UPDATE`, l.service, key).Scan(&currentCount, &firstFailedAt)
	rowExists := true
	if err == sql.ErrNoRows {
		rowExists = false
	} else if err != nil {
		log.Printf("login limiter lookup failed service=%q key=%q: %v", l.service, key, err)
		return
	}

	now := time.Now().UTC()
	first := now
	if rowExists && firstFailedAt.Valid {
		first = firstFailedAt.Time.UTC()
	}
	if !rowExists || !firstFailedAt.Valid || now.Sub(first) > l.window {
		currentCount = 0
		first = now
	}
	currentCount++
	var lockedUntil any
	if lockout := LockoutForFailureCount(currentCount, l.penalties); lockout > 0 {
		lockedUntil = now.Add(lockout)
	}

	if rowExists {
		_, err = tx.ExecContext(ctx, `
			UPDATE login_rate_limits
			SET failure_count = ?,
			    first_failed_at = ?,
			    last_failed_at = ?,
			    locked_until = ?,
			    updated_at = UTC_TIMESTAMP()
			WHERE service = ? AND limiter_key = ?`, currentCount, first, now, lockedUntil, l.service, key)
	} else {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO login_rate_limits(service, limiter_key, failure_count, first_failed_at, last_failed_at, locked_until, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, UTC_TIMESTAMP())`, l.service, key, currentCount, first, now, lockedUntil)
	}
	if err != nil {
		log.Printf("login limiter fail record failed service=%q key=%q: %v", l.service, key, err)
		return
	}
	if currentCount >= l.accountLockoutFailures {
		if email := AccountEmailFromLimiterKey(key); email != "" {
			if _, err := tx.ExecContext(ctx, `
				UPDATE users u
				JOIN domains d ON d.id = u.primary_domain_id
				SET u.status = 'locked'
				WHERE CONCAT(u.local_part, '@', d.name) = ?
				  AND u.mailbox_type = 'user'
				  AND u.status = 'active'`, email); err != nil {
				log.Printf("login limiter account lock failed service=%q email=%q: %v", l.service, email, err)
				return
			}
		}
	}
	l.recordAuthSprayAlert(ctx, tx, key, now)
	if err := tx.Commit(); err != nil {
		log.Printf("login limiter commit failed service=%q key=%q: %v", l.service, key, err)
	}
}

type pairLimiterKey struct {
	Scope   string
	Account string
	Client  string
}

func (l *SQLLoginLimiter) recordAuthSprayAlert(ctx context.Context, tx *sql.Tx, key string, now time.Time) {
	pair, ok := parsePairLimiterKey(key)
	if !ok || pair.Client == "unknown" {
		return
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT limiter_key
		FROM login_rate_limits
		WHERE service = ?
		  AND last_failed_at >= ?
		  AND limiter_key LIKE ?`, l.service, now.Add(-authSprayWindow), "%|pair|%")
	if err != nil {
		log.Printf("login limiter auth spray scan failed service=%q client=%q: %v", l.service, pair.Client, err)
		return
	}
	defer rows.Close()

	accounts := make(map[string]struct{})
	for rows.Next() {
		var rowKey string
		if err := rows.Scan(&rowKey); err != nil {
			log.Printf("login limiter auth spray row scan failed service=%q client=%q: %v", l.service, pair.Client, err)
			return
		}
		candidate, ok := parsePairLimiterKey(rowKey)
		if !ok || candidate.Client != pair.Client {
			continue
		}
		accounts[candidate.Account] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		log.Printf("login limiter auth spray rows failed service=%q client=%q: %v", l.service, pair.Client, err)
		return
	}
	if len(accounts) < authSprayDistinctAccountThreshold {
		return
	}

	var existingID uint64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM audit_events
		WHERE action = ?
		  AND target_type = 'client_ip'
		  AND target_id = ?
		  AND created_at >= ?
		  AND metadata_json LIKE ?
		LIMIT 1`, authSprayAlertAction, pair.Client, now.Add(-authSprayAlertCooldown), `%"service":"`+l.service+`"%`).Scan(&existingID)
	if err == nil {
		return
	}
	if err != sql.ErrNoRows {
		log.Printf("login limiter auth spray cooldown check failed service=%q client=%q: %v", l.service, pair.Client, err)
		return
	}

	samples := make([]string, 0, authSprayMaxSampleAccounts)
	for account := range accounts {
		if len(samples) >= authSprayMaxSampleAccounts {
			break
		}
		samples = append(samples, account)
	}
	metadata, err := json.Marshal(map[string]any{
		"service":           l.service,
		"client_ip":         pair.Client,
		"distinct_accounts": len(accounts),
		"window_seconds":    int(authSprayWindow.Seconds()),
		"sample_accounts":   samples,
	})
	if err != nil {
		log.Printf("login limiter auth spray metadata failed service=%q client=%q: %v", l.service, pair.Client, err)
		return
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO audit_events(actor_type, action, target_type, target_id, metadata_json)
		VALUES ('system', ?, 'client_ip', ?, ?)`, authSprayAlertAction, pair.Client, string(metadata)); err != nil {
		log.Printf("login limiter auth spray alert insert failed service=%q client=%q: %v", l.service, pair.Client, err)
	}
}

func (l *SQLLoginLimiter) Success(key string) {
	if l == nil || l.db == nil || strings.TrimSpace(key) == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := l.db.ExecContext(ctx, `DELETE FROM login_rate_limits WHERE service = ? AND limiter_key = ?`, l.service, key); err != nil {
		log.Printf("login limiter reset failed service=%q key=%q: %v", l.service, key, err)
	}
}

func AccountEmailFromLimiterKey(key string) string {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(key)), "|")
	if len(parts) != 3 || parts[1] != "account" {
		return ""
	}
	email := strings.TrimSpace(parts[2])
	if !strings.Contains(email, "@") || strings.ContainsAny(email, " \t\r\n\x00") {
		return ""
	}
	local, domainName, ok := strings.Cut(email, "@")
	if !ok || local == "" || domainName == "" || strings.Contains(domainName, "@") || !strings.Contains(domainName, ".") {
		return ""
	}
	return email
}

func parsePairLimiterKey(key string) (pairLimiterKey, bool) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(key)), "|")
	if len(parts) != 4 || parts[1] != "pair" {
		return pairLimiterKey{}, false
	}
	pair := pairLimiterKey{
		Scope:   strings.TrimSpace(parts[0]),
		Account: strings.TrimSpace(parts[2]),
		Client:  strings.TrimSpace(parts[3]),
	}
	if !safeLimiterPart(pair.Scope) || !safeLimiterPart(pair.Account) || !safeLimiterPart(pair.Client) {
		return pairLimiterKey{}, false
	}
	return pair, true
}

func safeLimiterPart(value string) bool {
	if value == "" || len(value) > 320 || strings.ContainsAny(value, "\x00\r\n\t") {
		return false
	}
	for _, char := range value {
		if char < 0x20 || char == 0x7f {
			return false
		}
	}
	return true
}
