package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"proidentity-mail/internal/app"
	"proidentity-mail/internal/domain"
)

func TestCertHostnamesIncludesDiscoveryHostnames(t *testing.T) {
	got := certHostnames(app.Config{
		AdminHostname:        "madmin.example.com",
		WebmailHostname:      "webmail.example.com",
		DAVHostname:          "webmail.example.com",
		MailHostname:         "mail.example.com",
		AutoconfigHostname:   "autoconfig.example.com",
		AutodiscoverHostname: "autodiscover.example.com",
	})
	want := map[string]bool{
		"madmin.example.com":       false,
		"webmail.example.com":      false,
		"mail.example.com":         false,
		"autoconfig.example.com":   false,
		"autodiscover.example.com": false,
	}
	for _, host := range got {
		if _, ok := want[host]; ok {
			want[host] = true
		}
	}
	for host, found := range want {
		if !found {
			t.Fatalf("certHostnames missing %q in %+v", host, got)
		}
	}
}

func TestProxyRenderDataUsesCloudflareRealIPSetting(t *testing.T) {
	got := proxyRenderData(app.Config{}, domain.MailServerSettings{CloudflareRealIPEnabled: true})
	if !got.CloudflareRealIPEnabled {
		t.Fatalf("CloudflareRealIPEnabled = false, want true")
	}
}

func TestRunSyncProxyWritesLiveFilesReadableByDriftChecker(t *testing.T) {
	dir := t.TempDir()
	nginxConf := filepath.Join(dir, "nginx", "conf.d", "proidentity.conf")
	commonDir := filepath.Join(dir, "nginx", "proidentity")
	commonConf := filepath.Join(commonDir, "proxy-common.conf")
	certScript := filepath.Join(dir, "bin", "proidentity-issue-cert")
	if err := os.MkdirAll(filepath.Dir(certScript), 0755); err != nil {
		t.Fatal(err)
	}

	runSyncProxy(app.Config{
		ProxyMode:                 "internal-nginx",
		TLSMode:                   "behind-proxy",
		AdminHostname:             "madmin.example.test",
		WebmailHostname:           "webmail.example.test",
		DAVHostname:               "webmail.example.test",
		MailHostname:              "mail.example.test",
		AutoconfigHostname:        "autoconfig.example.test",
		AutodiscoverHostname:      "autodiscover.example.test",
		ACMEWebroot:               filepath.Join(dir, "acme"),
		TLSCertPath:               "/etc/ssl/certs/ssl-cert-snakeoil.pem",
		TLSKeyPath:                "/etc/ssl/private/ssl-cert-snakeoil.key",
		CloudflareCertDomain:      "example.test",
		CloudflareCredentialsFile: filepath.Join(dir, "cloudflare.ini"),
	}, []string{"--nginx-conf", nginxConf, "--common-dir", commonDir, "--cert-script", certScript})

	assertMode := func(path string, want os.FileMode) {
		t.Helper()
		if runtime.GOOS == "windows" {
			return
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != want {
			t.Fatalf("%s mode = %v, want %v", path, got, want)
		}
	}
	assertMode(nginxConf, 0644)
	assertMode(commonConf, 0644)
	assertMode(certScript, 0750)
}

func TestLiveRestoreConfirmationUsesArchiveFilename(t *testing.T) {
	archive := filepath.Join("backups", "20260512-130000.proidentity-backup.enc")
	want := "RESTORE 20260512-130000.proidentity-backup.enc"
	if got := liveRestoreConfirmationPhrase(archive); got != want {
		t.Fatalf("confirmation phrase = %q, want %q", got, want)
	}
	if err := validateLiveRestoreConfirmation(archive, want); err != nil {
		t.Fatalf("valid confirmation rejected: %v", err)
	}
	if err := validateLiveRestoreConfirmation(archive, "RESTORE other.enc"); err == nil {
		t.Fatal("wrong live restore confirmation was accepted")
	}
}

func TestBackupAuditActionsAlertOnlyForManualRuns(t *testing.T) {
	if got := strings.Join(backupAuditActions(true), ","); got != "backup.completed" {
		t.Fatalf("scheduled actions = %q", got)
	}
	if got := strings.Join(backupAuditActions(false), ","); got != "backup.completed,security.alert.backup_manual" {
		t.Fatalf("manual actions = %q", got)
	}
}

func TestRotateAdminPasswordInEnvFileUpdatesExistingPassword(t *testing.T) {
	path := filepath.Join(t.TempDir(), "webadmin.env")
	if err := os.WriteFile(path, []byte("PROIDENTITY_ADMIN_USERNAME=admin\nPROIDENTITY_ADMIN_PASSWORD=old\n"), 0600); err != nil {
		t.Fatalf("write env: %v", err)
	}

	if err := rotateAdminPasswordInEnvFile(path, "new-secret"); err != nil {
		t.Fatalf("rotate password: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read env: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "PROIDENTITY_ADMIN_PASSWORD=new-secret") {
		t.Fatalf("password was not updated: %s", text)
	}
	if strings.Contains(text, "PROIDENTITY_ADMIN_PASSWORD=old") {
		t.Fatalf("old password still present: %s", text)
	}
}

func TestRotateAdminPasswordInEnvFileAddsMissingPassword(t *testing.T) {
	path := filepath.Join(t.TempDir(), "webadmin.env")
	if err := os.WriteFile(path, []byte("PROIDENTITY_ADMIN_USERNAME=admin\n"), 0600); err != nil {
		t.Fatalf("write env: %v", err)
	}

	if err := rotateAdminPasswordInEnvFile(path, "generated-secret"); err != nil {
		t.Fatalf("rotate password: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read env: %v", err)
	}
	if !strings.Contains(string(data), "PROIDENTITY_ADMIN_PASSWORD=generated-secret") {
		t.Fatalf("password was not added: %s", string(data))
	}
}

func TestAdminPasswordEnvPathsHonorsExplicitEnvFile(t *testing.T) {
	t.Setenv("PROIDENTITY_ENV_FILE", "/tmp/custom-webadmin.env")

	paths := adminPasswordEnvPaths()
	if len(paths) != 1 || paths[0] != "/tmp/custom-webadmin.env" {
		t.Fatalf("paths = %#v", paths)
	}
}
