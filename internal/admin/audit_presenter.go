package admin

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"proidentity-mail/internal/domain"
)

func enrichAuditEvents(events []domain.AuditEvent) []domain.AuditEvent {
	out := make([]domain.AuditEvent, len(events))
	for i, event := range events {
		out[i] = enrichAuditEvent(event)
	}
	return out
}

func enrichAuditEvent(event domain.AuditEvent) domain.AuditEvent {
	metadata := auditMetadata(event.MetadataJSON)
	event.Category = auditCategory(event.Action)
	event.Severity = auditSeverity(event.Action, metadata)
	event.ActorLabel = auditActorLabel(event, metadata)
	event.TargetLabel = auditTargetLabel(event, metadata)
	event.Title = auditTitle(event.Action)
	event.Summary = auditSummary(event, metadata)
	event.Details = auditDetails(metadata)
	return event
}

func auditMetadata(raw string) map[string]any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]any{}
	}
	var metadata map[string]any
	if err := json.Unmarshal([]byte(raw), &metadata); err != nil {
		return map[string]any{"raw": raw}
	}
	return metadata
}

func auditCategory(action string) string {
	switch {
	case strings.HasPrefix(action, "security.alert."):
		return "security"
	case strings.HasPrefix(action, "admin.") || strings.HasPrefix(action, "webmail.login") || action == "webmail.logout":
		return "auth"
	case strings.HasPrefix(action, "message.report") || strings.HasPrefix(action, "quarantine.") || action == "tenant_policy.update":
		return "mail_security"
	case strings.HasPrefix(action, "webmail.") || strings.HasPrefix(action, "message."):
		return "user_activity"
	case strings.HasPrefix(action, "cloudflare_") || strings.HasPrefix(action, "mail_server_settings.") || strings.HasPrefix(action, "backup.") || strings.HasPrefix(action, "proxy."):
		return "system"
	default:
		return "admin"
	}
}

func auditSeverity(action string, metadata map[string]any) string {
	switch {
	case strings.HasPrefix(action, "security.alert."):
		return "warning"
	case strings.Contains(action, "failed") || action == "quarantine.deleted" || metaString(metadata, "status") == "deleted":
		return "danger"
	case strings.Contains(action, "spam") || strings.Contains(action, "quarantine") || strings.Contains(action, "malware") || strings.Contains(action, "locked"):
		return "warning"
	case strings.HasSuffix(action, ".delete") || strings.HasSuffix(action, ".deleted"):
		return "danger"
	case strings.Contains(action, "login") || strings.Contains(action, "create") || strings.Contains(action, "released"):
		return "success"
	default:
		return "info"
	}
}

func auditActorLabel(event domain.AuditEvent, metadata map[string]any) string {
	switch event.ActorType {
	case "admin":
		if event.TargetType == "admin" && event.TargetID != "" && strings.HasPrefix(event.Action, "admin.") {
			return "Admin " + event.TargetID
		}
		return "Admin"
	case "user":
		if email := metaString(metadata, "email"); email != "" {
			return email
		}
		if event.ActorID != nil {
			return fmt.Sprintf("User %d", *event.ActorID)
		}
		return "Mailbox user"
	case "system":
		return "System"
	default:
		if event.ActorType == "" {
			return "Unknown actor"
		}
		return strings.Title(strings.ReplaceAll(event.ActorType, "_", " "))
	}
}

func auditTargetLabel(event domain.AuditEvent, metadata map[string]any) string {
	switch event.TargetType {
	case "message":
		if subject := metaString(metadata, "subject"); subject != "" {
			return "Message: " + subject
		}
		return "Message " + event.TargetID
	case "tenant":
		if name := metaString(metadata, "name"); name != "" {
			return "Tenant: " + name
		}
	case "domain":
		if name := metaString(metadata, "name"); name != "" {
			return "Domain: " + name
		}
	case "user":
		if local := metaString(metadata, "local_part"); local != "" {
			return "User: " + local
		}
	case "quarantine_event":
		if recipient := metaString(metadata, "recipient"); recipient != "" {
			return "Quarantine: " + recipient
		}
	case "contact":
		if name := metaString(metadata, "name"); name != "" {
			return "Contact: " + name
		}
	case "calendar_event":
		if title := metaString(metadata, "title"); title != "" {
			return "Calendar: " + title
		}
	}
	if event.TargetType == "" {
		return event.TargetID
	}
	if event.TargetID == "" {
		return strings.ReplaceAll(event.TargetType, "_", " ")
	}
	return strings.ReplaceAll(event.TargetType, "_", " ") + " " + event.TargetID
}

