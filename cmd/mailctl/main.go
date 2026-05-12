package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/textproto"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"proidentity-mail/internal/admin"
	"proidentity-mail/internal/app"
	"proidentity-mail/internal/backup"
	"proidentity-mail/internal/configdrift"
	"proidentity-mail/internal/db"
	"proidentity-mail/internal/domain"
	"proidentity-mail/internal/health"
	"proidentity-mail/internal/quarantine"
	"proidentity-mail/internal/render"
)

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: mailctl migrate|render|config-drift|health|seed-dev|rotate-admin-password|quarantine-message|release-quarantine|sync-rspamd-policy|sync-tls-inventory|process-tls-jobs|cloudflare-cert-credentials|render-proxy|sync-proxy|backup|backup-prune|backup-verify|restore")
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
		runRender(cfg, flag.Args()[1:])
	case "config-drift":
		runConfigDrift(cfg, flag.Args()[1:])
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
	case "sync-tls-inventory":
		runSyncTLSInventory(cfg, flag.Args()[1:])
	case "process-tls-jobs":
		runProcessTLSJobs(cfg, flag.Args()[1:])
	case "cloudflare-cert-credentials":
		runCloudflareCertCredentials(cfg, flag.Args()[1:])
	case "render-proxy":
		runRenderProxy(cfg, flag.Args()[1:])
	case "sync-proxy":
		runSyncProxy(cfg, flag.Args()[1:])
	case "backup":
		runBackup(cfg, flag.Args()[1:])
	case "backup-prune":
		runBackupPrune(flag.Args()[1:])
	case "backup-verify":
		runBackupVerify(flag.Args()[1:])
	case "restore":
		runRestore(cfg, flag.Args()[1:])
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
	envPaths := adminPasswordEnvPaths()
	if len(envPaths) == 0 {
		log.Fatal("no environment file available for admin password rotation")
	}
	for _, envPath := range envPaths {
		if err := rotateAdminPasswordInEnvFile(envPath, password); err != nil {
			log.Fatalf("rotate admin password in %s: %v", envPath, err)
		}
	}
	fmt.Println("admin password rotated")
	if os.Getenv("PROIDENTITY_PRINT_NEW_ADMIN_PASSWORD") == "1" {
		fmt.Println(password)
	}
	_ = cfg
}

func adminPasswordEnvPaths() []string {
	if envPath := strings.TrimSpace(os.Getenv("PROIDENTITY_ENV_FILE")); envPath != "" {
		return []string{envPath}
	}
	candidates := []string{
		"/etc/proidentity-mail/webadmin.env",
		"/etc/proidentity-mail/proidentity-mail.env",
	}
	paths := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			paths = append(paths, candidate)
		}
	}
	return paths
}

