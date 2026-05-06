# Mail Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first working single-server foundation for the production mail platform: Go control-plane scaffold, MariaDB schema, admin API, config rendering, deployment scripts, and service health checks.

**Architecture:** The first slice creates a Go monorepo that controls platform state and renders configuration for Postfix, Dovecot, Rspamd, ClamAV, MariaDB, and Redis. The Go services do not implement SMTP/IMAP/POP3; they manage tenants, domains, users, policies, audit logs, and safe config generation for proven mail daemons.

**Tech Stack:** Go 1.22+, MariaDB, Redis, Postfix, Dovecot, Rspamd, ClamAV, systemd, Debian 13, `chi` router, `go-sql-driver/mysql`, `golang-migrate`-style SQL files, `bcrypt`.

---

## Scope

This plan implements the foundation only. Webmail, calendar, contacts, CardDAV, CalDAV, JMAP, and ActiveSync are separate follow-up plans after the mail foundation is running.

## File Structure

Create these files and directories:

- `go.mod`: Go module declaration and dependencies.
- `cmd/mailctl/main.go`: CLI entrypoint for migrations, seed data, config rendering, and health checks.
- `cmd/webadmin/main.go`: webadmin/API service entrypoint.
- `internal/app/config.go`: environment-driven runtime configuration.
- `internal/app/server.go`: HTTP server wiring.
- `internal/db/db.go`: MariaDB connection helper.
- `internal/db/migrate.go`: embedded migration runner.
- `internal/db/migrations/0001_foundation.sql`: initial schema.
- `internal/domain/models.go`: tenant, domain, user, alias, policy, and audit structs.
- `internal/security/password.go`: password hashing and verification.
- `internal/security/password_test.go`: password hashing tests.
- `internal/admin/handlers.go`: admin HTTP handlers.
- `internal/admin/handlers_test.go`: admin handler tests.
- `internal/render/render.go`: service config renderer.
- `internal/render/templates.go`: embedded Postfix, Dovecot, and Rspamd templates.
- `internal/render/render_test.go`: rendering tests.
- `internal/health/checks.go`: local service health checks.
- `deploy/devmail/install-packages.sh`: package installation script for Debian 13.
- `deploy/devmail/proidentity-mail.env.example`: environment template.
- `deploy/devmail/proidentity-webadmin.service`: systemd unit for webadmin.
- `deploy/devmail/proidentity-mailctl.service`: systemd unit for one-shot config rendering.
- `deploy/devmail/README.md`: deployment instructions for `root@192.168.254.125`.
- `README.md`: project overview and first-run commands.

## Task 1: Initialize Go Module And Repo Docs

**Files:**
- Create: `go.mod`
- Create: `README.md`

- [ ] **Step 1: Create module file**

Create `go.mod`:

```go
module proidentity-mail

go 1.22

require (
	github.com/go-chi/chi/v5 v5.0.12
	github.com/go-sql-driver/mysql v1.8.1
	golang.org/x/crypto v0.22.0
)
```

- [ ] **Step 2: Create project README**

Create `README.md`:

```markdown
# ProIdentity Mail

Production-first multi-tenant mail and groupware platform.

The first foundation uses Go for the control plane and proven mail daemons for protocol-heavy services:

- Postfix for SMTP and submission
- Dovecot for IMAP, POP3, LMTP, Sieve, and ManageSieve
- Rspamd for spam filtering and mail authentication checks
- ClamAV for malware scanning
- MariaDB for platform state
- Redis for cache, sessions, and Rspamd state

## First Local Commands

```powershell
go test ./...
go run ./cmd/mailctl -help
go run ./cmd/webadmin
```

## Development Target

Initial deployment target:

- Host: DevMail
- SSH: root@192.168.254.125
- OS: Debian GNU/Linux 13 trixie
```

- [ ] **Step 3: Run module verification**

Run:

```powershell
go test ./...
```

Expected: command succeeds after later tasks add packages. At this point it may report no packages.

- [ ] **Step 4: Commit**

Run:

```powershell
git add go.mod README.md
git commit -m "chore: initialize mail platform module"
```

## Task 2: Add Runtime Configuration

**Files:**
- Create: `internal/app/config.go`
- Test: `internal/app/config_test.go`

- [ ] **Step 1: Write failing config tests**

