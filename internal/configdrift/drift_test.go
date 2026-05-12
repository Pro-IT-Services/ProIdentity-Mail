package configdrift

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompareReportsMatchAndDriftWithRedactedDiff(t *testing.T) {
	dir := t.TempDir()
	desired := filepath.Join(dir, "desired.cf")
	live := filepath.Join(dir, "live.cf")
	if err := os.WriteFile(desired, []byte("myhostname = mail.example\npassword = desired-secret\n"), 0640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(live, []byte("myhostname = old.example\npassword = live-secret\n"), 0640); err != nil {
		t.Fatal(err)
	}

	report := Compare(context.Background(), []Mapping{{ID: "postfix-main", Service: "postfix", Label: "Postfix main", DesiredPath: desired, LivePath: live}})
	if report.Status != "drift" || report.Summary.Drifted != 1 {
		t.Fatalf("expected drift report, got %+v", report)
	}
	diff := report.Items[0].Diff
	for _, want := range []string{"-myhostname = old.example", "+myhostname = mail.example", "password = <redacted>"} {
		if !strings.Contains(diff, want) {
			t.Fatalf("diff missing %q:\n%s", want, diff)
		}
	}
	if strings.Contains(diff, "desired-secret") || strings.Contains(diff, "live-secret") {
		t.Fatalf("diff leaked secret:\n%s", diff)
	}
}

func TestCompareReportsMissingLiveFile(t *testing.T) {
	dir := t.TempDir()
	desired := filepath.Join(dir, "desired.cf")
	if err := os.WriteFile(desired, []byte("content\n"), 0640); err != nil {
		t.Fatal(err)
	}
	report := Compare(context.Background(), []Mapping{{ID: "missing", Service: "postfix", Label: "Missing", DesiredPath: desired, LivePath: filepath.Join(dir, "missing.cf")}})
	if report.Status != "drift" || report.Summary.MissingLive != 1 || report.Items[0].Status != "missing_live" {
		t.Fatalf("expected missing live drift, got %+v", report)
	}
}

func TestDefaultMappingsUseDesiredDirsAndLiveRoot(t *testing.T) {
	liveRoot := t.TempDir()
	mappings := DefaultMappings(filepath.Join("tmp", "mail"), filepath.Join("tmp", "proxy"), liveRoot)

	assertMapping := func(id, desired, live string) {
		t.Helper()
		for _, mapping := range mappings {
			if mapping.ID == id {
				if mapping.DesiredPath != desired {
					t.Fatalf("%s desired path = %q, want %q", id, mapping.DesiredPath, desired)
				}
				if mapping.LivePath != live {
					t.Fatalf("%s live path = %q, want %q", id, mapping.LivePath, live)
				}
				return
			}
		}
		t.Fatalf("mapping %q not found in %+v", id, mappings)
	}

	assertMapping("postfix-main", filepath.Join("tmp", "mail", "postfix-main.cf"), filepath.Join(liveRoot, "etc", "postfix", "main.cf"))
	assertMapping("nginx-proxy", filepath.Join("tmp", "proxy", "proidentity-nginx.conf"), filepath.Join(liveRoot, "etc", "nginx", "conf.d", "proidentity.conf"))
	assertMapping("cert-helper", filepath.Join("tmp", "proxy", "issue-cert.sh"), filepath.Join(liveRoot, "opt", "proidentity-mail", "bin", "proidentity-issue-cert"))
}
