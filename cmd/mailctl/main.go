package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/textproto"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"proidentity-mail/internal/admin"
	"proidentity-mail/internal/app"
	"proidentity-mail/internal/db"
	"proidentity-mail/internal/health"
	"proidentity-mail/internal/quarantine"
	"proidentity-mail/internal/render"
)

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: mailctl migrate|render|health|seed-dev|rotate-admin-password|quarantine-message|release-quarantine|sync-rspamd-policy")
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
	case "seed-dev":
		runSeedDev(cfg)
	case "rotate-admin-password":
		runRotateAdminPassword(cfg)
	case "quarantine-message":
		runQuarantineMessage(cfg, flag.Args()[1:])
	case "release-quarantine":
		runReleaseQuarantine(cfg, flag.Args()[1:])
	case "sync-rspamd-policy":
		runSyncRspamdPolicy(cfg, flag.Args()[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", flag.Arg(0))
		os.Exit(2)
	}
}

func runRotateAdminPassword(cfg app.Config) {
	password := os.Getenv("PROIDENTITY_NEW_ADMIN_PASSWORD")
	if password == "" {
		password = randomHex(18)
	}
	envPath := os.Getenv("PROIDENTITY_ENV_FILE")
	if envPath == "" {
		envPath = "/etc/proidentity-mail/proidentity-mail.env"
	}
	data, err := os.ReadFile(envPath)
	if err != nil {
		log.Fatalf("read env file: %v", err)
	}
	lines := strings.Split(string(data), "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, "PROIDENTITY_ADMIN_PASSWORD=") {
			lines[i] = "PROIDENTITY_ADMIN_PASSWORD=" + password
			found = true
		}
	}
	if !found {
		lines = append(lines, "PROIDENTITY_ADMIN_PASSWORD="+password)
	}
	if err := os.WriteFile(envPath, []byte(strings.Join(lines, "\n")), 0640); err != nil {
		log.Fatalf("write env file: %v", err)
	}
	fmt.Println("admin password rotated")
	if os.Getenv("PROIDENTITY_PRINT_NEW_ADMIN_PASSWORD") == "1" {
		fmt.Println(password)
	}
	_ = cfg
}