Create `internal/app/config_test.go`:

```go
package app

import "testing"

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("PROIDENTITY_DB_DSN", "")
	t.Setenv("PROIDENTITY_HTTP_ADDR", "")
	t.Setenv("PROIDENTITY_CONFIG_DIR", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.HTTPAddr != "127.0.0.1:8080" {
		t.Fatalf("HTTPAddr = %q, want default", cfg.HTTPAddr)
	}
	if cfg.ConfigDir != "/etc/proidentity-mail/generated" {
		t.Fatalf("ConfigDir = %q, want default", cfg.ConfigDir)
	}
}

func TestLoadConfigRequiresDSNForDatabaseUse(t *testing.T) {
	t.Setenv("PROIDENTITY_DB_DSN", "mail:secret@tcp(127.0.0.1:3306)/proidentity_mail?parseTime=true")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.DBDSN == "" {
		t.Fatal("DBDSN is empty")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
go test ./internal/app
```

Expected: FAIL because `LoadConfig` is not defined.

- [ ] **Step 3: Implement config loader**

Create `internal/app/config.go`:

```go
package app

import "os"

type Config struct {
	HTTPAddr  string
	DBDSN     string
	ConfigDir string
}

func LoadConfig() (Config, error) {
	cfg := Config{
		HTTPAddr:  valueOrDefault(os.Getenv("PROIDENTITY_HTTP_ADDR"), "127.0.0.1:8080"),
		DBDSN:     os.Getenv("PROIDENTITY_DB_DSN"),
		ConfigDir: valueOrDefault(os.Getenv("PROIDENTITY_CONFIG_DIR"), "/etc/proidentity-mail/generated"),
	}
	return cfg, nil
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```powershell
go test ./internal/app
```

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```powershell
git add internal/app/config.go internal/app/config_test.go
git commit -m "feat: add runtime configuration"
```

## Task 3: Add MariaDB Connection And Foundation Schema

**Files:**
- Create: `internal/db/db.go`
- Create: `internal/db/migrate.go`
- Create: `internal/db/migrations/0001_foundation.sql`

- [ ] **Step 1: Create DB connection helper**

Create `internal/db/db.go`:

```go
package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(20)
	conn.SetMaxIdleConns(10)
	conn.SetConnMaxLifetime(10 * time.Minute)
	if err := conn.PingContext(ctx); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}
```

- [ ] **Step 2: Create migration runner**

Create `internal/db/migrate.go`:

```go
package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Migrate(ctx context.Context, conn *sql.DB) error {
	if _, err := conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version varchar(255) PRIMARY KEY,
		applied_at timestamp NOT NULL DEFAULT current_timestamp()
	)`); err != nil {
		return err
	}

	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		applied, err := migrationApplied(ctx, conn, name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		sqlText, err := migrationFiles.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		if _, err := conn.ExecContext(ctx, string(sqlText)); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}
		if _, err := conn.ExecContext(ctx, `INSERT INTO schema_migrations(version) VALUES (?)`, name); err != nil {
			return err
		}
	}
	return nil
}

func migrationApplied(ctx context.Context, conn *sql.DB, version string) (bool, error) {
	var count int
	err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version).Scan(&count)
	return count > 0, err
}
```

- [ ] **Step 3: Create foundation schema**

Create `internal/db/migrations/0001_foundation.sql`:

