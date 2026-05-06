package backup

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type commandCall struct {
	Name      string
	Args      []string
	Env       []string
	StdinPath string
}

type recordingRunner struct {
	calls []commandCall
}

func (r *recordingRunner) Run(_ context.Context, name string, args []string, options CommandOptions) error {
	r.calls = append(r.calls, commandCall{Name: name, Args: append([]string(nil), args...), Env: append([]string(nil), options.Env...), StdinPath: options.StdinPath})
	return nil
}

func TestRestoreLiveCopiesMappedSourcesImportsDatabaseAndRestarts(t *testing.T) {
	root := t.TempDir()
	sourceEtc := filepath.Join(root, "src", "etc")
	sourceMail := filepath.Join(root, "src", "mail")
	sourceDB := filepath.Join(root, "src", "db")
	if err := os.MkdirAll(sourceEtc, 0750); err != nil {
		t.Fatalf("mkdir etc: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(sourceMail, "example.com", "marko", "Maildir", "cur"), 0750); err != nil {
		t.Fatalf("mkdir mail: %v", err)
	}
	if err := os.MkdirAll(sourceDB, 0750); err != nil {
		t.Fatalf("mkdir db: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceEtc, "proidentity-mail.env"), []byte("PROIDENTITY=1\n"), 0640); err != nil {
		t.Fatalf("write env: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceMail, "example.com", "marko", "Maildir", "cur", "1.eml"), []byte("Subject: hi\r\n\r\nbody"), 0640); err != nil {
		t.Fatalf("write mail: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDB, "proidentity.sql"), []byte("CREATE TABLE restore_probe(id int);\n"), 0640); err != nil {
		t.Fatalf("write db: %v", err)
	}
	archive := filepath.Join(root, "backup.tar.gz")
	if _, err := Create(context.Background(), Options{OutputPath: archive, Sources: []Source{
		{Name: "etc-proidentity-mail", Path: sourceEtc, Required: true},
		{Name: "maildir", Path: sourceMail, Required: true},
		{Name: "database", Path: filepath.Join(sourceDB, "proidentity.sql"), Required: true},
	}}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	runner := &recordingRunner{}
	targetEtc := filepath.Join(root, "live", "etc", "proidentity-mail")
	targetMail := filepath.Join(root, "live", "var", "vmail")
	summary, err := RestoreLive(context.Background(), LiveRestoreOptions{
		ArchivePath:      archive,
		StagingDir:       filepath.Join(root, "stage"),
		Mappings:         []LiveMapping{{Source: "etc-proidentity-mail", Target: targetEtc}, {Source: "maildir", Target: targetMail}},
		Permissions:      []PermissionRule{{Path: targetEtc, Owner: "proidentity", Group: "proidentity", DirMode: 0750, FileMode: 0640, Recursive: true}},
		DatabaseName:     "proidentity_mail",
		DatabasePassword: "secret",
		ImportDatabase:   true,
		ControlServices:  true,
		Services:         []string{"postfix", "dovecot"},
		Runner:           runner,
		Overwrite:        true,
	})
	if err != nil {
		t.Fatalf("RestoreLive returned error: %v", err)
	}
	if summary.Files == 0 || summary.Bytes == 0 || !summary.DatabaseImported || !summary.PermissionsFixed {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if data, err := os.ReadFile(filepath.Join(targetEtc, "proidentity-mail.env")); err != nil || string(data) != "PROIDENTITY=1\n" {
		t.Fatalf("restored env mismatch data=%q err=%v", string(data), err)
	}
	if _, err := os.Stat(filepath.Join(targetMail, "example.com", "marko", "Maildir", "cur", "1.eml")); err != nil {
		t.Fatalf("restored mail missing: %v", err)
	}
	want := []string{"systemctl", "mariadb", "chown", "systemctl"}
	got := make([]string, 0, len(runner.calls))
	for _, call := range runner.calls {
		got = append(got, call.Name)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("command order = %v, want %v; calls=%+v", got, want, runner.calls)
	}
	if runner.calls[0].Args[0] != "stop" || runner.calls[3].Args[0] != "restart" {
		t.Fatalf("unexpected service actions: %+v", runner.calls)
	}
	if runner.calls[1].StdinPath == "" {
		t.Fatalf("database import did not use stdin path: %+v", runner.calls[1])
	}
	if !reflect.DeepEqual(runner.calls[1].Args, []string{"--database", "proidentity_mail"}) {
		t.Fatalf("database import args = %v", runner.calls[1].Args)
	}
	if !reflect.DeepEqual(runner.calls[1].Env, []string{"MYSQL_PWD=secret"}) {
		t.Fatalf("database import env = %v", runner.calls[1].Env)
	}
	if !reflect.DeepEqual(runner.calls[2].Args, []string{"-R", "proidentity:proidentity", targetEtc}) {
		t.Fatalf("permission command args = %v", runner.calls[2].Args)
	}
}

func TestRestoreLiveRequiresExplicitOverwriteForExistingTargets(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "src")
	if err := os.MkdirAll(source, 0750); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(source, "env"), []byte("new"), 0640); err != nil {
		t.Fatalf("write source: %v", err)
	}
	archive := filepath.Join(root, "backup.tar.gz")
	if _, err := Create(context.Background(), Options{OutputPath: archive, Sources: []Source{{Name: "etc-proidentity-mail", Path: source, Required: true}}}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	target := filepath.Join(root, "target")
	if err := os.MkdirAll(target, 0750); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "env"), []byte("old"), 0640); err != nil {
		t.Fatalf("write existing target: %v", err)
	}
	_, err := RestoreLive(context.Background(), LiveRestoreOptions{
		ArchivePath: archive,
		StagingDir:  filepath.Join(root, "stage"),
		Mappings:    []LiveMapping{{Source: "etc-proidentity-mail", Target: target}},
		Runner:      &recordingRunner{},
	})
	if err == nil {
		t.Fatal("expected restore to reject existing target without overwrite")
	}
}