func auditTitle(action string) string {
	titles := map[string]string{
		"admin.login":                                   "Admin signed in",
		"admin.login_failed":                            "Admin sign-in failed",
		"admin.logout":                                  "Admin signed out",
		"security.alert.admin_new_ip":                   "Admin sign-in from new IP",
		"security.alert.auth_spray":                     "Authentication spray detected",
		"security.alert.backup_manual":                  "Manual backup run detected",
		"security.alert.bulk_send":                      "Bulk send threshold exceeded",
		"backup.completed":                              "Backup completed",
		"admin.mfa.challenge_created":                   "Admin MFA challenge created",
		"admin.mfa.step_up":                             "Admin MFA step-up approved",
		"admin.mfa_failed":                              "Admin MFA failed",
		"admin.mfa.totp_enrollment_created":             "Admin TOTP enrollment started",
		"admin.mfa.totp_enabled":                        "Admin TOTP enabled",
		"admin.mfa.proidentity.update":                  "ProIdentity Auth settings changed",
		"admin.mfa.proidentity_totp_enrollment_created": "ProIdentity hosted TOTP enrollment started",
		"admin.mfa.proidentity_totp_code_verified":      "ProIdentity hosted TOTP code verified",
		"admin.mfa.proidentity_totp_enabled":            "ProIdentity hosted TOTP enabled",
		"admin.mfa.webauthn_registered":                 "Admin hardware key registered",
		"webmail.login":                                 "Mailbox user signed in",
		"webmail.login_failed":                          "Mailbox sign-in failed",
		"webmail.login_locked":                          "Mailbox sign-in locked",
		"webmail.logout":                                "Mailbox user signed out",
		"webmail.password_change":                       "Mailbox password changed",
		"webmail.password_change_failed":                "Mailbox password change failed",
		"message.report_spam":                           "User marked message as spam",
		"message.report_ham":                            "User marked message as not spam",
		"message.send":                                  "Message sent from webmail",
		"message.move":                                  "Message moved",
		"message.delete":                                "Message permanently deleted",
		"quarantine.created":                            "Message held in quarantine",
		"quarantine.released":                           "Quarantined message released",
		"quarantine.deleted":                            "Quarantined message deleted",
		"tenant.create":                                 "Tenant created",
		"tenant.update":                                 "Tenant updated",
		"tenant.delete":                                 "Tenant disabled",
		"domain.create":                                 "Domain created",
		"domain.update":                                 "Domain updated",
		"domain.delete":                                 "Domain disabled",
		"user.create":                                   "User created",
		"user.update":                                   "User updated",
		"user.delete":                                   "User disabled",
		"alias.create":                                  "Alias created",
		"alias.update":                                  "Alias updated",
		"alias.delete":                                  "Alias removed",
		"catch_all.create":                              "Catch-all mailbox created",
		"catch_all.update":                              "Catch-all mailbox updated",
		"catch_all.delete":                              "Catch-all mailbox removed",
		"shared_permission.create":                      "Shared mailbox permission granted",
		"shared_permission.update":                      "Shared mailbox permission changed",
		"shared_permission.delete":                      "Shared mailbox permission removed",
		"tenant_policy.update":                          "Mail security policy changed",
		"cloudflare_config.update":                      "Cloudflare settings updated",
		"cloudflare_dns.apply":                          "Cloudflare DNS changes applied",
		"mail_server_settings.update":                   "Mail server behavior changed",
		"webmail.folder_create":                         "Mailbox folder created",
		"webmail.folder_delete":                         "Mailbox folder deleted",
		"webmail.filter_create":                         "Mailbox filter created",
		"webmail.filter_update":                         "Mailbox filter updated",
		"webmail.filter_delete":                         "Mailbox filter deleted",
		"webmail.contact_create":                        "Contact created",
		"webmail.contact_update":                        "Contact updated",
		"webmail.contact_delete":                        "Contact deleted",
		"webmail.calendar_create":                       "Calendar event created",
		"webmail.calendar_update":                       "Calendar event updated",
		"webmail.calendar_delete":                       "Calendar event deleted",
	}
	if title, ok := titles[action]; ok {
		return title
	}
	parts := strings.Split(action, ".")
	for i := range parts {
		parts[i] = strings.Title(strings.ReplaceAll(parts[i], "_", " "))
	}
	return strings.Join(parts, " ")
}

