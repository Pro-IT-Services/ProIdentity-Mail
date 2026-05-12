package webmail

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMaildirStoreSyncFiltersWritesActiveSieveScript(t *testing.T) {
	root := t.TempDir()
	store := MaildirStore{Root: root}

	err := store.SyncFilters(context.Background(), "tester@proidentity.cloud", []MailFilter{
		{Name: "Gmail sender", Field: "from", Operator: "contains", Value: "marcel.panac.ipecon@gmail.com", Action: "move", Folder: "GMAIL", Enabled: true},
		{Name: "Dangerous chars", Field: "subject", Operator: "starts_with", Value: `quote"slash\star*`, Action: "delete", Enabled: true},
		{Name: "Off", Field: "subject", Operator: "contains", Value: "ignored", Action: "move", Folder: "Ignored", Enabled: false},
	})
	if err != nil {
		t.Fatalf("SyncFilters returned error: %v", err)
	}

	home := filepath.Join(root, "proidentity.cloud", "tester")
	scriptPath := filepath.Join(home, "sieve", "proidentity.sieve")
	activePath := filepath.Join(home, ".dovecot.sieve")
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read generated sieve script: %v", err)
	}
	text := string(data)
	for _, want := range []string{
		`require ["fileinto", "body"];`,
		`# Gmail sender`,
		`if header :contains "From" "marcel.panac.ipecon@gmail.com" {`,
		`fileinto "GMAIL";`,
		`stop;`,
		`if header :matches "Subject" "quote\"slash\\star**" {`,
		`discard;`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("generated sieve script missing %q in:\n%s", want, text)
		}
	}
	if strings.Contains(text, "ignored") {
		t.Fatalf("disabled filter should not be rendered:\n%s", text)
	}
	if runtime.GOOS == "windows" {
		active, err := os.ReadFile(activePath)
		if err != nil {
			t.Fatalf("read active sieve script: %v", err)
		}
		if string(active) != text {
			t.Fatalf("active sieve script should match generated script")
		}
	} else {
		target, err := os.Readlink(activePath)
		if err != nil {
			t.Fatalf("active sieve script should be a symlink: %v", err)
		}
		if target != filepath.ToSlash(filepath.Join("sieve", "proidentity.sieve")) && target != filepath.Join("sieve", "proidentity.sieve") {
			t.Fatalf("unexpected active sieve symlink target %q", target)
		}
	}
}

func TestCompositeStoreSyncsSieveAfterFilterChanges(t *testing.T) {
	root := t.TempDir()
	auth := &fakeFilterAuth{
		filters: []MailFilter{{ID: "1", Name: "Gmail", Field: "from", Operator: "contains", Value: "gmail.com", Action: "move", Folder: "GMAIL", Enabled: true}},
	}
	store := CompositeStore{Auth: auth, Mailbox: MaildirStore{Root: root}}

	if _, err := store.CreateFilter(context.Background(), "tester@proidentity.cloud", auth.filters[0]); err != nil {
		t.Fatalf("CreateFilter returned error: %v", err)
	}
	assertSieveContains(t, root, "tester@proidentity.cloud", `fileinto "GMAIL";`)

	auth.filters = []MailFilter{{ID: "1", Name: "Delete Gmail", Field: "from", Operator: "contains", Value: "gmail.com", Action: "delete", Enabled: true}}
	if _, err := store.UpdateFilter(context.Background(), "tester@proidentity.cloud", "1", auth.filters[0]); err != nil {
		t.Fatalf("UpdateFilter returned error: %v", err)
	}
	assertSieveContains(t, root, "tester@proidentity.cloud", "discard;")

	auth.filters = nil
	if err := store.DeleteFilter(context.Background(), "tester@proidentity.cloud", "1"); err != nil {
		t.Fatalf("DeleteFilter returned error: %v", err)
	}
	assertSieveContains(t, root, "tester@proidentity.cloud", "# No enabled ProIdentity filters.")
}

func assertSieveContains(t *testing.T, root, email, want string) {
	t.Helper()
	local, domain, _ := strings.Cut(email, "@")
	data, err := os.ReadFile(filepath.Join(root, domain, local, "sieve", "proidentity.sieve"))
	if err != nil {
		t.Fatalf("read generated sieve script: %v", err)
	}
	if !strings.Contains(string(data), want) {
		t.Fatalf("generated sieve script missing %q in:\n%s", want, string(data))
	}
}

type fakeFilterAuth struct {
	filters []MailFilter
}

func (f *fakeFilterAuth) VerifyUserPassword(context.Context, string, string) (bool, error) {
	return true, nil
}

func (f *fakeFilterAuth) ReportMessage(context.Context, string, string, string) error {
	return nil
}

func (f *fakeFilterAuth) ListFilters(context.Context, string) ([]MailFilter, error) {
	return append([]MailFilter(nil), f.filters...), nil
}

func (f *fakeFilterAuth) CreateFilter(context.Context, string, MailFilter) (MailFilter, error) {
	if len(f.filters) == 0 {
		return MailFilter{}, sql.ErrNoRows
	}
	return f.filters[0], nil
}

func (f *fakeFilterAuth) UpdateFilter(context.Context, string, string, MailFilter) (MailFilter, error) {
	if len(f.filters) == 0 {
		return MailFilter{}, sql.ErrNoRows
	}
	return f.filters[0], nil
}

func (f *fakeFilterAuth) DeleteFilter(context.Context, string, string) error {
	return nil
}