```sql
CREATE TABLE tenants (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name varchar(190) NOT NULL,
  slug varchar(120) NOT NULL,
  status enum('active','suspended') NOT NULL DEFAULT 'active',
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  UNIQUE KEY tenants_slug_unique (slug)
);

CREATE TABLE domains (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  tenant_id bigint unsigned NOT NULL,
  name varchar(253) NOT NULL,
  status enum('pending','active','disabled') NOT NULL DEFAULT 'pending',
  dkim_selector varchar(63) NOT NULL DEFAULT 'mail',
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  UNIQUE KEY domains_name_unique (name),
  KEY domains_tenant_id_idx (tenant_id),
  CONSTRAINT domains_tenant_id_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

CREATE TABLE users (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  tenant_id bigint unsigned NOT NULL,
  primary_domain_id bigint unsigned NOT NULL,
  local_part varchar(128) NOT NULL,
  display_name varchar(190) NOT NULL,
  password_hash varchar(255) NOT NULL,
  status enum('active','locked','disabled') NOT NULL DEFAULT 'active',
  quota_bytes bigint unsigned NOT NULL DEFAULT 10737418240,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  UNIQUE KEY users_mailbox_unique (primary_domain_id, local_part),
  KEY users_tenant_id_idx (tenant_id),
  CONSTRAINT users_tenant_id_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id),
  CONSTRAINT users_primary_domain_id_fk FOREIGN KEY (primary_domain_id) REFERENCES domains(id)
);

CREATE TABLE aliases (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  tenant_id bigint unsigned NOT NULL,
  domain_id bigint unsigned NOT NULL,
  source_local_part varchar(128) NOT NULL,
  destination varchar(320) NOT NULL,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  UNIQUE KEY aliases_source_destination_unique (domain_id, source_local_part, destination),
  KEY aliases_tenant_id_idx (tenant_id),
  CONSTRAINT aliases_tenant_id_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id),
  CONSTRAINT aliases_domain_id_fk FOREIGN KEY (domain_id) REFERENCES domains(id)
);

CREATE TABLE tenant_policies (
  tenant_id bigint unsigned NOT NULL PRIMARY KEY,
  spam_action enum('mark','quarantine','reject') NOT NULL DEFAULT 'mark',
  malware_action enum('quarantine','reject') NOT NULL DEFAULT 'quarantine',
  require_tls_for_auth tinyint(1) NOT NULL DEFAULT 1,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  CONSTRAINT tenant_policies_tenant_id_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

CREATE TABLE audit_events (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  tenant_id bigint unsigned NULL,
  actor_type varchar(40) NOT NULL,
  actor_id bigint unsigned NULL,
  action varchar(120) NOT NULL,
  target_type varchar(80) NOT NULL,
  target_id varchar(120) NOT NULL,
  metadata_json json NOT NULL,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  KEY audit_events_tenant_id_created_idx (tenant_id, created_at),
  KEY audit_events_action_created_idx (action, created_at)
);
```

- [ ] **Step 4: Run tests/build**

Run:

```powershell
go test ./...
```

Expected: PASS or no test files for `internal/db`.

- [ ] **Step 5: Commit**

Run:

```powershell
git add internal/db
git commit -m "feat: add foundation database schema"
```

## Task 4: Add Domain Models And Password Security

**Files:**
- Create: `internal/domain/models.go`
- Create: `internal/security/password.go`
- Create: `internal/security/password_test.go`

- [ ] **Step 1: Create domain structs**

Create `internal/domain/models.go`:

```go
package domain

import "time"

type Tenant struct {
	ID        uint64
	Name      string
	Slug      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Domain struct {
	ID           uint64
	TenantID     uint64
	Name         string
	Status       string
	DKIMSelector string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type User struct {
	ID              uint64
	TenantID        uint64
	PrimaryDomainID uint64
	LocalPart       string
	DisplayName     string
	PasswordHash    string
	Status          string
	QuotaBytes      uint64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
```

- [ ] **Step 2: Write failing password tests**

Create `internal/security/password_test.go`:

```go
package security

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("hash is empty")
	}
	if hash == "correct horse battery staple" {
		t.Fatal("hash contains plaintext password")
	}
	if !VerifyPassword(hash, "correct horse battery staple") {
		t.Fatal("expected password to verify")
	}
	if VerifyPassword(hash, "wrong password") {
		t.Fatal("wrong password verified")
	}
}

func TestHashPasswordRejectsEmptyPassword(t *testing.T) {
	if _, err := HashPassword(""); err == nil {
		t.Fatal("expected empty password to fail")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run:

```powershell
go test ./internal/security
```

Expected: FAIL because `HashPassword` and `VerifyPassword` are not defined.

- [ ] **Step 4: Implement password hashing**

Create `internal/security/password.go`:

```go
package security

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password must not be empty")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func VerifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
```

- [ ] **Step 5: Run test to verify it passes**

Run:

```powershell
go test ./internal/security
```

Expected: PASS.

- [ ] **Step 6: Commit**

Run:

```powershell
git add internal/domain internal/security
git commit -m "feat: add domain models and password hashing"
```

## Task 5: Add Admin HTTP API

**Files:**
- Create: `internal/admin/handlers.go`
- Create: `internal/admin/handlers_test.go`
- Create: `internal/app/server.go`
- Create: `cmd/webadmin/main.go`

- [ ] **Step 1: Write failing handler tests**

Create `internal/admin/handlers_test.go`:

```go
package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	handler := NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
go test ./internal/admin
```

Expected: FAIL because `NewRouter` is not defined.

- [ ] **Step 3: Implement admin router**

Create `internal/admin/handlers.go`:

```go
package admin

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/healthz", health)
	return r
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