func randomHex(bytesCount int) string {
	buf := make([]byte, bytesCount)
	if _, err := rand.Read(buf); err != nil {
		log.Fatalf("generate password: %v", err)
	}
	return hex.EncodeToString(buf)
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
	if cfg.DBDSN == "" {
		log.Fatal("PROIDENTITY_DB_DSN is required")
	}
	ctx := context.Background()
	conn, err := db.Open(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer conn.Close()
	dkimDomains, err := loadDKIMSigningDomains(ctx, conn)
	if err != nil {
		log.Fatalf("load dkim signing domains: %v", err)
	}

	writeRendered(filepath.Join(cfg.ConfigDir, "postfix-main.cf"), must(render.RenderPostfixMain(render.PostfixMainData{Hostname: cfg.MailHostname})))
	writeRendered(filepath.Join(cfg.ConfigDir, "postfix-master.cf"), must(render.RenderPostfixMaster()))
	writeRendered(filepath.Join(cfg.ConfigDir, "dovecot-sql.conf.ext"), must(render.RenderDovecotSQL(render.DovecotSQLData{Database: cfg.DBName, User: cfg.DBUser, Password: cfg.DBPassword})))
	writeRendered(filepath.Join(cfg.ConfigDir, "dovecot-proidentity.conf"), must(render.RenderDovecotLocal()))
	mysqlData := render.PostfixMySQLData{Database: cfg.DBName, User: cfg.DBUser, Password: cfg.DBPassword}
	writeRendered(filepath.Join(cfg.ConfigDir, "virtual-mailbox-domains.cf"), must(render.RenderPostfixVirtualMailboxDomains(mysqlData)))
	writeRendered(filepath.Join(cfg.ConfigDir, "virtual-mailbox-maps.cf"), must(render.RenderPostfixVirtualMailboxMaps(mysqlData)))
	writeRendered(filepath.Join(cfg.ConfigDir, "virtual-alias-maps.cf"), must(render.RenderPostfixVirtualAliasMaps(mysqlData)))
	writeRendered(filepath.Join(cfg.ConfigDir, "rspamd-local.d-redis.conf"), must(render.RenderRspamdLocal()))
	writeRendered(filepath.Join(cfg.ConfigDir, "rspamd-local.d-antivirus.conf"), must(render.RenderRspamdAntivirus()))
	writeRendered(filepath.Join(cfg.ConfigDir, "rspamd-local.d-dkim_signing.conf"), must(render.RenderRspamdDKIMSigning(render.RspamdDKIMSigningData{Domains: dkimDomains})))
	writeRendered(filepath.Join(cfg.ConfigDir, "rspamd-local.d-actions.conf"), must(render.RenderRspamdActions()))
	writeRendered(filepath.Join(cfg.ConfigDir, "rspamd-local.d-milter_headers.conf"), must(render.RenderRspamdMilterHeaders()))
	policies, err := loadRspamdTenantPolicies(ctx, conn)
	if err != nil {
		log.Fatalf("load rspamd tenant policies: %v", err)
	}
	writeRendered(filepath.Join(cfg.ConfigDir, "rspamd-local.d-settings.conf"), must(render.RenderRspamdTenantSettings(render.RspamdTenantPolicyData{Domains: policies})))
	writeRendered(filepath.Join(cfg.ConfigDir, "rspamd-local.d-force_actions.conf"), must(render.RenderRspamdForceActions(render.RspamdTenantPolicyData{Domains: policies})))
	fmt.Printf("rendered configs to %s\n", cfg.ConfigDir)
}

func loadDKIMSigningDomains(ctx context.Context, conn interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}) ([]render.DKIMSigningDomain, error) {
	rows, err := conn.QueryContext(ctx, `
		SELECT d.name, k.selector, k.key_path
		FROM dkim_keys k
		JOIN domains d ON d.id = k.domain_id
		WHERE k.status = 'active'
		  AND d.status IN ('pending', 'active')
		ORDER BY d.name, k.selector`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []render.DKIMSigningDomain
	for rows.Next() {
		var item render.DKIMSigningDomain
		if err := rows.Scan(&item.Domain, &item.Selector, &item.KeyPath); err != nil {
			return nil, err
		}
		domains = append(domains, item)
	}
	return domains, rows.Err()
}

func loadRspamdTenantPolicies(ctx context.Context, conn interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}) ([]render.RspamdTenantPolicyDomain, error) {
	rows, err := conn.QueryContext(ctx, `
		SELECT d.name, p.spam_action, p.malware_action
		FROM domains d
		JOIN tenant_policies p ON p.tenant_id = d.tenant_id
		WHERE d.status IN ('pending', 'active')
		ORDER BY d.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []render.RspamdTenantPolicyDomain
	for rows.Next() {
		var policy render.RspamdTenantPolicyDomain
		if err := rows.Scan(&policy.Domain, &policy.SpamAction, &policy.MalwareAction); err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}
	return policies, rows.Err()
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
		health.TCP(ctx, "groupware", "127.0.0.1:8081"),
		health.TCP(ctx, "webmail", "127.0.0.1:8082"),
		health.Unix(ctx, "clamav", "/run/clamav/clamd.ctl"),
	}
	for _, result := range results {
		if result.OK {
			fmt.Printf("ok %s\n", result.Name)
			continue
		}
		fmt.Printf("fail %s: %s\n", result.Name, result.Err)
	}
}

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

func runQuarantineMessage(cfg app.Config, args []string) {
	flags := flag.NewFlagSet("quarantine-message", flag.ExitOnError)
	recipient := flags.String("recipient", "", "recipient email address")
	sender := flags.String("sender", "", "sender email address")
	messageID := flags.String("message-id", "", "message id")
	verdict := flags.String("verdict", "malware", "spam, malware, phishing, or policy")
	action := flags.String("action", "quarantine", "reject, quarantine, or mark")
	scanner := flags.String("scanner", "manual", "scanner name")
	symbols := flags.String("symbols", "{}", "scanner symbols JSON")
	if err := flags.Parse(args); err != nil {
		log.Fatalf("parse quarantine-message flags: %v", err)
	}
	if cfg.DBDSN == "" {
		log.Fatal("PROIDENTITY_DB_DSN is required")
	}
	if strings.TrimSpace(*recipient) == "" {
		log.Fatal("-recipient is required")
	}
	switch *verdict {
	case "spam", "malware", "phishing", "policy":
	default:
		log.Fatal("-verdict must be spam, malware, phishing, or policy")
	}
	switch *action {
	case "reject", "quarantine", "mark":
	default:
		log.Fatal("-action must be reject, quarantine, or mark")
	}
	ctx := context.Background()
	conn, err := db.Open(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer conn.Close()
	store := admin.NewSQLStore(conn, quarantine.FileStore{Root: cfg.QuarantineDir, MailRoot: cfg.MailRoot, DeliveryAddr: cfg.ReleaseSMTPAddr})
	event, err := store.StoreQuarantineMessage(ctx, admin.QuarantineMessageInput{
		Recipient:   *recipient,
		Sender:      *sender,
		MessageID:   *messageID,
		Verdict:     *verdict,
		Action:      *action,
		Scanner:     *scanner,
		SymbolsJSON: *symbols,
		Reader:      os.Stdin,
	})
	if err != nil {
		log.Fatalf("store quarantine message: %v", err)
	}
	fmt.Printf("quarantined id=%d recipient=%s verdict=%s size=%d\n", event.ID, event.Recipient, event.Verdict, event.SizeBytes)
}

func runReleaseQuarantine(cfg app.Config, args []string) {
	flags := flag.NewFlagSet("release-quarantine", flag.ExitOnError)
	id := flags.Uint64("id", 0, "quarantine event id")
	note := flags.String("note", "mailctl release", "resolution note")
	if err := flags.Parse(args); err != nil {
		log.Fatalf("parse release-quarantine flags: %v", err)
	}
	if *id == 0 {
		log.Fatal("-id is required")
	}
	if cfg.DBDSN == "" {
		log.Fatal("PROIDENTITY_DB_DSN is required")
	}
	ctx := context.Background()
	conn, err := db.Open(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer conn.Close()
	store := admin.NewSQLStore(conn, quarantine.FileStore{Root: cfg.QuarantineDir, MailRoot: cfg.MailRoot, DeliveryAddr: cfg.ReleaseSMTPAddr})
	event, err := store.ResolveQuarantineEvent(ctx, *id, "released", *note)
	if err != nil {
		var smtpErr *textproto.Error
		if errors.As(err, &smtpErr) {
			log.Fatalf("release quarantine failed: smtp_code=%d", smtpErr.Code)
		}
		log.Fatalf("release quarantine failed: %T", err)
	}
	fmt.Printf("released id=%d status=%s\n", event.ID, event.Status)
}

func runSyncRspamdPolicy(cfg app.Config, args []string) {
	flags := flag.NewFlagSet("sync-rspamd-policy", flag.ExitOnError)
	targetDir := flags.String("target-dir", "/etc/rspamd/local.d", "rspamd local.d directory")
	reload := flags.Bool("reload", false, "reload rspamd after writing policy files")
	if err := flags.Parse(args); err != nil {
		log.Fatalf("parse sync-rspamd-policy flags: %v", err)
	}
	if cfg.DBDSN == "" {
		log.Fatal("PROIDENTITY_DB_DSN is required")
	}
	ctx := context.Background()
	conn, err := db.Open(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer conn.Close()
	policies, err := loadRspamdTenantPolicies(ctx, conn)
	if err != nil {
		log.Fatalf("load rspamd tenant policies: %v", err)
	}
	if err := os.MkdirAll(*targetDir, 0750); err != nil {
		log.Fatalf("create rspamd policy dir: %v", err)
	}
	writeRenderedAtomic(filepath.Join(*targetDir, "settings.conf"), must(render.RenderRspamdTenantSettings(render.RspamdTenantPolicyData{Domains: policies})))
	writeRenderedAtomic(filepath.Join(*targetDir, "force_actions.conf"), must(render.RenderRspamdForceActions(render.RspamdTenantPolicyData{Domains: policies})))
	if *reload {
		if err := reloadRspamd(); err != nil {
			log.Fatalf("reload rspamd: %v", err)
		}
	}
	fmt.Printf("synced rspamd policy domains=%d target=%s\n", len(policies), *targetDir)
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

func writeRenderedAtomic(path string, data []byte) {
	temp := path + ".tmp"
	if err := os.WriteFile(temp, data, 0640); err != nil {
		log.Fatalf("write %s: %v", temp, err)
	}
	if err := os.Rename(temp, path); err != nil {
		log.Fatalf("rename %s: %v", path, err)
	}
}

func reloadRspamd() error {
	if path, err := exec.LookPath("systemctl"); err == nil {
		cmd := exec.Command(path, "reload-or-restart", "rspamd")
		return cmd.Run()
	}
	if path, err := exec.LookPath("rspamadm"); err == nil {
		cmd := exec.Command(path, "control", "reload")
		return cmd.Run()
	}
	return errors.New("no rspamd reload command available")
}
