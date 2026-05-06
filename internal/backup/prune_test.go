package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPruneBackupsKeepsNewestDailyWeeklyMonthly(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC)
	dates := []time.Time{
		now.AddDate(0, 0, 0),
		now.AddDate(0, 0, -1),
		now.AddDate(0, 0, -2),
		now.AddDate(0, 0, -3),
		now.AddDate(0, 0, -4),
		now.AddDate(0, 0, -5),
		now.AddDate(0, 0, -6),
		now.AddDate(0, 0, -7),
		now.AddDate(0, 0, -14),
		now.AddDate(0, 0, -21),
		now.AddDate(0, 0, -28),
		now.AddDate(0, 0, -35),
		now.AddDate(0, -1, 0),
		now.AddDate(0, -2, 0),
		now.AddDate(0, -3, 0),
		now.AddDate(0, -4, 0),
		now.AddDate(0, -13, 0),
	}
	for _, date := range dates {
		name := "proidentity-mail-" + date.Format("20060102-150405") + ".tar.gz"
		path := filepath.Join(root, name)
		if err := os.WriteFile(path, []byte(name), 0640); err != nil {
			t.Fatalf("write backup: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "notes.txt"), []byte("keep me"), 0640); err != nil {
		t.Fatalf("write unrelated file: %v", err)
	}

	result, err := Prune(root, RetentionPolicy{Daily: 7, Weekly: 4, Monthly: 12}, PruneOptions{Now: now, Apply: true})
	if err != nil {
		t.Fatalf("Prune returned error: %v", err)
	}
	if result.Scanned != len(dates) {
		t.Fatalf("Scanned=%d, want %d", result.Scanned, len(dates))
	}
	if result.Deleted == 0 || result.Kept == 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if _, err := os.Stat(filepath.Join(root, "notes.txt")); err != nil {
		t.Fatalf("unrelated file removed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "proidentity-mail-"+dates[0].Format("20060102-150405")+".tar.gz")); err != nil {
		t.Fatalf("newest daily backup removed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "proidentity-mail-"+dates[len(dates)-1].Format("20060102-150405")+".tar.gz")); !os.IsNotExist(err) {
		t.Fatalf("old monthly backup was not pruned; err=%v", err)
	}
}

func TestPruneBackupsDryRunDoesNotDelete(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC)
	old := now.AddDate(0, -13, 0)
	path := filepath.Join(root, "proidentity-mail-"+old.Format("20060102-150405")+".tar.gz")
	if err := os.WriteFile(path, []byte("old"), 0640); err != nil {
		t.Fatalf("write backup: %v", err)
	}
	result, err := Prune(root, RetentionPolicy{Daily: 0, Weekly: 0, Monthly: 0}, PruneOptions{Now: now, Apply: false})
	if err != nil {
		t.Fatalf("Prune returned error: %v", err)
	}
	if result.WouldDelete != 1 || result.Deleted != 0 {
		t.Fatalf("unexpected dry-run result: %+v", result)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("dry-run deleted backup: %v", err)
	}
}
