package configdrift

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Mapping struct {
	ID          string `json:"id"`
	Service     string `json:"service"`
	Label       string `json:"label"`
	DesiredPath string `json:"desired_path"`
	LivePath    string `json:"live_path"`
}

type Report struct {
	Status    string    `json:"status"`
	CheckedAt time.Time `json:"checked_at"`
	Summary   Summary   `json:"summary"`
	Items     []Item    `json:"items"`
}

type Summary struct {
	Total       int `json:"total"`
	Matching    int `json:"matching"`
	Drifted     int `json:"drifted"`
	MissingLive int `json:"missing_live"`
	Errors      int `json:"errors"`
}

type Item struct {
	ID            string `json:"id"`
	Service       string `json:"service"`
	Label         string `json:"label"`
	Status        string `json:"status"`
	DesiredPath   string `json:"desired_path"`
	LivePath      string `json:"live_path"`
	DesiredSHA256 string `json:"desired_sha256,omitempty"`
	LiveSHA256    string `json:"live_sha256,omitempty"`
	DesiredSize   int    `json:"desired_size,omitempty"`
	LiveSize      int    `json:"live_size,omitempty"`
	Diff          string `json:"diff,omitempty"`
	Error         string `json:"error,omitempty"`
}

func DefaultMappings(desiredMailDir, desiredProxyDir, liveRoot string) []Mapping {
	return []Mapping{
		{ID: "postfix-main", Service: "postfix", Label: "Postfix main.cf", DesiredPath: filepath.Join(desiredMailDir, "postfix-main.cf"), LivePath: livePath(liveRoot, "/etc/postfix/main.cf")},
		{ID: "postfix-master", Service: "postfix", Label: "Postfix master.cf", DesiredPath: filepath.Join(desiredMailDir, "postfix-master.cf"), LivePath: livePath(liveRoot, "/etc/postfix/master.cf")},
		{ID: "virtual-mailbox-domains", Service: "postfix", Label: "Virtual mailbox domains", DesiredPath: filepath.Join(desiredMailDir, "virtual-mailbox-domains.cf"), LivePath: livePath(liveRoot, "/etc/postfix/proidentity/virtual-mailbox-domains.cf")},
		{ID: "virtual-mailbox-maps", Service: "postfix", Label: "Virtual mailbox maps", DesiredPath: filepath.Join(desiredMailDir, "virtual-mailbox-maps.cf"), LivePath: livePath(liveRoot, "/etc/postfix/proidentity/virtual-mailbox-maps.cf")},
		{ID: "virtual-alias-maps", Service: "postfix", Label: "Virtual alias maps", DesiredPath: filepath.Join(desiredMailDir, "virtual-alias-maps.cf"), LivePath: livePath(liveRoot, "/etc/postfix/proidentity/virtual-alias-maps.cf")},
		{ID: "sender-login-maps", Service: "postfix", Label: "Sender login maps", DesiredPath: filepath.Join(desiredMailDir, "sender-login-maps.cf"), LivePath: livePath(liveRoot, "/etc/postfix/proidentity/sender-login-maps.cf")},
		{ID: "postfix-tls-sni-map", Service: "postfix", Label: "Postfix TLS SNI map", DesiredPath: filepath.Join(desiredMailDir, "postfix-tls-sni-map"), LivePath: livePath(liveRoot, "/etc/postfix/proidentity/tls-sni-map")},
		{ID: "dovecot-local", Service: "dovecot", Label: "Dovecot local configuration", DesiredPath: filepath.Join(desiredMailDir, "dovecot-proidentity.conf"), LivePath: livePath(liveRoot, "/etc/dovecot/conf.d/99-proidentity.conf")},
		{ID: "dovecot-sql", Service: "dovecot", Label: "Dovecot SQL auth", DesiredPath: filepath.Join(desiredMailDir, "dovecot-sql.conf.ext"), LivePath: livePath(liveRoot, "/etc/dovecot/proidentity-sql.conf.ext")},
		{ID: "rspamd-redis", Service: "rspamd", Label: "Rspamd Redis", DesiredPath: filepath.Join(desiredMailDir, "rspamd-local.d-redis.conf"), LivePath: livePath(liveRoot, "/etc/rspamd/local.d/redis.conf")},
		{ID: "rspamd-antivirus", Service: "rspamd", Label: "Rspamd antivirus", DesiredPath: filepath.Join(desiredMailDir, "rspamd-local.d-antivirus.conf"), LivePath: livePath(liveRoot, "/etc/rspamd/local.d/antivirus.conf")},
		{ID: "rspamd-dkim-signing", Service: "rspamd", Label: "Rspamd DKIM signing", DesiredPath: filepath.Join(desiredMailDir, "rspamd-local.d-dkim_signing.conf"), LivePath: livePath(liveRoot, "/etc/rspamd/local.d/dkim_signing.conf")},
		{ID: "rspamd-actions", Service: "rspamd", Label: "Rspamd actions", DesiredPath: filepath.Join(desiredMailDir, "rspamd-local.d-actions.conf"), LivePath: livePath(liveRoot, "/etc/rspamd/local.d/actions.conf")},
		{ID: "rspamd-milter-headers", Service: "rspamd", Label: "Rspamd milter headers", DesiredPath: filepath.Join(desiredMailDir, "rspamd-local.d-milter_headers.conf"), LivePath: livePath(liveRoot, "/etc/rspamd/local.d/milter_headers.conf")},
		{ID: "rspamd-tenant-settings", Service: "rspamd", Label: "Rspamd tenant settings", DesiredPath: filepath.Join(desiredMailDir, "rspamd-local.d-settings.conf"), LivePath: livePath(liveRoot, "/etc/rspamd/local.d/settings.conf")},
		{ID: "rspamd-force-actions", Service: "rspamd", Label: "Rspamd force actions", DesiredPath: filepath.Join(desiredMailDir, "rspamd-local.d-force_actions.conf"), LivePath: livePath(liveRoot, "/etc/rspamd/local.d/force_actions.conf")},
		{ID: "nginx-proxy", Service: "nginx", Label: "Nginx virtual hosts", DesiredPath: filepath.Join(desiredProxyDir, "proidentity-nginx.conf"), LivePath: livePath(liveRoot, "/etc/nginx/conf.d/proidentity.conf")},
		{ID: "nginx-common", Service: "nginx", Label: "Nginx shared proxy include", DesiredPath: filepath.Join(desiredProxyDir, "proxy-common.conf"), LivePath: livePath(liveRoot, "/etc/nginx/proidentity/proxy-common.conf")},
		{ID: "cert-helper", Service: "nginx", Label: "Certificate helper script", DesiredPath: filepath.Join(desiredProxyDir, "issue-cert.sh"), LivePath: livePath(liveRoot, "/opt/proidentity-mail/bin/proidentity-issue-cert")},
	}
}