func auditSummary(event domain.AuditEvent, metadata map[string]any) string {
	switch event.Action {
	case "security.alert.admin_new_ip":
		return fmt.Sprintf("Admin %s signed in from a new client IP %s.", metaFallback(metadata, "username", event.TargetID), metaFallback(metadata, "client_ip", "unknown"))
	case "security.alert.auth_spray":
		return fmt.Sprintf("%s failed logins touched %s accounts from %s in %s seconds.", metaFallback(metadata, "service", "A service"), metaFallback(metadata, "distinct_accounts", "multiple"), metaFallback(metadata, "client_ip", event.TargetID), metaFallback(metadata, "window_seconds", "60"))
	case "security.alert.backup_manual":
		return fmt.Sprintf("Backup %s was started outside the scheduled timer path.", metaFallback(metadata, "archive_name", event.TargetID))
	case "security.alert.bulk_send":
		return fmt.Sprintf("%s sent one message to %s recipients.", metaFallback(metadata, "email", event.TargetID), metaFallback(metadata, "recipient_count", "many"))
	case "backup.completed":
		return fmt.Sprintf("Backup %s completed with %s files and %s verified entries.", metaFallback(metadata, "archive_name", event.TargetID), metaFallback(metadata, "files", "unknown"), metaFallback(metadata, "verified_entries", "unknown"))
	case "admin.mfa.challenge_created":
		return fmt.Sprintf("Admin MFA challenge created using %s.", metaFallback(metadata, "provider", "configured provider"))
	case "admin.mfa.step_up":
		return fmt.Sprintf("A fresh admin MFA step-up was approved using %s.", metaFallback(metadata, "provider", "configured provider"))
	case "admin.mfa_failed":
		return fmt.Sprintf("Admin MFA failed using %s.", metaFallback(metadata, "provider", "configured provider"))
	case "admin.mfa.proidentity.update":
		return fmt.Sprintf("ProIdentity Auth was %s for admin login.", map[bool]string{true: "enabled or updated", false: "disabled"}[metaBool(metadata, "enabled")])
	case "message.report_spam", "message.report_ham":
		verdict := metaString(metadata, "verdict")
		if verdict == "ham" {
			verdict = "not spam"
		}
		return strings.TrimSpace(fmt.Sprintf("%s trained message %s as %s.", auditActorLabel(event, metadata), event.TargetID, verdict))
	case "quarantine.created":
		return fmt.Sprintf("%s was held as %s by %s.", metaFallback(metadata, "recipient", "unknown recipient"), metaFallback(metadata, "verdict", "unknown"), metaFallback(metadata, "scanner", "scanner"))
	case "quarantine.released", "quarantine.deleted":
		return fmt.Sprintf("%s was %s by an admin.", metaFallback(metadata, "recipient", "A quarantined message"), metaFallback(metadata, "status", strings.TrimPrefix(event.Action, "quarantine.")))
	case "webmail.login", "webmail.logout", "webmail.login_failed":
		return fmt.Sprintf("%s for %s.", auditTitle(event.Action), metaFallback(metadata, "email", event.TargetID))
	case "webmail.password_change":
		return fmt.Sprintf("Password changed for %s.", metaFallback(metadata, "email", event.TargetID))
	case "message.send":
		return fmt.Sprintf("Sent to %s recipient(s).", metaFallback(metadata, "recipient_count", "unknown"))
	case "message.move":
		return fmt.Sprintf("Moved %s message(s) to %s.", metaFallback(metadata, "count", "1"), metaFallback(metadata, "folder", "folder"))
	case "message.delete":
		return fmt.Sprintf("Permanently deleted %s message(s) from Trash.", metaFallback(metadata, "count", "1"))
	}
	if email := metaString(metadata, "email"); email != "" {
		return fmt.Sprintf("%s by %s.", auditTitle(event.Action), email)
	}
	if name := metaString(metadata, "name"); name != "" {
		return fmt.Sprintf("%s: %s.", auditTitle(event.Action), name)
	}
	if status := metaString(metadata, "status"); status != "" {
		return fmt.Sprintf("%s, status %s.", auditTitle(event.Action), status)
	}
	return auditTitle(event.Action) + "."
}

func auditDetails(metadata map[string]any) []domain.Detail {
	if len(metadata) == 0 {
		return nil
	}
	preferred := []string{"email", "recipient", "sender", "verdict", "status", "folder", "count", "recipient_count", "subject", "name", "domain", "local_part", "mailbox_type", "quota_bytes", "rights", "spam_action", "malware_action", "require_tls_for_auth", "hostname_mode", "mail_hostname", "sni_enabled", "zone_id", "changed", "backup_id", "note"}
	seen := map[string]bool{}
	details := make([]domain.Detail, 0, len(metadata))
	for _, key := range preferred {
		if value, ok := metadata[key]; ok {
			details = append(details, domain.Detail{Label: detailLabel(key), Value: metadataString(value)})
			seen[key] = true
		}
	}
	keys := make([]string, 0, len(metadata))
	for key := range metadata {
		if !seen[key] {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		details = append(details, domain.Detail{Label: detailLabel(key), Value: metadataString(metadata[key])})
	}
	return details
}

func detailLabel(key string) string {
	key = strings.ReplaceAll(key, "_", " ")
	return strings.Title(key)
}

func metaString(metadata map[string]any, key string) string {
	return strings.TrimSpace(metadataString(metadata[key]))
}

func metaFallback(metadata map[string]any, key, fallback string) string {
	if value := metaString(metadata, key); value != "" {
		return value
	}
	return fallback
}

func metaBool(metadata map[string]any, key string) bool {
	switch value := metadata[key].(type) {
	case bool:
		return value
	case string:
		return strings.EqualFold(value, "true") || value == "1" || strings.EqualFold(value, "yes")
	default:
		return false
	}
}

func metadataString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%g", v)
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(bytes)
	}
}