- [ ] **Step 4: Add app server wiring**

Create `internal/app/server.go`:

```go
package app

import (
	"net/http"
	"time"

	"proidentity-mail/internal/admin"
)

func NewHTTPServer(cfg Config) *http.Server {
	return &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           admin.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}
}
```

- [ ] **Step 5: Add webadmin entrypoint**

Create `cmd/webadmin/main.go`:

```go
package main

import (
	"log"

	"proidentity-mail/internal/app"
)

func main() {
	cfg, err := app.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	server := app.NewHTTPServer(cfg)
	log.Printf("webadmin listening on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("webadmin stopped: %v", err)
	}
}
```

- [ ] **Step 6: Run tests**

Run:

```powershell
go test ./...
```

Expected: PASS.

- [ ] **Step 7: Commit**

Run:

```powershell
git add internal/admin internal/app/server.go cmd/webadmin
git commit -m "feat: add webadmin health endpoint"
```

## Task 6: Add Config Renderer

**Files:**
- Create: `internal/render/templates.go`
- Create: `internal/render/render.go`
- Create: `internal/render/render_test.go`

- [ ] **Step 1: Write failing renderer test**

Create `internal/render/render_test.go`:

```go
package render

import (
	"strings"
	"testing"
)

func TestRenderPostfixMainIncludesVirtualMailboxDomain(t *testing.T) {
	out, err := RenderPostfixMain(PostfixMainData{
		Hostname: "mail.example.com",
	})
	if err != nil {
		t.Fatalf("RenderPostfixMain returned error: %v", err)
	}
	text := string(out)
	if !strings.Contains(text, "myhostname = mail.example.com") {
		t.Fatalf("rendered config missing hostname: %s", text)
	}
	if !strings.Contains(text, "smtpd_milters = inet:127.0.0.1:11332") {
		t.Fatalf("rendered config missing rspamd milter: %s", text)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
go test ./internal/render
```

Expected: FAIL because renderer functions are not defined.

- [ ] **Step 3: Implement templates**

Create `internal/render/templates.go`:

```go
package render

const postfixMainTemplate = `
myhostname = {{ .Hostname }}
myorigin = $myhostname
inet_interfaces = all
inet_protocols = ipv4
smtpd_tls_security_level = may
smtp_tls_security_level = may
smtpd_sasl_auth_enable = yes
smtpd_sasl_type = dovecot
smtpd_sasl_path = private/auth
virtual_transport = lmtp:unix:private/dovecot-lmtp
smtpd_milters = inet:127.0.0.1:11332
non_smtpd_milters = inet:127.0.0.1:11332
milter_protocol = 6
milter_default_action = tempfail
`

const dovecotSQLTemplate = `
driver = mysql
connect = host=127.0.0.1 dbname={{ .Database }} user={{ .User }} password={{ .Password }}
password_query = SELECT local_part AS user, password_hash AS password FROM users WHERE local_part = '%n' AND status = 'active'
`

const rspamdLocalTemplate = `
redis {
  servers = "127.0.0.1";
}
`
```

- [ ] **Step 4: Implement renderer**

Create `internal/render/render.go`:

```go
package render

import (
	"bytes"
	"text/template"
)

type PostfixMainData struct {
	Hostname string
}

type DovecotSQLData struct {
	Database string
	User     string
	Password string
}

func RenderPostfixMain(data PostfixMainData) ([]byte, error) {
	return renderTemplate("postfix-main", postfixMainTemplate, data)
}

func RenderDovecotSQL(data DovecotSQLData) ([]byte, error) {
	return renderTemplate("dovecot-sql", dovecotSQLTemplate, data)
}

func RenderRspamdLocal() ([]byte, error) {
	return []byte(rspamdLocalTemplate), nil
}

func renderTemplate(name, text string, data any) ([]byte, error) {
	tmpl, err := template.New(name).Parse(text)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return bytes.TrimLeft(buf.Bytes(), "\n"), nil
}
```