func rotateAdminPasswordInEnvFile(envPath, password string) error {
	data, err := os.ReadFile(envPath)
	if err != nil {
		return err
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
		return err
	}
	return nil
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

func runRender(cfg app.Config, args []string) {
	flags := flag.NewFlagSet("render", flag.ExitOnError)
	targetDir := flags.String("target-dir", cfg.ConfigDir, "directory for rendered mail config files")
	if err := flags.Parse(args); err != nil {
		log.Fatalf("parse render flags: %v", err)
	}
	renderMailConfigToDir(cfg, *targetDir)
	fmt.Printf("rendered configs to %s\n", *targetDir)
}

func renderMailConfigToDir(cfg app.Config, targetDir string) {
	if err := os.MkdirAll(targetDir, 0750); err != nil {
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
	store := admin.NewSQLStore(conn).WithDNSSettings(admin.DNSSettings{
		MailHostname:    cfg.MailHostname,
		AdminHostname:   cfg.AdminHostname,
		WebmailHostname: cfg.WebmailHostname,
		PublicIPv4:      cfg.PublicIPv4,
		PublicIPv6:      cfg.PublicIPv6,
		TLSMode:         cfg.TLSMode,
		ForceHTTPS:      cfg.ForceHTTPS,
	})
	mailSettings, err := store.GetMailServerSettings(ctx)
	if err != nil {
		log.Fatalf("load mail server settings: %v", err)
	}
	sniHosts, err := loadMailServerSNIHosts(ctx, conn, mailSettings)
	if err != nil {
		log.Fatalf("load mail server SNI hosts: %v", err)
	}

	writeRendered(filepath.Join(targetDir, "postfix-main.cf"), must(render.RenderPostfixMain(render.PostfixMainData{
		Hostname:    mailHostnameForRender(cfg, mailSettings),
		TLSCertFile: cfg.MailTLSCertPath,
		TLSKeyFile:  cfg.MailTLSKeyPath,
		SNIEnabled:  mailSettings.SNIEnabled,
		SNIMapPath:  "/etc/postfix/proidentity/tls-sni-map",
	})))
	writeRendered(filepath.Join(targetDir, "postfix-master.cf"), must(render.RenderPostfixMaster()))
	writeRendered(filepath.Join(targetDir, "dovecot-sql.conf.ext"), must(render.RenderDovecotSQL(render.DovecotSQLData{Database: cfg.DBName, User: cfg.DBUser, Password: cfg.DBPassword})))
	writeRendered(filepath.Join(targetDir, "dovecot-proidentity.conf"), must(render.RenderDovecotLocal(render.DovecotLocalData{
		TLSCertFile: cfg.MailTLSCertPath,
		TLSKeyFile:  cfg.MailTLSKeyPath,
		SNIHosts:    existingSNIHosts(sniHosts),
		AuthPolicy: render.DovecotAuthPolicyData{
			ServerURL: "http://127.0.0.1:8080/internal/dovecot/auth-policy",
			APIHeader: "X-ProIdentity-Auth-Policy: " + cfg.AuthPolicyToken,
			Nonce:     cfg.AuthPolicyNonce,
		},
	})))
	writeRendered(filepath.Join(targetDir, "postfix-tls-sni-map"), must(render.RenderPostfixSNIMap(existingSNIHosts(sniHosts))))
	writeRendered(filepath.Join(targetDir, "tls-sni-source-map"), []byte(renderSNISourceMap(sniHosts)))
	mysqlData := render.PostfixMySQLData{Database: cfg.DBName, User: cfg.DBUser, Password: cfg.DBPassword}
	writeRendered(filepath.Join(targetDir, "virtual-mailbox-domains.cf"), must(render.RenderPostfixVirtualMailboxDomains(mysqlData)))
	writeRendered(filepath.Join(targetDir, "virtual-mailbox-maps.cf"), must(render.RenderPostfixVirtualMailboxMaps(mysqlData)))
	writeRendered(filepath.Join(targetDir, "virtual-alias-maps.cf"), must(render.RenderPostfixVirtualAliasMaps(mysqlData)))
	writeRendered(filepath.Join(targetDir, "sender-login-maps.cf"), must(render.RenderPostfixSenderLoginMaps(mysqlData)))
	writeRendered(filepath.Join(targetDir, "rspamd-local.d-redis.conf"), must(render.RenderRspamdLocal()))
	writeRendered(filepath.Join(targetDir, "rspamd-local.d-antivirus.conf"), must(render.RenderRspamdAntivirus()))
	writeRendered(filepath.Join(targetDir, "rspamd-local.d-dkim_signing.conf"), must(render.RenderRspamdDKIMSigning(render.RspamdDKIMSigningData{Domains: dkimDomains})))
	writeRendered(filepath.Join(targetDir, "rspamd-local.d-actions.conf"), must(render.RenderRspamdActions()))
	writeRendered(filepath.Join(targetDir, "rspamd-local.d-milter_headers.conf"), must(render.RenderRspamdMilterHeaders()))
	policies, err := loadRspamdTenantPolicies(ctx, conn)
	if err != nil {
		log.Fatalf("load rspamd tenant policies: %v", err)
	}
	writeRendered(filepath.Join(targetDir, "rspamd-local.d-settings.conf"), must(render.RenderRspamdTenantSettings(render.RspamdTenantPolicyData{Domains: policies})))
	writeRendered(filepath.Join(targetDir, "rspamd-local.d-force_actions.conf"), must(render.RenderRspamdForceActions(render.RspamdTenantPolicyData{Domains: policies})))
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

func loadMailServerSNIHosts(ctx context.Context, conn interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, settings domain.MailServerSettings) ([]render.MailServerSNIHost, error) {
	if !settings.SNIEnabled {
		return nil, nil
	}
	rows, err := conn.QueryContext(ctx, `
		SELECT name
		FROM domains
		WHERE status IN ('pending', 'active')
		ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	seen := map[string]bool{}
	hosts := make([]render.MailServerSNIHost, 0)
	for rows.Next() {
		var domainName string
		if err := rows.Scan(&domainName); err != nil {
			return nil, err
		}
		host := "mail." + strings.ToLower(strings.TrimSpace(domainName))
		if settings.HostnameMode == "shared" && host == settings.EffectiveHostname {
			continue
		}
		if seen[host] {
			continue
		}
		seen[host] = true
		hosts = append(hosts, render.MailServerSNIHost{
			Hostname:     host,
			TLSChainFile: filepath.Join("/etc/postfix/proidentity/tls-sni", host+".pem"),
			TLSCertFile:  filepath.Join("/etc/letsencrypt/live", host, "fullchain.pem"),
			TLSKeyFile:   filepath.Join("/etc/letsencrypt/live", host, "privkey.pem"),
		})
	}
	return hosts, rows.Err()
}

func existingSNIHosts(hosts []render.MailServerSNIHost) []render.MailServerSNIHost {
	out := make([]render.MailServerSNIHost, 0, len(hosts))
	for _, host := range hosts {
		if fileReadable(host.TLSCertFile) && fileReadable(host.TLSKeyFile) {
			out = append(out, host)
		}
	}
	return out
}

func fileReadable(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func renderSNISourceMap(hosts []render.MailServerSNIHost) string {
	var builder strings.Builder
	for _, host := range hosts {
		if host.Hostname == "" || host.TLSCertFile == "" || host.TLSKeyFile == "" || host.TLSChainFile == "" {
			continue
		}
		builder.WriteString(host.Hostname)
		builder.WriteByte('\t')
		builder.WriteString(host.TLSCertFile)
		builder.WriteByte('\t')
		builder.WriteString(host.TLSKeyFile)
		builder.WriteByte('\t')
		builder.WriteString(host.TLSChainFile)
		builder.WriteByte('\n')
	}
	return builder.String()
}

func mailHostnameForRender(cfg app.Config, settings domain.MailServerSettings) string {
	if settings.HostnameMode != "per-domain" && strings.TrimSpace(settings.EffectiveHostname) != "" && settings.EffectiveHostname != "mail.<domain>" {
		return settings.EffectiveHostname
	}
	return cfg.MailHostname
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
	store := admin.NewSQLStore(conn, quarantine.FileStore{Root: cfg.QuarantineDir, MailRoot: cfg.MailRoot, DeliveryAddr: cfg.ReleaseSMTPAddr}).WithDNSSettings(admin.DNSSettings{
		MailHostname: cfg.MailHostname,
		PublicIPv4:   cfg.PublicIPv4,
		PublicIPv6:   cfg.PublicIPv6,
	})
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
	store := admin.NewSQLStore(conn, quarantine.FileStore{Root: cfg.QuarantineDir, MailRoot: cfg.MailRoot, DeliveryAddr: cfg.ReleaseSMTPAddr}).WithDNSSettings(admin.DNSSettings{
		MailHostname: cfg.MailHostname,
		PublicIPv4:   cfg.PublicIPv4,
		PublicIPv6:   cfg.PublicIPv6,
	})
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

func runSyncTLSInventory(cfg app.Config, args []string) {
	flags := flag.NewFlagSet("sync-tls-inventory", flag.ExitOnError)
	certRoot := flags.String("cert-root", "/etc/letsencrypt/live", "Let's Encrypt live certificate root")
	if err := flags.Parse(args); err != nil {
		log.Fatalf("parse sync-tls-inventory flags: %v", err)
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
	synced, err := syncTLSInventory(ctx, conn, *certRoot)
	if err != nil {
		log.Fatalf("sync tls inventory: %v", err)
	}
	fmt.Printf("synced tls certificate inventory domains=%d\n", synced)
}

func syncTLSInventory(ctx context.Context, conn *sql.DB, certRoot string) (int, error) {
	rows, err := conn.QueryContext(ctx, `
		SELECT d.id, d.name,
		       COALESCE(s.certificate_name, ''), COALESCE(s.custom_cert_path, ''), COALESCE(s.custom_key_path, ''), COALESCE(s.custom_chain_path, ''),
		       COALESCE(s.use_for_https, 1), COALESCE(s.use_for_mail_sni, 1), COALESCE(s.dns_webmail_alias_enabled, 1), COALESCE(s.dns_admin_alias_enabled, 1),
		       COALESCE(s.include_mail_hostname, 1), COALESCE(s.include_webmail_hostname, 1), COALESCE(s.include_admin_hostname, 1)
		FROM domains d
		LEFT JOIN domain_tls_settings s ON s.domain_id = d.id
		WHERE d.status IN ('pending', 'active')
		ORDER BY d.name`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	synced := 0
	for rows.Next() {
		var item tlsInventoryDomain
		if err := rows.Scan(&item.ID, &item.Name, &item.CertificateName, &item.CustomCertPath, &item.CustomKeyPath, &item.CustomChainPath, &item.UseForHTTPS, &item.UseForMailSNI, &item.DNSWebmailAlias, &item.DNSAdminAlias, &item.IncludeMail, &item.IncludeWebmail, &item.IncludeAdmin); err != nil {
			return synced, err
		}
		cert, ok := findTLSInventoryCertificate(item, certRoot)
		if !ok {
			continue
		}
		if err := upsertTLSInventoryCertificate(ctx, conn, cert); err != nil {
			return synced, fmt.Errorf("store certificate for %s: %w", item.Name, err)
		}
		synced++
	}
	if err := rows.Err(); err != nil {
		return synced, err
	}
	return synced, nil
}

func runProcessTLSJobs(cfg app.Config, args []string) {
	flags := flag.NewFlagSet("process-tls-jobs", flag.ExitOnError)
	limit := flags.Int("limit", 1, "maximum queued jobs to process")
	certRoot := flags.String("cert-root", "/etc/letsencrypt/live", "Let's Encrypt live certificate root")
	applyConfig := flags.Bool("apply-config", true, "render and reload proxy/mail services after successful certbot")
	if err := flags.Parse(args); err != nil {
		log.Fatalf("parse process-tls-jobs flags: %v", err)
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
	processed := 0
	for processed < *limit {
		job, ok, err := claimTLSJob(ctx, conn)
		if err != nil {
			log.Fatalf("claim tls job: %v", err)
		}
		if !ok {
			break
		}
		if err := processTLSJob(ctx, conn, cfg, job, *certRoot, *applyConfig); err != nil {
			_ = updateTLSJobFailed(ctx, conn, job.ID, err)
			log.Printf("tls job %d failed: %v", job.ID, err)
		}
		processed++
	}
	fmt.Printf("processed tls jobs=%d\n", processed)
}

type tlsQueuedJob struct {
	ID            uint64
	DomainID      uint64
	Domain        string
	JobType       string
	ChallengeType string
	Hostnames     []string
}

func claimTLSJob(ctx context.Context, conn *sql.DB) (tlsQueuedJob, bool, error) {
	var job tlsQueuedJob
	var hostnamesJSON string
	err := conn.QueryRowContext(ctx, `
		SELECT j.id, j.domain_id, d.name, j.job_type, j.challenge_type, CAST(j.hostnames_json AS CHAR)
		FROM tls_certificate_jobs j
		JOIN domains d ON d.id = j.domain_id
		WHERE j.status = 'queued'
		ORDER BY j.created_at, j.id
		LIMIT 1`).Scan(&job.ID, &job.DomainID, &job.Domain, &job.JobType, &job.ChallengeType, &hostnamesJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return tlsQueuedJob{}, false, nil
	}
	if err != nil {
		return tlsQueuedJob{}, false, err
	}
	_ = json.Unmarshal([]byte(hostnamesJSON), &job.Hostnames)
	result, err := conn.ExecContext(ctx, `
		UPDATE tls_certificate_jobs
		SET status = 'running', step = 'preparing', progress = 5, message = 'TLS worker claimed the job.', started_at = COALESCE(started_at, CURRENT_TIMESTAMP), error = ''
		WHERE id = ? AND status = 'queued'`, job.ID)
	if err != nil {
		return tlsQueuedJob{}, false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return tlsQueuedJob{}, false, err
	}
	return job, affected > 0, nil
}

func processTLSJob(ctx context.Context, conn *sql.DB, cfg app.Config, job tlsQueuedJob, certRoot string, applyConfig bool) error {
	hostnames := uniqueHosts(job.Hostnames)
	if len(hostnames) == 0 {
		return errors.New("tls job has no hostnames")
	}
	switch job.JobType {
	case "check":
		_ = updateTLSJobProgress(ctx, conn, job.ID, "syncing inventory", 60, "Checking certificate files and refreshing inventory.")
	case "deploy":
		_ = updateTLSJobProgress(ctx, conn, job.ID, "deploying", 70, "Reapplying proxy and mail TLS configuration.")
	default:
		if err := runCertbotForTLSJob(ctx, conn, cfg, job, hostnames); err != nil {
			return err
		}
	}
	_ = updateTLSJobProgress(ctx, conn, job.ID, "syncing inventory", 70, "Recording issued certificate metadata.")
	if _, err := syncTLSInventory(ctx, conn, certRoot); err != nil {
		return err
	}
	if applyConfig {
		_ = updateTLSJobProgress(ctx, conn, job.ID, "deploying nginx", 82, "Rendering and reloading Nginx proxy TLS config.")
		if err := runCommand(ctx, "/opt/proidentity-mail/bin/mailctl", "sync-proxy", "--reload"); err != nil {
			return err
		}
		_ = updateTLSJobProgress(ctx, conn, job.ID, "deploying mail services", 92, "Rendering Postfix/Dovecot SNI maps and restarting mail services.")
		if err := runCommand(ctx, "/opt/proidentity-mail/bin/mailctl", "render"); err != nil {
			return err
		}
		if err := runCommand(ctx, "/opt/proidentity-mail/bin/apply-mail-config"); err != nil {
			return err
		}
	}
	_, err := conn.ExecContext(ctx, `
		UPDATE tls_certificate_jobs
		SET status = 'succeeded', step = 'done', progress = 100, message = 'TLS certificate job completed.', finished_at = CURRENT_TIMESTAMP
		WHERE id = ?`, job.ID)
	return err
}

func runCertbotForTLSJob(ctx context.Context, conn *sql.DB, cfg app.Config, job tlsQueuedJob, hostnames []string) error {
	_ = updateTLSJobProgress(ctx, conn, job.ID, "requesting certificate", 20, "Starting certbot.")
	args := []string{"certonly", "--non-interactive", "--agree-tos", "--register-unsafely-without-email", "--cert-name", hostnames[0]}
	switch job.ChallengeType {
	case "http-01":
		args = append(args, "--webroot", "-w", valueOrDefault(cfg.ACMEWebroot, "/var/lib/proidentity-mail/acme"))
	case "dns-cloudflare":
		credentialsPath, cleanup, err := cloudflareCredentialsForTLSJob(ctx, conn, job)
		if err != nil {
			return err
		}
		defer cleanup()
		args = append(args, "--dns-cloudflare", "--dns-cloudflare-credentials", credentialsPath, "--dns-cloudflare-propagation-seconds", "60")
	case "manual-dns", "custom-import", "none":
		return fmt.Errorf("challenge type %s is not automated by the TLS worker yet", job.ChallengeType)
	default:
		return fmt.Errorf("unsupported challenge type %s", job.ChallengeType)
	}
	if job.JobType == "renew" {
		args = append(args, "--force-renewal")
	}
	for _, host := range hostnames {
		args = append(args, "-d", host)
	}
	if err := runCommand(ctx, "certbot", args...); err != nil {
		return err
	}
	return updateTLSJobProgress(ctx, conn, job.ID, "certificate issued", 65, "Certbot finished successfully.")
}

func cloudflareCredentialsForTLSJob(ctx context.Context, conn *sql.DB, job tlsQueuedJob) (string, func(), error) {
	var token string
	err := conn.QueryRowContext(ctx, `
		SELECT api_token
		FROM cloudflare_domain_configs
		WHERE domain_id = ? AND api_token <> ''
		LIMIT 1`, job.DomainID).Scan(&token)
	if errors.Is(err, sql.ErrNoRows) {
		return "", func() {}, fmt.Errorf("no Cloudflare token stored for %s", job.Domain)
	}
	if err != nil {
		return "", func() {}, err
	}
	file, err := os.CreateTemp("", fmt.Sprintf("proidentity-cf-%d-*.ini", job.ID))
	if err != nil {
		return "", func() {}, err
	}
	path := file.Name()
	if _, err := file.WriteString("dns_cloudflare_api_token = " + token + "\n"); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", func() {}, err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", func() {}, err
	}
	if err := os.Chmod(path, 0600); err != nil {
		_ = os.Remove(path)
		return "", func() {}, err
	}
	return path, func() { _ = os.Remove(path) }, nil
}

func updateTLSJobProgress(ctx context.Context, conn *sql.DB, jobID uint64, step string, progress int, message string) error {
	_, err := conn.ExecContext(ctx, `
		UPDATE tls_certificate_jobs
		SET step = ?, progress = ?, message = ?
		WHERE id = ?`, step, progress, message, jobID)
	return err
}

func updateTLSJobFailed(ctx context.Context, conn *sql.DB, jobID uint64, cause error) error {
	_, err := conn.ExecContext(ctx, `
		UPDATE tls_certificate_jobs
		SET status = 'failed', step = 'failed', progress = 100, error = ?, message = 'TLS certificate job failed.', finished_at = CURRENT_TIMESTAMP
		WHERE id = ?`, truncateForDB(cause.Error(), 1000), jobID)
	return err
}

func runCommand(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s failed: %w: %s", name, strings.Join(args, " "), err, truncateForDB(string(output), 800))
	}
	return nil
}

func truncateForDB(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}

type tlsInventoryDomain struct {
	ID              uint64
	Name            string
	CertificateName string
	CustomCertPath  string
	CustomKeyPath   string
	CustomChainPath string
	UseForHTTPS     bool
	UseForMailSNI   bool
	DNSWebmailAlias bool
	DNSAdminAlias   bool
	IncludeMail     bool
	IncludeWebmail  bool
	IncludeAdmin    bool
}

type tlsInventoryCertificate struct {
	DomainID          uint64
	Source            string
	Status            string
	CommonName        string
	SANs              []string
	CertPath          string
	KeyPath           string
	ChainPath         string
	Issuer            string
	Serial            string
	FingerprintSHA256 string
	NotBefore         time.Time
	NotAfter          time.Time
	UsedForHTTPS      bool
	UsedForMailSNI    bool
}

func findTLSInventoryCertificate(item tlsInventoryDomain, certRoot string) (tlsInventoryCertificate, bool) {
	candidates := tlsInventoryCandidateNames(item)
	for _, name := range candidates {
		certPath := strings.TrimSpace(item.CustomCertPath)
		keyPath := strings.TrimSpace(item.CustomKeyPath)
		chainPath := strings.TrimSpace(item.CustomChainPath)
		source := "custom"
		if certPath == "" {
			source = "letsencrypt"
			certPath = filepath.Join(certRoot, name, "fullchain.pem")
			keyPath = filepath.Join(certRoot, name, "privkey.pem")
			chainPath = filepath.Join(certRoot, name, "chain.pem")
		}
		cert, ok := parseTLSInventoryCertificate(item.ID, source, certPath, keyPath, chainPath)
		if ok {
			cert.UsedForHTTPS = item.UseForHTTPS
			cert.UsedForMailSNI = item.UseForMailSNI
			return cert, true
		}
		if strings.TrimSpace(item.CustomCertPath) != "" {
			break
		}
	}
	return tlsInventoryCertificate{}, false
}

func tlsInventoryCandidateNames(item tlsInventoryDomain) []string {
	domainName := normalizeHost(item.Name)
	candidates := []string{item.CertificateName}
	if item.IncludeMail {
		candidates = append(candidates, "mail."+domainName)
	}
	if item.DNSWebmailAlias && item.IncludeWebmail {
		candidates = append(candidates, "webmail."+domainName)
	}
	if item.DNSAdminAlias && item.IncludeAdmin {
		candidates = append(candidates, "madmin."+domainName)
	}
	candidates = append(candidates, "madmin."+domainName, "webmail."+domainName, "mail."+domainName)
	return uniqueHosts(candidates)
}

func parseTLSInventoryCertificate(domainID uint64, source, certPath, keyPath, chainPath string) (tlsInventoryCertificate, bool) {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return tlsInventoryCertificate{}, false
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return tlsInventoryCertificate{}, false
	}
	parsed, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return tlsInventoryCertificate{}, false
	}
	fingerprint := sha256.Sum256(parsed.Raw)
	status := "active"
	if time.Now().After(parsed.NotAfter) {
		status = "expired"
	} else if time.Until(parsed.NotAfter) < 30*24*time.Hour {
		status = "expiring"
	}
	return tlsInventoryCertificate{
		DomainID:          domainID,
		Source:            source,
		Status:            status,
		CommonName:        parsed.Subject.CommonName,
		SANs:              parsed.DNSNames,
		CertPath:          certPath,
		KeyPath:           keyPath,
		ChainPath:         chainPath,
		Issuer:            parsed.Issuer.String(),
		Serial:            parsed.SerialNumber.String(),
		FingerprintSHA256: strings.ToUpper(hex.EncodeToString(fingerprint[:])),
		NotBefore:         parsed.NotBefore,
		NotAfter:          parsed.NotAfter,
	}, true
}

func upsertTLSInventoryCertificate(ctx context.Context, conn *sql.DB, cert tlsInventoryCertificate) error {
	sansJSON, err := json.Marshal(cert.SANs)
	if err != nil {
		return err
	}
	result, err := conn.ExecContext(ctx, `
		UPDATE tls_certificates
		SET source = ?, status = ?, common_name = ?, sans_json = ?, key_path = ?, chain_path = ?, issuer = ?, serial = ?, fingerprint_sha256 = ?,
		    not_before = ?, not_after = ?, used_for_https = ?, used_for_mail_sni = ?, last_error = ''
		WHERE domain_id = ? AND cert_path = ?`,
		cert.Source, cert.Status, cert.CommonName, string(sansJSON), cert.KeyPath, cert.ChainPath, cert.Issuer, cert.Serial, cert.FingerprintSHA256,
		cert.NotBefore, cert.NotAfter, cert.UsedForHTTPS, cert.UsedForMailSNI, cert.DomainID, cert.CertPath)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}
	_, err = conn.ExecContext(ctx, `
		INSERT INTO tls_certificates(domain_id, source, status, common_name, sans_json, cert_path, key_path, chain_path, issuer, serial, fingerprint_sha256, not_before, not_after, used_for_https, used_for_mail_sni)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cert.DomainID, cert.Source, cert.Status, cert.CommonName, string(sansJSON), cert.CertPath, cert.KeyPath, cert.ChainPath, cert.Issuer, cert.Serial, cert.FingerprintSHA256, cert.NotBefore, cert.NotAfter, cert.UsedForHTTPS, cert.UsedForMailSNI)
	return err
}

func normalizeHost(value string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
}

func uniqueHosts(values []string) []string {
	seen := make(map[string]bool)
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = normalizeHost(value)
		if value == "" || !strings.Contains(value, ".") || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func runCloudflareCertCredentials(cfg app.Config, args []string) {
	flags := flag.NewFlagSet("cloudflare-cert-credentials", flag.ExitOnError)
	domainName := flags.String("domain", cfg.CloudflareCertDomain, "domain whose stored Cloudflare token should be exported for certbot")
	output := flags.String("output", valueOrDefault(cfg.CloudflareCredentialsFile, "/etc/proidentity-mail/cloudflare.ini"), "certbot dns_cloudflare credentials file")
	if err := flags.Parse(args); err != nil {
		log.Fatalf("parse cloudflare-cert-credentials flags: %v", err)
	}
	if strings.TrimSpace(*domainName) == "" {
		log.Fatal("-domain or PROIDENTITY_CLOUDFLARE_CERT_DOMAIN is required")
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
	var token string
	err = conn.QueryRowContext(ctx, `
		SELECT c.api_token
		FROM cloudflare_domain_configs c
		JOIN domains d ON d.id = c.domain_id
		WHERE d.name = ? AND c.api_token <> ''
		LIMIT 1`, strings.ToLower(strings.TrimSpace(*domainName))).Scan(&token)
	if errors.Is(err, sql.ErrNoRows) {
		log.Fatalf("no Cloudflare token stored for domain %s", *domainName)
	}
	if err != nil {
		log.Fatalf("load Cloudflare token: %v", err)
	}
	if err := writeSecretFile(*output, "dns_cloudflare_api_token = "+token+"\n"); err != nil {
		log.Fatalf("write Cloudflare credentials: %v", err)
	}
	fmt.Printf("wrote Cloudflare certbot credentials for %s to %s\n", *domainName, *output)
}

func runRenderProxy(cfg app.Config, args []string) {
	flags := flag.NewFlagSet("render-proxy", flag.ExitOnError)
	targetDir := flags.String("target-dir", filepath.Join(cfg.ConfigDir, "proxy"), "directory for rendered proxy files")
	if err := flags.Parse(args); err != nil {
		log.Fatalf("parse render-proxy flags: %v", err)
	}
	writeProxyFiles(cfg, *targetDir)
	fmt.Printf("rendered proxy config to %s\n", *targetDir)
}

func runSyncProxy(cfg app.Config, args []string) {
	flags := flag.NewFlagSet("sync-proxy", flag.ExitOnError)
	nginxConf := flags.String("nginx-conf", "/etc/nginx/conf.d/proidentity.conf", "nginx server config path")
	commonDir := flags.String("common-dir", "/etc/nginx/proidentity", "nginx shared include directory")
	certScript := flags.String("cert-script", "/opt/proidentity-mail/bin/proidentity-issue-cert", "certbot helper script path")
	reload := flags.Bool("reload", false, "test and reload nginx after writing config")
	if err := flags.Parse(args); err != nil {
		log.Fatalf("parse sync-proxy flags: %v", err)
	}
	if cfg.ProxyMode != "internal-nginx" {
		log.Fatalf("unsupported proxy mode %q", cfg.ProxyMode)
	}
	if err := os.MkdirAll(filepath.Dir(*nginxConf), 0755); err != nil {
		log.Fatalf("create nginx config dir: %v", err)
	}
	if err := os.MkdirAll(*commonDir, 0755); err != nil {
		log.Fatalf("create nginx common dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(*certScript), 0755); err != nil {
		log.Fatalf("create cert helper dir: %v", err)
	}
	if err := os.MkdirAll(cfg.ACMEWebroot, 0755); err != nil {
		log.Fatalf("create acme webroot: %v", err)
	}
	mailSettings := loadProxyMailServerSettings(cfg)
	writeRenderedAtomic(*nginxConf, must(render.RenderNginxProxy(proxyRenderData(cfg, mailSettings))))
	chmodLiveFile(*nginxConf, 0644)
	commonPath := filepath.Join(*commonDir, "proxy-common.conf")
	writeRenderedAtomic(commonPath, must(render.RenderNginxProxyCommon()))
	chmodLiveFile(commonPath, 0644)
	writeRenderedAtomic(*certScript, must(render.RenderCertbotScript(certbotRenderData(cfg, mailSettings))))
	chmodLiveFile(*certScript, 0750)
	chgrpIfExists(*certScript, "proidentity")
	if *reload {
		if err := testNginx(); err != nil {
			log.Fatalf("nginx config test: %v", err)
		}
		if err := reloadNginx(); err != nil {
			log.Fatalf("reload nginx: %v", err)
		}
	}
	fmt.Printf("synced proxy config nginx=%s tls_mode=%s\n", *nginxConf, cfg.TLSMode)
}

func runConfigDrift(cfg app.Config, args []string) {
	flags := flag.NewFlagSet("config-drift", flag.ExitOnError)
	desiredDir := flags.String("desired-dir", "", "directory for temporary desired render output")
	liveRoot := flags.String("live-root", "", "optional root prefix for live paths, used by tests/staging")
	jsonOutput := flags.Bool("json", false, "write the drift report as JSON")
	if err := flags.Parse(args); err != nil {
		log.Fatalf("parse config-drift flags: %v", err)
	}
	renderRoot := *desiredDir
	cleanup := func() {}
	if renderRoot == "" {
		tempDir, err := os.MkdirTemp("", "proidentity-config-drift-*")
		if err != nil {
			log.Fatalf("create drift render dir: %v", err)
		}
		renderRoot = tempDir
		cleanup = func() { _ = os.RemoveAll(tempDir) }
	}
	defer cleanup()

	mailDir := filepath.Join(renderRoot, "mail")
	proxyDir := filepath.Join(renderRoot, "proxy")
	renderMailConfigToDir(cfg, mailDir)
	writeProxyFiles(cfg, proxyDir)
	report := configdrift.Compare(context.Background(), configdrift.DefaultMappings(mailDir, proxyDir, *liveRoot))
	if *jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			log.Fatalf("encode report: %v", err)
		}
		return
	}
	fmt.Printf("status=%s total=%d matching=%d drifted=%d missing_live=%d errors=%d\n",
		report.Status,
		report.Summary.Total,
		report.Summary.Matching,
		report.Summary.Drifted,
		report.Summary.MissingLive,
		report.Summary.Errors,
	)
	for _, item := range report.Items {
		if item.Status == "match" {
			continue
		}
		fmt.Printf("%s\t%s\t%s -> %s\n", item.Status, item.Label, item.DesiredPath, item.LivePath)
		if item.Error != "" {
			fmt.Printf("  error: %s\n", item.Error)
		}
	}
}

func chmodLiveFile(path string, mode os.FileMode) {
	if err := os.Chmod(path, mode); err != nil {
		log.Fatalf("chmod %s: %v", path, err)
	}
}

func chgrpIfExists(path, groupName string) {
	group, err := user.LookupGroup(groupName)
	if err != nil {
		return
	}
	gid, err := strconv.Atoi(group.Gid)
	if err != nil {
		log.Printf("warning: group %s has non-numeric gid %q", groupName, group.Gid)
		return
	}
	if err := os.Chown(path, -1, gid); err != nil {
		log.Printf("warning: chgrp %s %s failed: %v", groupName, path, err)
	}
}

func runBackup(cfg app.Config, args []string) {
	flags := flag.NewFlagSet("backup", flag.ExitOnError)
	outputDir := flags.String("output-dir", "/var/backups/proidentity-mail", "backup output directory")
	output := flags.String("output", "", "exact backup archive path")
	includeDB := flags.Bool("include-db", true, "include MariaDB logical dump")
	allowPlain := flags.Bool("allow-plain", false, "allow unencrypted backup output; intended only for local debugging")
	pruneAfter := flags.Bool("prune-after", false, "prune old backups after successful backup")
	keepDaily := flags.Int("keep-daily", 7, "daily backups to keep during --prune-after")
	keepWeekly := flags.Int("keep-weekly", 4, "weekly backups to keep during --prune-after")
	keepMonthly := flags.Int("keep-monthly", 12, "monthly backups to keep during --prune-after")
	if err := flags.Parse(args); err != nil {
		log.Fatalf("parse backup flags: %v", err)
	}
	if err := os.MkdirAll(*outputDir, 0750); err != nil {
		log.Fatalf("create backup dir: %v", err)
	}
	outputPath := *output
	if outputPath == "" {
		outputPath = filepath.Join(*outputDir, "proidentity-mail-"+time.Now().UTC().Format("20060102-150405")+".tar.gz.enc")
	}
	encryptionKey, err := backupEncryptionKey(false)
	if err != nil {
		if !*allowPlain {
			log.Fatalf("backup encryption key: %v", err)
		}
		log.Printf("warning: creating unencrypted backup because --allow-plain was set")
	}
	tempDir, err := os.MkdirTemp("", "proidentity-backup-*")
	if err != nil {
		log.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	sources := backupSources(cfg)
	if *includeDB {
		dumpPath := filepath.Join(tempDir, "proidentity.sql")
		if err := dumpDatabase(cfg, dumpPath); err != nil {
			log.Fatalf("database dump failed: %v", err)
		}
		sources = append(sources, backup.Source{Name: "database", Path: dumpPath, Required: true})
	}
	host, _ := os.Hostname()
	manifest, err := backup.Create(context.Background(), backup.Options{OutputPath: outputPath, Sources: sources, Hostname: host, EncryptionKey: encryptionKey})
	if err != nil {
		log.Fatalf("backup create failed: %v", err)
	}
	summary, err := backup.VerifyWithKey(context.Background(), outputPath, encryptionKey)
	if err != nil {
		log.Fatalf("backup verify failed: %v", err)
	}
	fmt.Printf("backup=%s entries=%d files=%d bytes=%d\n", outputPath, len(manifest.Entries), summary.Files, summary.Bytes)
	recordBackupAudit(cfg, outputPath, len(manifest.Entries), summary, os.Getenv("PROIDENTITY_BACKUP_SCHEDULED") == "1")
	if *pruneAfter {
		result, err := backup.Prune(*outputDir, backup.RetentionPolicy{Daily: *keepDaily, Weekly: *keepWeekly, Monthly: *keepMonthly}, backup.PruneOptions{Apply: true})
		if err != nil {
			log.Fatalf("backup prune failed: %v", err)
		}
		fmt.Printf("backup_prune dir=%s scanned=%d kept=%d deleted=%d\n", *outputDir, result.Scanned, result.Kept, result.Deleted)
	}
}

func recordBackupAudit(cfg app.Config, outputPath string, manifestEntries int, summary backup.VerifySummary, scheduled bool) {
	if strings.TrimSpace(cfg.DBDSN) == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := db.Open(ctx, cfg.DBDSN)
	if err != nil {
		log.Printf("warning: backup audit db open failed: %v", err)
		return
	}
	defer conn.Close()
	metadata, err := json.Marshal(map[string]any{
		"archive":          outputPath,
		"archive_name":     filepath.Base(outputPath),
		"scheduled":        scheduled,
		"manifest_entries": manifestEntries,
		"verified_entries": summary.Entries,
		"files":            summary.Files,
		"bytes":            summary.Bytes,
	})
	if err != nil {
		log.Printf("warning: backup audit metadata failed: %v", err)
		return
	}
	for _, action := range backupAuditActions(scheduled) {
		if _, err := conn.ExecContext(ctx, `
			INSERT INTO audit_events(actor_type, action, target_type, target_id, metadata_json)
			VALUES ('system', ?, 'backup', ?, ?)`, action, filepath.Base(outputPath), string(metadata)); err != nil {
			log.Printf("warning: backup audit insert failed action=%s: %v", action, err)
		}
	}
}

func backupAuditActions(scheduled bool) []string {
	if scheduled {
		return []string{"backup.completed"}
	}
	return []string{"backup.completed", "security.alert.backup_manual"}
}

func backupEncryptionKey(allowMissing bool) ([]byte, error) {
	raw := strings.TrimSpace(os.Getenv("PROIDENTITY_BACKUP_ENCRYPTION_KEY"))
	if raw == "" {
		if allowMissing {
			return nil, nil
		}
		return nil, errors.New("PROIDENTITY_BACKUP_ENCRYPTION_KEY is required for encrypted backups")
	}
	if key, err := base64.StdEncoding.DecodeString(raw); err == nil && len(key) == 32 {
		return key, nil
	}
	if key, err := base64.RawStdEncoding.DecodeString(raw); err == nil && len(key) == 32 {
		return key, nil
	}
	if key, err := base64.RawURLEncoding.DecodeString(raw); err == nil && len(key) == 32 {
		return key, nil
	}
	if key, err := hex.DecodeString(raw); err == nil && len(key) == 32 {
		return key, nil
	}
	return nil, errors.New("PROIDENTITY_BACKUP_ENCRYPTION_KEY must decode to 32 bytes")
}

func runBackupPrune(args []string) {
	flags := flag.NewFlagSet("backup-prune", flag.ExitOnError)
	dir := flags.String("dir", "/var/backups/proidentity-mail", "backup directory")
	keepDaily := flags.Int("keep-daily", 7, "daily backups to keep")
	keepWeekly := flags.Int("keep-weekly", 4, "weekly backups to keep")
	keepMonthly := flags.Int("keep-monthly", 12, "monthly backups to keep")
	apply := flags.Bool("apply", false, "delete old backups; without this only reports what would be deleted")
	if err := flags.Parse(args); err != nil {
		log.Fatalf("parse backup-prune flags: %v", err)
	}
	result, err := backup.Prune(*dir, backup.RetentionPolicy{Daily: *keepDaily, Weekly: *keepWeekly, Monthly: *keepMonthly}, backup.PruneOptions{Apply: *apply})
	if err != nil {
		log.Fatalf("backup prune failed: %v", err)
	}
	if *apply {
		fmt.Printf("backup_prune dir=%s scanned=%d kept=%d deleted=%d\n", *dir, result.Scanned, result.Kept, result.Deleted)
		return
	}
	fmt.Printf("backup_prune_dry_run dir=%s scanned=%d kept=%d would_delete=%d\n", *dir, result.Scanned, result.Kept, result.WouldDelete)
}

func runBackupVerify(args []string) {
	flags := flag.NewFlagSet("backup-verify", flag.ExitOnError)
	archive := flags.String("archive", "", "backup archive path")
	if err := flags.Parse(args); err != nil {
		log.Fatalf("parse backup-verify flags: %v", err)
	}
	if *archive == "" {
		log.Fatal("-archive is required")
	}
	encryptionKey, err := backupEncryptionKey(true)
	if err != nil {
		log.Fatalf("backup encryption key: %v", err)
	}
	summary, err := backup.VerifyWithKey(context.Background(), *archive, encryptionKey)
	if err != nil {
		log.Fatalf("backup verification failed: %v", err)
	}
	fmt.Printf("backup_verified archive=%s entries=%d files=%d bytes=%d\n", *archive, summary.Entries, summary.Files, summary.Bytes)
}

func runRestore(cfg app.Config, args []string) {
	flags := flag.NewFlagSet("restore", flag.ExitOnError)
	archive := flags.String("archive", "", "backup archive path")
	targetRoot := flags.String("target-root", "", "restore target root or staging directory")
	stageDir := flags.String("stage-dir", "", "live restore staging directory; defaults to a temporary directory")
	apply := flags.Bool("apply", false, "actually restore; without this restore only verifies")
	live := flags.Bool("live", false, "restore into configured live paths, import database, and restart services")
	overwrite := flags.Bool("overwrite", false, "allow overwrite while extracting")
	importDB := flags.Bool("import-db", true, "import database dump during --live restore")
	serviceControl := flags.Bool("service-control", true, "stop and restart services during --live restore")
	fixPermissions := flags.Bool("fix-permissions", true, "repair ownership and permissions during --live restore")
	healthCheck := flags.Bool("health-check", true, "run mailctl health after --live restore")
	keepStaging := flags.Bool("keep-staging", false, "keep live restore staging directory")
	confirmLive := flags.String("confirm-live-restore", "", "required phrase for --live --apply: RESTORE <archive filename>")
	if err := flags.Parse(args); err != nil {
		log.Fatalf("parse restore flags: %v", err)
	}
	if *archive == "" {
		log.Fatal("-archive is required")
	}
	encryptionKey, err := backupEncryptionKey(true)
	if err != nil {
		log.Fatalf("backup encryption key: %v", err)
	}
	summary, err := backup.VerifyWithKey(context.Background(), *archive, encryptionKey)
	if err != nil {
		log.Fatalf("restore verification failed: %v", err)
	}
	if !*apply {
		fmt.Printf("restore_dry_run archive=%s entries=%d files=%d bytes=%d\n", *archive, summary.Entries, summary.Files, summary.Bytes)
		return
	}
	if *live {
		if err := validateLiveRestoreConfirmation(*archive, *confirmLive); err != nil {
			log.Fatal(err)
		}
		options := backup.LiveRestoreOptions{
			ArchivePath:      *archive,
			StagingDir:       *stageDir,
			EncryptionKey:    encryptionKey,
			Mappings:         liveRestoreMappings(cfg),
			DatabaseName:     cfg.DBName,
			DatabasePassword: cfg.DBPassword,
			ImportDatabase:   *importDB,
			ControlServices:  *serviceControl,
			Services:         liveRestoreServices(),
			Runner:           backup.ExecRunner{},
			Overwrite:        *overwrite,
			KeepStaging:      *keepStaging,
		}
		if *fixPermissions {
			options.Permissions = liveRestorePermissions(cfg)
		}
		result, err := backup.RestoreLive(context.Background(), options)
		if err != nil {
			log.Fatalf("live restore failed: %v", err)
		}
		fmt.Printf("restore_live archive=%s staging=%s files=%d bytes=%d database_imported=%t permissions_fixed=%t services_restarted=%t\n", *archive, result.StagingDir, result.Files, result.Bytes, result.DatabaseImported, result.PermissionsFixed, result.ServicesStarted)
		if *healthCheck {
			runHealth()
		}
		return
	}
	if *targetRoot == "" {
		log.Fatal("-target-root is required when --apply is used")
	}
	if err := backup.RestoreWithKey(context.Background(), *archive, *targetRoot, backup.RestoreOptions{Overwrite: *overwrite}, encryptionKey); err != nil {
		log.Fatalf("restore failed: %v", err)
	}
	fmt.Printf("restore_extracted archive=%s target=%s files=%d bytes=%d\n", *archive, *targetRoot, summary.Files, summary.Bytes)
}

func validateLiveRestoreConfirmation(archive, provided string) error {
	required := liveRestoreConfirmationPhrase(archive)
	if strings.TrimSpace(provided) != required {
		return fmt.Errorf("live restore requires --confirm-live-restore=%q", required)
	}
	return nil
}

func liveRestoreConfirmationPhrase(archive string) string {
	base := filepath.Base(strings.TrimSpace(archive))
	if base == "." || base == string(filepath.Separator) || base == "" {
		base = "backup"
	}
	return "RESTORE " + base
}

func liveRestoreMappings(cfg app.Config) []backup.LiveMapping {
	return []backup.LiveMapping{
		{Source: "etc-proidentity-mail", Target: "/etc/proidentity-mail"},
		{Source: "maildir", Target: cfg.MailRoot},
		{Source: "quarantine", Target: cfg.QuarantineDir},
		{Source: "generated-config", Target: cfg.ConfigDir},
		{Source: "rspamd-dkim", Target: "/var/lib/rspamd/dkim"},
		{Source: "nginx-proidentity", Target: "/etc/nginx/proidentity"},
		{Source: "nginx-conf/proidentity.conf", Target: "/etc/nginx/conf.d/proidentity.conf"},
		{Source: "certbot", Target: "/etc/letsencrypt"},
	}
}

func liveRestorePermissions(cfg app.Config) []backup.PermissionRule {
	return []backup.PermissionRule{
		{Path: "/etc/proidentity-mail", Owner: "proidentity", Group: "proidentity", DirMode: 0750, FileMode: 0640, Recursive: true},
		{Path: cfg.MailRoot, Owner: "vmail", Group: "vmail", DirMode: 0750, FileMode: 0640, Recursive: true},
		{Path: cfg.QuarantineDir, Owner: "proidentity", Group: "proidentity", DirMode: 0750, FileMode: 0640, Recursive: true},
		{Path: "/var/lib/rspamd/dkim", Owner: "_rspamd", Group: "_rspamd", DirMode: 0750, FileMode: 0640, Recursive: true},
		{Path: "/etc/nginx/proidentity", Owner: "root", Group: "root", DirMode: 0755, FileMode: 0644, Recursive: true},
		{Path: "/etc/nginx/conf.d/proidentity.conf", Owner: "root", Group: "root", FileMode: 0644},
	}
}

func liveRestoreServices() []string {
	return []string{"postfix", "dovecot", "rspamd", "nginx", "proidentity-webadmin", "proidentity-webmail", "proidentity-groupware"}
}

func backupSources(cfg app.Config) []backup.Source {
	return []backup.Source{
		{Name: "etc-proidentity-mail", Path: "/etc/proidentity-mail"},
		{Name: "maildir", Path: cfg.MailRoot},
		{Name: "quarantine", Path: cfg.QuarantineDir},
		{Name: "generated-config", Path: cfg.ConfigDir},
		{Name: "rspamd-dkim", Path: "/var/lib/rspamd/dkim"},
		{Name: "nginx-proidentity", Path: "/etc/nginx/proidentity"},
		{Name: "nginx-conf", Path: "/etc/nginx/conf.d/proidentity.conf"},
		{Name: "certbot", Path: "/etc/letsencrypt"},
	}
}

func writeProxyFiles(cfg app.Config, targetDir string) {
	if err := os.MkdirAll(targetDir, 0750); err != nil {
		log.Fatalf("create proxy render dir: %v", err)
	}
	mailSettings := loadProxyMailServerSettings(cfg)
	writeRendered(filepath.Join(targetDir, "proidentity-nginx.conf"), must(render.RenderNginxProxy(proxyRenderData(cfg, mailSettings))))
	writeRendered(filepath.Join(targetDir, "proxy-common.conf"), must(render.RenderNginxProxyCommon()))
	certPath := filepath.Join(targetDir, "issue-cert.sh")
	writeRendered(certPath, must(render.RenderCertbotScript(certbotRenderData(cfg, mailSettings))))
	if err := os.Chmod(certPath, 0750); err != nil {
		log.Fatalf("chmod cert script: %v", err)
	}
}

func proxyRenderData(cfg app.Config, mailSettings domain.MailServerSettings) render.NginxProxyData {
	return render.NginxProxyData{
		TLSMode:                 effectiveProxyTLSMode(cfg, mailSettings),
		AdminHostname:           cfg.AdminHostname,
		WebmailHostname:         cfg.WebmailHostname,
		DAVHostname:             cfg.DAVHostname,
		MailHostname:            cfg.MailHostname,
		AutoconfigHostname:      cfg.AutoconfigHostname,
		AutodiscoverHostname:    cfg.AutodiscoverHostname,
		ACMEWebroot:             cfg.ACMEWebroot,
		CertPath:                cfg.TLSCertPath,
		KeyPath:                 cfg.TLSKeyPath,
		ForceHTTPS:              effectiveProxyForceHTTPS(cfg, mailSettings),
		TrustProxyHeaders:       cfg.TrustProxyHeaders,
		TrustedProxyCIDRs:       cfg.TrustedProxyCIDRs,
		CloudflareRealIPEnabled: mailSettings.CloudflareRealIPEnabled,
	}
}

func effectiveProxyTLSMode(cfg app.Config, mailSettings domain.MailServerSettings) string {
	mode := strings.ToLower(strings.TrimSpace(mailSettings.TLSMode))
	if mode == "" || mode == "system" {
		return cfg.TLSMode
	}
	return mode
}

func effectiveProxyForceHTTPS(cfg app.Config, mailSettings domain.MailServerSettings) bool {
	if strings.TrimSpace(mailSettings.TLSMode) == "" {
		return cfg.ForceHTTPS
	}
	return mailSettings.ForceHTTPS
}

func loadProxyMailServerSettings(cfg app.Config) domain.MailServerSettings {
	if strings.TrimSpace(cfg.DBDSN) == "" {
		return domain.MailServerSettings{}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := db.Open(ctx, cfg.DBDSN)
	if err != nil {
		log.Printf("load proxy mail server settings skipped: %v", err)
		return domain.MailServerSettings{}
	}
	defer conn.Close()
	store := admin.NewSQLStore(conn).WithDNSSettings(admin.DNSSettings{
		MailHostname:    cfg.MailHostname,
		AdminHostname:   cfg.AdminHostname,
		WebmailHostname: cfg.WebmailHostname,
		PublicIPv4:      cfg.PublicIPv4,
		PublicIPv6:      cfg.PublicIPv6,
		TLSMode:         cfg.TLSMode,
		ForceHTTPS:      cfg.ForceHTTPS,
	})
	settings, err := store.GetMailServerSettings(ctx)
	if err != nil {
		log.Printf("load proxy mail server settings skipped: %v", err)
		return domain.MailServerSettings{}
	}
	return settings
}

func certbotRenderData(cfg app.Config, mailSettings domain.MailServerSettings) render.CertbotScriptData {
	return render.CertbotScriptData{
		TLSMode:                   effectiveProxyTLSMode(cfg, mailSettings),
		Hostnames:                 certHostnames(cfg),
		ACMEWebroot:               cfg.ACMEWebroot,
		CloudflareCredentialsFile: cfg.CloudflareCredentialsFile,
		CloudflareCertDomain:      cfg.CloudflareCertDomain,
		MailctlPath:               "/opt/proidentity-mail/bin/mailctl",
		CloudflarePropagationSec:  60,
	}
}

func certHostnames(cfg app.Config) []string {
	candidates := []string{cfg.AdminHostname, cfg.WebmailHostname, cfg.DAVHostname, cfg.MailHostname, cfg.AutoconfigHostname, cfg.AutodiscoverHostname}
	out := make([]string, 0, len(candidates))
	for _, host := range candidates {
		host = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(host, ".")))
		if host == "" || strings.HasSuffix(host, ".local") || !strings.Contains(host, ".") {
			continue
		}
		out = append(out, host)
	}
	return out
}

func dumpDatabase(cfg app.Config, outputPath string) error {
	if cfg.DBName == "" {
		return errors.New("PROIDENTITY_DB_NAME is required")
	}
	bin, err := exec.LookPath("mariadb-dump")
	if err != nil {
		bin, err = exec.LookPath("mysqldump")
		if err != nil {
			return err
		}
	}
	args := []string{"--single-transaction", "--routines", "--events"}
	if cfg.DBUser != "" {
		args = append(args, "-u", cfg.DBUser)
	}
	args = append(args, cfg.DBName)
	cmd := exec.Command(bin, args...)
	cmd.Env = os.Environ()
	if cfg.DBPassword != "" {
		cmd.Env = append(cmd.Env, "MYSQL_PWD="+cfg.DBPassword)
	}
	out, err := os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}
	defer out.Close()
	cmd.Stdout = out
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err == nil {
		return nil
	}
	if _, err := out.Seek(0, 0); err != nil {
		return err
	}
	if err := out.Truncate(0); err != nil {
		return err
	}
	fallback := exec.Command(bin, "--single-transaction", "--routines", "--events", cfg.DBName)
	fallback.Stdout = out
	fallback.Stderr = io.Discard
	return fallback.Run()
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

func writeSecretFile(path, data string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("output path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return err
	}
	temp := path + ".tmp"
	if err := os.WriteFile(temp, []byte(data), 0600); err != nil {
		return err
	}
	if err := os.Chmod(temp, 0600); err != nil {
		_ = os.Remove(temp)
		return err
	}
	if err := os.Rename(temp, path); err != nil {
		_ = os.Remove(temp)
		return err
	}
	return nil
}

func valueOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
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

func testNginx() error {
	path, err := exec.LookPath("nginx")
	if err != nil {
		return err
	}
	return exec.Command(path, "-t").Run()
}

func reloadNginx() error {
	if path, err := exec.LookPath("systemctl"); err == nil {
		return exec.Command(path, "reload-or-restart", "nginx").Run()
	}
	path, err := exec.LookPath("nginx")
	if err != nil {
		return err
	}
	return exec.Command(path, "-s", "reload").Run()
}