func livePath(root, absolutePath string) string {
	if strings.TrimSpace(root) == "" {
		return absolutePath
	}
	return filepath.Join(root, strings.TrimLeft(absolutePath, `/\`))
}

func Compare(ctx context.Context, mappings []Mapping) Report {
	report := Report{Status: "ok", CheckedAt: time.Now().UTC(), Items: make([]Item, 0, len(mappings))}
	for _, mapping := range mappings {
		select {
		case <-ctx.Done():
			report.Items = append(report.Items, Item{
				ID:          mapping.ID,
				Service:     mapping.Service,
				Label:       mapping.Label,
				Status:      "error",
				DesiredPath: mapping.DesiredPath,
				LivePath:    mapping.LivePath,
				Error:       ctx.Err().Error(),
			})
			report.Summary.Errors++
			continue
		default:
		}
		item := compareOne(mapping)
		report.Items = append(report.Items, item)
		switch item.Status {
		case "match":
			report.Summary.Matching++
		case "drift":
			report.Summary.Drifted++
		case "missing_live":
			report.Summary.MissingLive++
		default:
			report.Summary.Errors++
		}
	}
	report.Summary.Total = len(report.Items)
	if report.Summary.Errors > 0 {
		report.Status = "error"
	} else if report.Summary.Drifted > 0 || report.Summary.MissingLive > 0 {
		report.Status = "drift"
	}
	return report
}

func compareOne(mapping Mapping) Item {
	item := Item{ID: mapping.ID, Service: mapping.Service, Label: mapping.Label, DesiredPath: mapping.DesiredPath, LivePath: mapping.LivePath}
	desired, err := os.ReadFile(mapping.DesiredPath)
	if err != nil {
		item.Status = "error"
		item.Error = "read desired: " + err.Error()
		return item
	}
	live, err := os.ReadFile(mapping.LivePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			item.Status = "missing_live"
			item.Error = "live file missing"
			item.DesiredSHA256 = sha256Hex(desired)
			item.DesiredSize = len(desired)
			item.Diff = unifiedDiff(mapping.DesiredPath, mapping.LivePath, desired, nil)
			return item
		}
		item.Status = "error"
		item.Error = "read live: " + err.Error()
		return item
	}
	item.DesiredSHA256 = sha256Hex(desired)
	item.LiveSHA256 = sha256Hex(live)
	item.DesiredSize = len(desired)
	item.LiveSize = len(live)
	if bytes.Equal(desired, live) {
		item.Status = "match"
		return item
	}
	item.Status = "drift"
	item.Diff = unifiedDiff(mapping.DesiredPath, mapping.LivePath, desired, live)
	return item
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func unifiedDiff(desiredPath, livePath string, desired, live []byte) string {
	desiredText := redactConfig(string(desired))
	liveText := redactConfig(string(live))
	var out strings.Builder
	out.WriteString("--- live: ")
	out.WriteString(livePath)
	out.WriteByte('\n')
	out.WriteString("+++ desired: ")
	out.WriteString(desiredPath)
	out.WriteByte('\n')
	liveLines := splitLines(liveText)
	desiredLines := splitLines(desiredText)
	if len(liveLines)*len(desiredLines) > 1000000 {
		out.WriteString("@@ full diff omitted because files are large @@\n")
		out.WriteString("- live sha/content differs\n")
		out.WriteString("+ desired sha/content differs\n")
		return out.String()
	}
	for _, line := range lineDiff(liveLines, desiredLines) {
		out.WriteString(line)
		out.WriteByte('\n')
	}
	return strings.TrimRight(out.String(), "\n")
}

func splitLines(value string) []string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.TrimSuffix(value, "\n")
	if value == "" {
		return nil
	}
	return strings.Split(value, "\n")
}

func lineDiff(a, b []string) []string {
	dp := make([][]int, len(a)+1)
	for i := range dp {
		dp[i] = make([]int, len(b)+1)
	}
	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i] == b[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}
	lines := []string{"@@"}
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			lines = append(lines, " "+a[i])
			i++
			j++
		} else if dp[i+1][j] >= dp[i][j+1] {
			lines = append(lines, "-"+a[i])
			i++
		} else {
			lines = append(lines, "+"+b[j])
			j++
		}
	}
	for ; i < len(a); i++ {
		lines = append(lines, "-"+a[i])
	}
	for ; j < len(b); j++ {
		lines = append(lines, "+"+b[j])
	}
	return lines
}

func redactConfig(value string) string {
	lines := strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n")
	for i, line := range lines {
		if secretLikeLine(line) {
			if idx := strings.Index(line, "="); idx >= 0 {
				lines[i] = strings.TrimRight(line[:idx+1], " \t") + " <redacted>"
			} else {
				lines[i] = "<redacted>"
			}
		}
	}
	return strings.Join(lines, "\n")
}

func secretLikeLine(line string) bool {
	lower := strings.ToLower(strings.TrimSpace(line))
	if lower == "" || strings.HasPrefix(lower, "#") {
		return false
	}
	for _, needle := range []string{
		"password", "api_key", "api-token", "secret", "private_key", "privkey", "server_api_header", "hash_nonce", "dns_cloudflare_api_token",
	} {
		if strings.Contains(lower, needle) {
			return true
		}
	}
	return false
}
