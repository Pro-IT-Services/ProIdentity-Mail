package backup

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const backupTimeLayout = "20060102-150405"

type RetentionPolicy struct {
	Daily   int
	Weekly  int
	Monthly int
}

type PruneOptions struct {
	Now   time.Time
	Apply bool
}

type PruneResult struct {
	Scanned     int
	Kept        int
	WouldDelete int
	Deleted     int
}

type backupCandidate struct {
	Path string
	Time time.Time
}

func Prune(dir string, policy RetentionPolicy, options PruneOptions) (PruneResult, error) {
	if strings.TrimSpace(dir) == "" {
		return PruneResult{}, errors.New("backup directory is required")
	}
	if options.Now.IsZero() {
		options.Now = time.Now().UTC()
	}
	candidates, err := listBackupCandidates(dir)
	if err != nil {
		return PruneResult{}, err
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Time.After(candidates[j].Time)
	})
	keep := selectBackupsToKeep(candidates, policy, options.Now)
	result := PruneResult{Scanned: len(candidates)}
	for _, candidate := range candidates {
		if keep[candidate.Path] {
			result.Kept++
			continue
		}
		result.WouldDelete++
		if options.Apply {
			if err := os.Remove(candidate.Path); err != nil {
				return result, err
			}
			result.Deleted++
		}
	}
	return result, nil
}

func listBackupCandidates(dir string) ([]backupCandidate, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	candidates := make([]backupCandidate, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		when, ok := parseBackupName(entry.Name())
		if !ok {
			continue
		}
		candidates = append(candidates, backupCandidate{Path: filepath.Join(dir, entry.Name()), Time: when})
	}
	return candidates, nil
}

func parseBackupName(name string) (time.Time, bool) {
	if !strings.HasPrefix(name, "proidentity-mail-") || !strings.HasSuffix(name, ".tar.gz") {
		return time.Time{}, false
	}
	stamp := strings.TrimSuffix(strings.TrimPrefix(name, "proidentity-mail-"), ".tar.gz")
	when, err := time.ParseInLocation(backupTimeLayout, stamp, time.UTC)
	return when, err == nil
}

func selectBackupsToKeep(candidates []backupCandidate, policy RetentionPolicy, now time.Time) map[string]bool {
	keep := make(map[string]bool)
	daily := make(map[string]bool)
	weekly := make(map[string]bool)
	monthly := make(map[string]bool)
	dailyCutoff := now.AddDate(0, 0, -policy.Daily+1)
	weeklyCutoff := now.AddDate(0, 0, -7*policy.Weekly+1)
	monthlyCutoff := now.AddDate(0, -policy.Monthly+1, 0)
	for _, candidate := range candidates {
		if policy.Daily > 0 && !candidate.Time.Before(startOfDay(dailyCutoff)) {
			key := candidate.Time.Format("2006-01-02")
			if !daily[key] {
				daily[key] = true
				keep[candidate.Path] = true
				if len(daily) >= policy.Daily {
					policy.Daily = 0
				}
			}
		}
		if policy.Weekly > 0 && !candidate.Time.Before(startOfDay(weeklyCutoff)) {
			year, week := candidate.Time.ISOWeek()
			key := strconv.Itoa(year) + "-" + twoDigits(week)
			if !weekly[key] {
				weekly[key] = true
				keep[candidate.Path] = true
				if len(weekly) >= policy.Weekly {
					policy.Weekly = 0
				}
			}
		}
		if policy.Monthly > 0 && !candidate.Time.Before(startOfMonth(monthlyCutoff)) {
			key := candidate.Time.Format("2006-01")
			if !monthly[key] {
				monthly[key] = true
				keep[candidate.Path] = true
				if len(monthly) >= policy.Monthly {
					policy.Monthly = 0
				}
			}
		}
	}
	return keep
}

func startOfDay(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func startOfMonth(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func twoDigits(value int) string {
	if value < 10 {
		return "0" + strconv.Itoa(value)
	}
	return strconv.Itoa(value)
}
