package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateAndVerifyArchiveWithManifestAndHashes(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "mail", "example.com"), 0750); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "mail", "example.com", "message.eml"), []byte("Subject: test\r\n\r\nbody"), 0640); err != nil {
		t.Fatalf("write source: %v", err)
	}
	out := filepath.Join(t.TempDir(), "backup.tar.gz")
	manifest, err := Create(context.Background(), Options{
		OutputPath: out,
		Sources: []Source{{
			Name: "maildir",
			Path: filepath.Join(root, "mail"),
		}},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if manifest.Version != 1 || len(manifest.Entries) < 1 {
		t.Fatalf("unexpected manifest: %+v", manifest)
	}
	summary, err := Verify(context.Background(), out)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if summary.Files != 1 || summary.Bytes == 0 {
		t.Fatalf("unexpected verify summary: %+v", summary)
	}
}

func TestCreateVerifyAndRestoreEncryptedArchive(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "secret.txt"), []byte("sensitive mail and keys"), 0640); err != nil {
		t.Fatalf("write source: %v", err)
	}
	key := []byte("0123456789abcdef0123456789abcdef")
	out := filepath.Join(t.TempDir(), "backup.tar.gz.enc")
	if _, err := Create(context.Background(), Options{OutputPath: out, Sources: []Source{{Name: "config", Path: root}}, EncryptionKey: key}); err != nil {
		t.Fatalf("Create encrypted returned error: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read encrypted archive: %v", err)
	}
	if strings.Contains(string(data), "sensitive mail and keys") {
		t.Fatalf("encrypted archive leaked plaintext payload")
	}
	if _, err := Verify(context.Background(), out); err == nil {
		t.Fatal("Verify without key accepted encrypted archive")
	}
	if _, err := VerifyWithKey(context.Background(), out, key); err != nil {
		t.Fatalf("VerifyWithKey returned error: %v", err)
	}
	target := t.TempDir()
	if err := RestoreWithKey(context.Background(), out, target, RestoreOptions{}, key); err != nil {
		t.Fatalf("RestoreWithKey returned error: %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(target, "config", "secret.txt")); err != nil || string(got) != "sensitive mail and keys" {
		t.Fatalf("restored encrypted data mismatch got=%q err=%v", string(got), err)
	}
}

func writeTestArchive(path string, files map[string]string) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	gz := gzip.NewWriter(out)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()
	for name, body := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0640, Size: int64(len(body))}); err != nil {
			return err
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			return err
		}
	}
	return nil
}

func TestVerifyRejectsTamperedArchive(t *testing.T) {
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "secret.txt"), []byte("before"), 0640); err != nil {
		t.Fatalf("write source: %v", err)
	}
	out := filepath.Join(t.TempDir(), "backup.tar.gz")
	if _, err := Create(context.Background(), Options{OutputPath: out, Sources: []Source{{Name: "config", Path: source}}}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read archive: %v", err)
	}
	replaced := strings.Replace(string(data), "before", "after!", 1)
	if replaced == string(data) {
		t.Skip("gzip data did not contain literal test payload")
	}
	if err := os.WriteFile(out, []byte(replaced), 0640); err != nil {
		t.Fatalf("tamper archive: %v", err)
	}
	if _, err := Verify(context.Background(), out); err == nil {
		t.Fatal("expected tampered archive verification to fail")
	}
}

func TestRestoreRejectsUnsafeArchivePath(t *testing.T) {
	out := filepath.Join(t.TempDir(), "unsafe.tar.gz")
	if err := writeTestArchive(out, map[string]string{
		"proidentity-backup-manifest.json": `{"version":1,"entries":[{"path":"../escape","type":"file","size":1,"sha256":"x"}]}`,
		"../escape":                        "x",
	}); err != nil {
		t.Fatalf("write test archive: %v", err)
	}
	if err := Restore(context.Background(), out, t.TempDir(), RestoreOptions{}); err == nil {
		t.Fatal("expected unsafe restore path to fail")
	}
}

func TestRestoreExtractsVerifiedArchive(t *testing.T) {
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "env"), []byte("PROIDENTITY=1\n"), 0640); err != nil {
		t.Fatalf("write source: %v", err)
	}
	out := filepath.Join(t.TempDir(), "backup.tar.gz")
	if _, err := Create(context.Background(), Options{OutputPath: out, Sources: []Source{{Name: "config", Path: source}}}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	target := t.TempDir()
	if err := Restore(context.Background(), out, target, RestoreOptions{}); err != nil {
		t.Fatalf("Restore returned error: %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(target, "config", "env")); err != nil || string(data) != "PROIDENTITY=1\n" {
		t.Fatalf("restored file mismatch data=%q err=%v", string(data), err)
	}
}