- [ ] **Step 5: Run tests**

Run:

```powershell
go test ./internal/render
```

Expected: PASS.

- [ ] **Step 6: Commit**

Run:

```powershell
git add internal/render
git commit -m "feat: add mail service config renderer"
```

## Task 7: Add mailctl CLI For Migrate, Render, And Health

**Files:**
- Create: `internal/health/checks.go`
- Create: `cmd/mailctl/main.go`

- [ ] **Step 1: Create health checks**

Create `internal/health/checks.go`:

```go
package health

import (
	"context"
	"net"
	"time"
)

type CheckResult struct {
	Name string
	OK   bool
	Err  string
}

func TCP(ctx context.Context, name, address string) CheckResult {
	dialer := net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return CheckResult{Name: name, OK: false, Err: err.Error()}
	}
	_ = conn.Close()
	return CheckResult{Name: name, OK: true}
}
```

- [ ] **Step 2: Create CLI**

Create `cmd/mailctl/main.go`:

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"proidentity-mail/internal/app"
	"proidentity-mail/internal/db"
	"proidentity-mail/internal/health"
	"proidentity-mail/internal/render"
)

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: mailctl migrate|render|health")
		os.Exit(2)
	}

	cfg, err := app.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	switch flag.Arg(0) {
	case "migrate":
		runMigrate(cfg)
	case "render":
		runRender(cfg)
	case "health":
		runHealth()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", flag.Arg(0))
		os.Exit(2)
	}
}

