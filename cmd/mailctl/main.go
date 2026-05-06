package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/textproto"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"proidentity-mail/internal/admin"
	"proidentity-mail/internal/app"
	"proidentity-mail/internal/backup"
	"proidentity-mail/internal/db"
	"proidentity-mail/internal/health"
	"proidentity-mail/internal/quarantine"
	"proidentity-mail/internal/render"
)

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: mailctl migrate|render|health|seed-dev|rotate-admin-password|quarantine-message|release-quarantine|sync-rspamd-policy|render-proxy|sync-proxy|backup|backup-prune|backup-verify|restore")
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
	if err := os.MkdirAll(cfg.ACMEWebroot, 0755); err != nil {
		log.Fatalf("create acme webroot: %v", err)
	}
	writeRenderedAtomic(*nginxConf, must(render.RenderNginxProxy(proxyRenderData(cfg))))
	writeRenderedAtomic(filepath.Join(*commonDir, "proxy-common.conf"), must(render.RenderNginxProxyCommon()))
	writeRenderedAtomic(*certScript, must(render.RenderCertbotScript(certbotRenderData(cfg))))
	if err := os.Chmod(*certScript, 0750); err != nil {
		log.Fatalf("chmod cert script: %v", err)
	}
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

func runBackup(cfg app.Config, args []string) {
	flags := flag.NewFlagSet("backup", flag.ExitOnError)
	outputDir := flags.String("output-dir", "/var/backups/proidentity-mail", "backup output directory")
	output := flags.String("output", "", "exact backup archive path")
	includeDB := flags.Bool("include-db", true, "include MariaDB logical dump")
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
		outputPath = filepath.Join(*outputDir, "proidentity-mail-"+time.Now().UTC().Format("20060102-150405")+".tar.gz")
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
	manifest, err := backup.Create(context.Background(), backup.Options{OutputPath: outputPath, Sources: sources, Hostname: host})
	if err != nil {
		log.Fatalf("backup create failed: %v", err)
	}
	summary, err := backup.Verify(context.Background(), outputPath)
	if err != nil {
		log.Fatalf("backup verify failed: %v", err)
	}
	fmt.Printf("backup=%s entries=%d files=%d bytes=%d\n", outputPath, len(manifest.Entries), summary.Files, summary.Bytes)
	if *pruneAfter {
		result, err := backup.Prune(*outputDir, backup.RetentionPolicy{Daily: *keepDaily, Weekly: *keepWeekly, Monthly: *keepMonthly}, backup.PruneOptions{Apply: true})
		if err != nil {
			log.Fatalf("backup prune failed: %v", err)
		}
		fmt.Printf("backup_prune dir=%s scanned=%d kept=%d deleted=%d\n", *outputDir, result.Scanned, result.Kept, result.Deleted)
	}
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
	summary, err := backup.Verify(context.Background(), *archive)
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
	if err := flags.Parse(args); err != nil {
		log.Fatalf("parse restore flags: %v", err)
	}
	if *archive == "" {
		log.Fatal("-archive is required")
	}
	summary, err := backup.Verify(context.Background(), *archive)
	if err != nil {
		log.Fatalf("restore verification failed: %v", err)
	}
	if !*apply {
		fmt.Printf("restore_dry_run archive=%s entries=%d files=%d bytes=%d\n", *archive, summary.Entries, summary.Files, summary.Bytes)
		return
	}
	if *live {
		options := backup.LiveRestoreOptions{
			ArchivePath:      *archive,
			StagingDir:       *stageDir,
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
	if err := backup.Restore(context.Background(), *archive, *targetRoot, backup.RestoreOptions{Overwrite: *overwrite}); err != nil {
		log.Fatalf("restore failed: %v", err)
	}
	fmt.Printf("restore_extracted archive=%s target=%s files=%d bytes=%d\n", *archive, *targetRoot, summary.Files, summary.Bytes)
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
	writeRendered(filepath.Join(targetDir, "proidentity-nginx.conf"), must(render.RenderNginxProxy(proxyRenderData(cfg))))
	writeRendered(filepath.Join(targetDir, "proxy-common.conf"), must(render.RenderNginxProxyCommon()))
	certPath := filepath.Join(targetDir, "issue-cert.sh")
	writeRendered(certPath, must(render.RenderCertbotScript(certbotRenderData(cfg))))
	if err := os.Chmod(certPath, 0750); err != nil {
		log.Fatalf("chmod cert script: %v", err)
	}
}

func proxyRenderData(cfg app.Config) render.NginxProxyData {
	return render.NginxProxyData{
		TLSMode:           cfg.TLSMode,
		AdminHostname:     cfg.AdminHostname,
		WebmailHostname:   cfg.WebmailHostname,
		DAVHostname:       cfg.DAVHostname,
		ACMEWebroot:       cfg.ACMEWebroot,
		CertPath:          cfg.TLSCertPath,
		KeyPath:           cfg.TLSKeyPath,
		ForceHTTPS:        cfg.ForceHTTPS,
		TrustProxyHeaders: cfg.TrustProxyHeaders,
		TrustedProxyCIDRs: cfg.TrustedProxyCIDRs,
	}
}

func certbotRenderData(cfg app.Config) render.CertbotScriptData {
	return render.CertbotScriptData{
		TLSMode:                   cfg.TLSMode,
		Hostnames:                 []string{cfg.AdminHostname, cfg.WebmailHostname, cfg.DAVHostname},
		ACMEWebroot:               cfg.ACMEWebroot,
		CloudflareCredentialsFile: cfg.CloudflareCredentialsFile,
		CloudflarePropagationSec:  60,
	}
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
