package main

import (
	"context"
	"database/sql"
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
		fmt.Fprintln(os.Stderr, "usage: mailctl migrate|render|health|seed-dev")
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

func runHealth() {
	ctx := context.Background()
	results := []health.CheckResult{
		health.TCP(ctx, "smtp", "127.0.0.1:25"),
		health.TCP(ctx, "submission", "127.0.0.1:587"),
		health.TCP(ctx, "imap", "127.0.0.1:143"),
		health.TCP(ctx, "pop3", "127.0.0.1:110"),
		health.TCP(ctx, "rspamd", "127.0.0.1:11334"),
		health.TCP(ctx, "redis", "127.0.0.1:6379"),
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