func runMigrate(cfg app.Config) {
	if cfg.DBDSN == "" {
		log.Fatal("PROIDENTITY_DB_DSN is required")
	}
	ctx := context.Background()
	conn, err := db.Open(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer conn.Close()
	if err := db.Migrate(ctx, conn); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	fmt.Println("migrations applied")
}

func runRender(cfg app.Config) {
	if err := os.MkdirAll(cfg.ConfigDir, 0750); err != nil {
		log.Fatalf("create config dir: %v", err)
	}
	writeRendered(filepath.Join(cfg.ConfigDir, "postfix-main.cf"), must(render.RenderPostfixMain(render.PostfixMainData{Hostname: "mail.local"})))
	writeRendered(filepath.Join(cfg.ConfigDir, "dovecot-sql.conf.ext"), must(render.RenderDovecotSQL(render.DovecotSQLData{Database: "proidentity_mail", User: "proidentity_mail", Password: "change-me"})))
	writeRendered(filepath.Join(cfg.ConfigDir, "rspamd-local.d-redis.conf"), must(render.RenderRspamdLocal()))
	fmt.Printf("rendered configs to %s\n", cfg.ConfigDir)
}

func runHealth() {
	ctx := context.Background()
	results := []health.CheckResult{
		health.TCP(ctx, "smtp", "127.0.0.1:25"),
		health.TCP(ctx, "submission", "127.0.0.1:587"),
		health.TCP(ctx, "imap", "127.0.0.1:143"),
		health.TCP(ctx, "pop3", "127.0.0.1:110"),
		health.TCP(ctx, "rspamd", "127.0.0.1:11334"),
		health.TCP(ctx, "redis", "127.0.0.1:6379"),
	}
	for _, result := range results {
		if result.OK {
			fmt.Printf("ok %s\n", result.Name)
			continue
		}
		fmt.Printf("fail %s: %s\n", result.Name, result.Err)
	}
}

func must(data []byte, err error) []byte {
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func writeRendered(path string, data []byte) {
	if err := os.WriteFile(path, data, 0640); err != nil {
		log.Fatalf("write %s: %v", path, err)
	}
}
```

- [ ] **Step 3: Run tests and CLI help**

Run:

```powershell
go test ./...
go run ./cmd/mailctl
```

Expected: tests pass, CLI exits with usage text and status 2 when no command is provided.

- [ ] **Step 4: Commit**

Run:

```powershell
git add internal/health cmd/mailctl
git commit -m "feat: add mail control CLI"
```

## Task 8: Add Debian 13 Deployment Assets

**Files:**
- Create: `deploy/devmail/install-packages.sh`
- Create: `deploy/devmail/proidentity-mail.env.example`
- Create: `deploy/devmail/proidentity-webadmin.service`
- Create: `deploy/devmail/proidentity-mailctl.service`
- Create: `deploy/devmail/README.md`

- [ ] **Step 1: Create package install script**

Create `deploy/devmail/install-packages.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

apt-get update
DEBIAN_FRONTEND=noninteractive apt-get install -y \
  ca-certificates \
  curl \
  mariadb-server \
  redis-server \
  postfix \
  postfix-mysql \
  dovecot-core \
  dovecot-imapd \
  dovecot-pop3d \
  dovecot-lmtpd \
  dovecot-mysql \
  dovecot-sieve \
  dovecot-managesieved \
  rspamd \
  clamav-daemon \
  clamav-freshclam

systemctl enable mariadb redis-server postfix dovecot rspamd clamav-daemon
```

- [ ] **Step 2: Create env example**

Create `deploy/devmail/proidentity-mail.env.example`:

```dotenv
PROIDENTITY_HTTP_ADDR=127.0.0.1:8080
PROIDENTITY_DB_DSN=proidentity_mail:change-me@tcp(127.0.0.1:3306)/proidentity_mail?parseTime=true
PROIDENTITY_CONFIG_DIR=/etc/proidentity-mail/generated
```

- [ ] **Step 3: Create webadmin systemd unit**

Create `deploy/devmail/proidentity-webadmin.service`:

```ini
[Unit]
Description=ProIdentity Mail Webadmin
After=network-online.target mariadb.service
Wants=network-online.target

[Service]
Type=simple
User=proidentity
Group=proidentity
EnvironmentFile=/etc/proidentity-mail/proidentity-mail.env
ExecStart=/opt/proidentity-mail/bin/webadmin
Restart=on-failure
RestartSec=5s
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full
ProtectHome=true
ReadWritePaths=/etc/proidentity-mail

[Install]
WantedBy=multi-user.target
```

- [ ] **Step 4: Create mailctl one-shot systemd unit**

Create `deploy/devmail/proidentity-mailctl.service`:

```ini
[Unit]
Description=ProIdentity Mail Config Render
After=mariadb.service

[Service]
Type=oneshot
User=proidentity
Group=proidentity
EnvironmentFile=/etc/proidentity-mail/proidentity-mail.env
ExecStart=/opt/proidentity-mail/bin/mailctl render
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full
ProtectHome=true
ReadWritePaths=/etc/proidentity-mail
```

- [ ] **Step 5: Create deployment README**

Create `deploy/devmail/README.md`:

```markdown
# DevMail Deployment

Target:

- Host: `root@192.168.254.125`
- OS: Debian GNU/Linux 13 trixie

## Package Install

```bash
cd /opt/proidentity-mail
bash deploy/devmail/install-packages.sh
```

## Runtime User

```bash
useradd --system --home /opt/proidentity-mail --shell /usr/sbin/nologin proidentity
mkdir -p /etc/proidentity-mail/generated /opt/proidentity-mail/bin
chown -R proidentity:proidentity /etc/proidentity-mail /opt/proidentity-mail
chmod 750 /etc/proidentity-mail
```

## Database

```sql
CREATE DATABASE proidentity_mail CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'proidentity_mail'@'127.0.0.1' IDENTIFIED BY 'change-me';
GRANT ALL PRIVILEGES ON proidentity_mail.* TO 'proidentity_mail'@'127.0.0.1';
FLUSH PRIVILEGES;
```

## Services

Copy `proidentity-mail.env.example` to `/etc/proidentity-mail/proidentity-mail.env`, edit secrets, install units into `/etc/systemd/system`, then run:

```bash
systemctl daemon-reload
systemctl enable --now proidentity-webadmin
systemctl start proidentity-mailctl
```
```

- [ ] **Step 6: Run shell syntax check**

Run:

```powershell
go test ./...
```

Expected: PASS. The shell script will be validated on the Debian host before execution.

- [ ] **Step 7: Commit**

Run:

```powershell
git add deploy/devmail
git commit -m "feat: add Debian deployment assets"
```

## Task 9: Add Seed Command For First Tenant

**Files:**
- Modify: `cmd/mailctl/main.go`

- [ ] **Step 1: Add seed command implementation**

Modify `cmd/mailctl/main.go` so the command switch includes:

```go
	case "seed-dev":
		runSeedDev(cfg)
```

Add this function:

```go
func runSeedDev(cfg app.Config) {
	if cfg.DBDSN == "" {
		log.Fatal("PROIDENTITY_DB_DSN is required")
	}
	ctx := context.Background()
	conn, err := db.Open(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `INSERT IGNORE INTO tenants(name, slug, status) VALUES ('Development Tenant', 'dev', 'active')`); err != nil {
		log.Fatalf("insert tenant: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `INSERT IGNORE INTO tenant_policies(tenant_id) SELECT id FROM tenants WHERE slug = 'dev'`); err != nil {
		log.Fatalf("insert tenant policy: %v", err)
	}
	fmt.Println("seeded development tenant")
}
```

Update usage text:

```go
fmt.Fprintln(os.Stderr, "usage: mailctl migrate|render|health|seed-dev")
```

- [ ] **Step 2: Run tests**

Run:

```powershell
go test ./...
```

Expected: PASS.

- [ ] **Step 3: Commit**

Run:

```powershell
git add cmd/mailctl/main.go
git commit -m "feat: add development seed command"
```

## Task 10: Verify On DevMail Host

**Files:**
- Modify only if verification exposes a concrete bug in files from earlier tasks.

- [ ] **Step 1: Build binaries locally**

Run:

```powershell
go test ./...
go build -o .\bin\webadmin.exe .\cmd\webadmin
go build -o .\bin\mailctl.exe .\cmd\mailctl
```

Expected: tests pass and binaries are created.

- [ ] **Step 2: Build Linux binaries**

Run:

```powershell
$env:GOOS='linux'; $env:GOARCH='amd64'; go build -o .\bin\webadmin-linux-amd64 .\cmd\webadmin; go build -o .\bin\mailctl-linux-amd64 .\cmd\mailctl; Remove-Item Env:\GOOS,Env:\GOARCH
```

Expected: Linux binaries are created in `bin`.

- [ ] **Step 3: Copy binaries and deployment files to DevMail**

Run:

```powershell
scp .\bin\webadmin-linux-amd64 root@192.168.254.125:/tmp/webadmin
scp .\bin\mailctl-linux-amd64 root@192.168.254.125:/tmp/mailctl
scp -r .\deploy\devmail root@192.168.254.125:/tmp/proidentity-devmail
```

Expected: files are copied successfully.

- [ ] **Step 4: Inspect target host before installing**

Run:

```powershell
ssh root@192.168.254.125 "hostname; cat /etc/os-release | sed -n '1,8p'; systemctl is-active mariadb redis-server postfix dovecot rspamd clamav-daemon 2>/dev/null || true"
```

Expected: host is `DevMail`; inactive services are acceptable before package installation.

- [ ] **Step 5: Install packages on DevMail**

Package installation on `root@192.168.254.125` is approved by the project owner. Run:

```powershell
ssh root@192.168.254.125 "bash /tmp/proidentity-devmail/install-packages.sh"
```

Expected: packages install and services are enabled. If this changes the production-like host, record the exact package output summary in the final report.

- [ ] **Step 6: Run health check**

Run after packages are installed and binaries are placed:

```powershell
ssh root@192.168.254.125 "chmod +x /tmp/mailctl; /tmp/mailctl health"
```

Expected: health reports available local service ports. Failures identify which daemon needs configuration next.

- [ ] **Step 7: Commit verification fixes**

If code or deploy files changed during verification, run:

```powershell
git add .
git commit -m "fix: address DevMail foundation verification"
```

## Self-Review Notes

- Spec coverage: this plan covers the first implementation slice from the approved design: Go scaffold, MariaDB schema, admin API, config rendering, deployment scripts, baseline service integration path, and DevMail verification. Webmail, calendar, contacts, CardDAV, CalDAV, JMAP, and ActiveSync are intentionally excluded for later plans.
- Completeness scan: the plan uses concrete files, commands, and expected outcomes throughout.
- Type consistency: package names, functions, and command names are consistent across tasks: `app.LoadConfig`, `db.Open`, `db.Migrate`, `admin.NewRouter`, `render.RenderPostfixMain`, `mailctl migrate|render|health|seed-dev`.
