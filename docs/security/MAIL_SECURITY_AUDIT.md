# Mail Server Security Audit

Date: 2026-05-10
Status: rolling pre-production audit and hardening pass

This document tracks the security inventory, confirmed findings, code fixes, and remaining work for the ProIdentity Mail platform. It is intentionally practical: every fixed item links to the code path that was changed, and every open item remains visible until implemented and verified.

## Phase 1 Inventory

### Service Map

#### Mail Core
- Public protocols: SMTP 25, submission 587, SMTPS 465, IMAP 143/993 through Dovecot, POP3 110/995 through Dovecot, LMTP over the Postfix/Dovecot private socket.
- Config generation: `internal/render/templates.go`, `internal/render/render.go`, `cmd/mailctl/main.go`.
- Runtime components: Postfix, Dovecot, Rspamd, ClamAV.
- Shared state: MariaDB tables for tenants, domains, users, aliases, catch-all routes, DKIM keys, TLS inventory; Maildir under `/var/vmail`; DKIM keys under `/var/lib/rspamd/dkim`; TLS keys under `/etc/letsencrypt` and configured custom paths.
- Process users: Postfix and Dovecot service users from OS packages; mailbox files owned by `vmail`; Rspamd DKIM keys owned by `_rspamd`.

#### Autoconfig
- Public HTTP endpoints: `/.well-known/autoconfig/mail/config-v1.1.xml`, `/.well-known/proidentity-mail/config.json`, `/autodiscover/autodiscover.xml`.
- Implemented in `internal/admin/handlers.go`.
- Inputs: email/domain query parameters and Outlook XML POST bodies.
- External dependencies: configured DNS hostnames and mail server settings from MariaDB.

#### Webmail
- Public HTTP/API routes: `/`, `/api/v1/session`, `/api/v1/mailboxes`, `/api/v1/messages`, `/api/v1/messages/{id}`, batch move/delete, send, drafts, folders, filters, contacts, calendar, profile, password, content-trust.
- Implemented in `internal/webmail/handlers.go`, `internal/webmail/maildir.go`, `internal/webmail/smtp_sender.go`, `internal/webmail/static.go`.
- Shared state: MariaDB auth/profile/DAV metadata, Maildir, local SMTP on `127.0.0.1:25`, Rspamd learning commands.
- Process user: `vmail`.

#### Admin Panel
- Public HTTP/API routes: `/`, `/api/v1/session`, tenants, domains, users, aliases, catch-all, shared permissions, DNS, Cloudflare DNS, TLS, quarantine, audit, policies, mail server settings.
- Implemented in `internal/admin/handlers.go`, `internal/admin/sql_store.go`, `internal/admin/static.go`.
- Shared state: full management access to MariaDB tenant/domain/user/security tables; writes generated config through `mailctl` flows.
- Process user: `proidentity`.

#### Server Manager
- No separate public HTTP server exists yet. Current implementation is the `mailctl` CLI plus systemd units/timers.
- Operations: render configs, sync proxy, process TLS jobs, call certbot, restart/apply mail config, sync Rspamd policy.
- OS interactions: `cmd/mailctl/main.go`, `deploy/devmail/apply-mail-config.sh`.
- Process user: several operations run as root through systemd or manual execution. This remains a high-risk trust boundary.

#### Backuper
- No public HTTP API exists yet. Current implementation is `mailctl backup`, `backup-prune`, `backup-verify`, and `restore`.
- Inputs: archive path, output path, target root, retention flags, restore flags.
- Reads: `/etc/proidentity-mail`, `/var/vmail`, `/var/lib/proidentity-mail`, `/var/lib/rspamd/dkim`, `/etc/nginx`, `/etc/letsencrypt`, database dump.
- Process user: backup/restore systemd unit currently runs as root. This remains a critical blast-radius item even with encryption.

### Trust Boundaries

- MariaDB is shared by admin, webmail, groupware, mailctl, Postfix MySQL maps, and Dovecot SQL auth. A DB compromise can affect all services.
- Admin and webmail use separate cookie names and separate session kinds. Basic Auth fallback is disabled when session managers are configured.
- Webmail cannot directly call mailctl/server-manager operations through an HTTP API, but can indirectly cause SMTP delivery, Rspamd learning, Maildir writes, and DAV writes.
- Admin can configure domains, TLS, DNS, and user routing; TLS worker/mailctl can change OS-level config and restart services.
- Backuper can read mail content, app env, DB dumps, DKIM keys, TLS material, and generated configs. Its archive confidentiality is critical.

### Input Entry Points

- SMTP/IMAP/POP inputs are handled by Postfix/Dovecot, with generated config from the app.
- Webmail inputs: login credentials, message IDs, folder names, compose headers/body/attachments, contacts, calendar events, filters, profile settings, content-trust rules.
- Admin inputs: all tenant/domain/user/routing fields, Cloudflare token, DNS record apply flags, custom cert/key paths, mail server hostname/IP settings.
- Autoconfig inputs: query email/domain and Outlook autodiscover XML body.
- Backuper inputs: archive/output/target paths, service control flags, retention flags.
- External data: inbound email headers/body/attachments, DNS provider API responses, TLS certificate metadata, Rspamd/ClamAV verdicts, database backup metadata.

### File And OS Interactions

- Mail storage: `internal/webmail/maildir.go` reads/writes/moves/deletes Maildir files under `/var/vmail`.
- Attachments: `internal/webmail/handlers.go` reads uploaded attachment parts into outbound MIME messages.
- Config generation: `cmd/mailctl/main.go` writes generated files under the configured config dir; `deploy/devmail/apply-mail-config.sh` installs them into Postfix/Dovecot/Rspamd/Nginx paths.
- TLS and DNS: `cmd/mailctl/main.go` calls `certbot`, `nginx`, `systemctl`, and app-owned scripts with argument arrays.
- Backups: `internal/backup/backup.go` walks allowed source trees, creates/restores archives, skips symlinks, validates restore targets.

## Fixed Findings

### F-001: Backend Services Bound To Public Interfaces
- CWE: CWE-200 / CWE-284
- Services: Webmail, Admin Panel, Groupware
- Severity: High
- Evidence: backend defaults previously used `0.0.0.0`.
- Fix: default backend binds are now loopback in `internal/app/config.go`, `deploy/proidentity-production-setup.sh`, `deploy/devmail/proidentity-mail.env.example`, and `deploy/devmail/setup-runtime.sh`.
- Verification: public dev server was checked with `ss` and external curl; only nginx is public.

### F-002: Basic Auth Fallback Accepted On Session-Protected APIs
- CWE: CWE-287
- Services: Webmail, Admin Panel
- Severity: High
- Fix: `internal/admin/handlers.go` and `internal/webmail/handlers.go` reject Basic Auth when session managers are configured.
- Verification: tests cover wrong Basic Auth returning JSON 401 without `WWW-Authenticate`; public dev server was checked.

### F-003: Catch-All Routing Could Capture Real Aliases Or Users
- CWE: CWE-284
- Services: Mail Core
- Severity: High
- Fix: `internal/render/render.go` catch-all SQL now uses `NOT EXISTS` against active users and aliases before returning a catch-all route.
- Verification: `internal/render/render_test.go`.

### F-004: Missing Browser Security Headers And Request Body Caps
- CWE: CWE-693 / CWE-770
- Services: Webmail, Admin Panel, Groupware
- Severity: Medium
- Fix: `internal/security/headers.go` adds baseline browser headers and `http.MaxBytesReader` middleware; wired into admin, webmail, and groupware routers.
- Verification: `internal/security/headers_test.go`, deployed header checks on dev HTTPS endpoints.

### F-005: Admin Resource Names And Custom TLS Paths Were Under-Validated
- CWE: CWE-20 / CWE-22
- Services: Admin Panel, Server Manager
- Severity: High
- Fix: `internal/admin/handlers.go` normalizes tenant slugs, domains, DKIM selectors, local parts, email destinations, hostnames, IP settings, and allows custom cert/key paths only under approved certificate roots.
- Verification: `internal/admin/handlers_test.go`.

### F-006: Autoconfig Accepted Invalid Email/Domain Inputs
- CWE: CWE-20 / CWE-611
- Services: Autoconfig
- Severity: Medium
- Fix: `internal/admin/handlers.go` validates autoconfig email/domain inputs through normalized email parsing before rendering XML/JSON.
- Verification: `TestAutoconfigRejectsInvalidEmailAndDomainInput`.

### F-007: Webmail Compose Allowed Header/Recipient Injection Edges
- CWE: CWE-93 / CWE-20
- Services: Webmail, Mail Core
- Severity: High
- Attack scenario: a malicious recipient or subject containing CRLF could try to inject extra SMTP headers into outbound messages.
- Fix: `internal/webmail/handlers.go:616` validates outbound recipients with `net/mail`, rejects CR/LF/null bytes and invalid domains; `internal/webmail/smtp_sender.go:189` truncates header values at the first newline before writing RFC822 headers.
- Verification: `internal/webmail/smtp_sender_test.go`.

### F-008: Outbound STARTTLS Disabled Certificate Verification
- CWE: CWE-295
- Services: Webmail, Mail Core
- Severity: High
- Attack scenario: a MITM on a remote SMTP hop could present any certificate if outbound sender code disabled verification.
- Fix: `internal/webmail/smtp_sender.go:51` now uses a TLS config with `ServerName` and TLS 1.2+ for non-loopback SMTP hosts; loopback local relay skips unnecessary STARTTLS.
- Verification: `TestSMTPStartTLSConfigNeverDisablesVerification`.

### F-009: Postfix Port 25 Exposed SMTP AUTH And Lacked Sender Login Enforcement
- CWE: CWE-287 / CWE-290
- Services: Mail Core
- Severity: Critical
- Attack scenario: an exposed relay listener advertising auth on port 25 and no sender-login map can increase credential attack surface and allow authenticated spoofing of local/shared senders.
- Fix: `internal/render/templates.go:23` disables global SMTP AUTH; submission/SMTPS enable auth only on 587/465 and enforce `reject_authenticated_sender_login_mismatch`; `internal/render/render.go:177` renders `sender-login-maps.cf` for own addresses, aliases, and shared mailbox send-as; `cmd/mailctl/main.go:199` renders the map; `deploy/devmail/apply-mail-config.sh:62` installs it.
- Verification: `internal/render/render_test.go`.

### F-010: Mail Core Lacked Baseline Protocol Hardening
- CWE: CWE-326 / CWE-319 / CWE-200
- Services: Mail Core
- Severity: High
- Fix: `internal/render/templates.go` disables VRFY, disables TLS 1.0/1.1 and weak ciphers, adds message size limits, rejects unauth pipelining, adds Spamhaus ZEN on inbound SMTP, and enables SMTPS 465.
- Verification: `internal/render/render_test.go`.

### F-011: Session Cookies Were Not Strict Enough For Admin/Webmail
- CWE: CWE-352 / CWE-613
- Services: Webmail, Admin Panel
- Severity: Medium
- Fix: `internal/session/manager.go` defaults cookies to `SameSite=Strict`; `cmd/webadmin/main.go:37` sets admin TTL to 15 minutes and stricter login thresholds; webmail explicitly uses Strict cookies.
- Verification: `internal/session/manager_test.go`.

### F-012: Backups Were Plain Tarballs Containing Mail, DB Dumps, DKIM Keys, TLS Keys, And App Secrets
- CWE: CWE-311 / CWE-522
- Services: Backuper
- Severity: Critical
- Attack scenario: a copied backup archive reveals every mailbox and most platform secrets.
- Fix: `internal/backup/backup.go:25` adds a chunked AES-GCM encrypted archive format; `cmd/mailctl/main.go:1032` makes encrypted output the default and requires `PROIDENTITY_BACKUP_ENCRYPTION_KEY` unless `--allow-plain` is explicitly used; setup scripts generate/store the key.
- Verification: `internal/backup/backup_test.go`, `internal/backup/prune_test.go`, `cmd/mailctl` tests.

### F-013: Web/Admin Login Lockout Was Memory-Only
- CWE: CWE-307
- Services: Webmail, Admin Panel
- Severity: High
- Attack scenario: restarting the service cleared failed login counters, and a distributed attack could avoid the single IP+account key.
- Fix: `internal/session/sql_limiter.go` adds a SQL-backed limiter; `internal/db/migrations/0015_login_rate_limits.sql` persists failure counters and lock expiry; admin and webmail now record IP, account, and IP+account keys.
- Verification: `internal/session`, `internal/admin`, and `internal/webmail` tests.

### F-014: Webmail HTML Sanitizer Was Regex-Based And Allowed Unsafe Links
- CWE: CWE-79 / CWE-80
- Services: Webmail
- Severity: High
- Attack scenario: hostile HTML mail could keep encoded `javascript:` links or non-allowlisted markup that later runs in the mail reader context.
- Fix: `internal/webmail/maildir.go` now uses the `golang.org/x/net/html` tokenizer with explicit tag/attribute allowlists; unsafe links and schemes are dropped; external http/https images are deferred with `data-external-src`; unsafe image schemes are not deferrable.
- Verification: `TestSanitizeMailHTMLDropsUnsafeLinksAndNonAllowlistedTags`, `TestSanitizeMailHTMLOmitsUnsafeImageSchemesInsteadOfDeferringThem`, and full webmail tests.

### F-015: IMAP/POP/SMTP AUTH Failures Did Not Feed Native Persistent Lockout
- CWE: CWE-307
- Services: Mail Core, Admin Panel
- Severity: High
- Attack scenario: a remote attacker could brute-force Dovecot-backed IMAP/POP/SMTP AUTH across service restarts without those failures entering the application-owned persistent limiter.
- Fix: `internal/render/templates.go` enables Dovecot auth-policy callbacks with a secret API header and cluster nonce; `cmd/mailctl/main.go` renders the callback URL; `internal/admin/handlers.go` exposes a loopback-only `/internal/dovecot/auth-policy` endpoint that records Dovecot report failures into the SQL limiter and rejects locked logins; `internal/render/templates.go` also hides `/internal/` from the public admin nginx host.
- Verification: `TestDovecotAuthPolicyRequiresLoopbackAndToken`, `TestDovecotAuthPolicyReportsFailuresToLimiter`, `TestDovecotAuthPolicyBlocksLockedLogin`, `TestDovecotAuthPolicyKeysDoNotTreatPolicyServerLoopbackAsClient`, `TestRenderDovecotLocalConfiguresAuthPolicyWhenProvided`, and `TestRenderNginxProxyBlocksInternalAdminCallbacksPublicly`.

### F-016: Login Limiter Used Flat Lockout And Protocol Connection Caps Were Too Loose
- CWE: CWE-307 / CWE-770
- Services: Mail Core, Webmail, Admin Panel
- Severity: High
- Attack scenario: an attacker could pace attempts around a flat lockout threshold and open too many SMTP/IMAP/POP login connections while avoiding escalating penalties.
- Fix: `internal/session/manager.go` and `internal/session/sql_limiter.go` now use a sliding one-hour failure window with progressive penalties: 4-5 failures = 30 seconds, 6-10 = 5 minutes, 11-20 = 1 hour, 21+ = 24 hours; admin uses stricter thresholds. `internal/render/templates.go` now renders Postfix anvil limits for connection count/rate, AUTH rate, TLS session rate, and Dovecot `mail_max_userip_connections` plus high-security login process limits.
- Verification: `TestDefaultPenaltyScheduleMatchesSecurityPolicy`, `TestAdminPenaltyScheduleIsStricter`, `TestLoginLimiterDefaultUsesProgressiveSchedule`, `TestLoginLimiterKeepsFailureCountAfterTemporaryLockExpires`, `TestSQLLoginLimiterUsesSamePenaltySchedule`, Postfix/Dovecot render tests, and dev-server verification that the fourth Dovecot auth-policy failure returns `status:-1`.

### F-017: Failed Account Attacks Did Not Lock Mailbox Accounts Or Expose Release Controls
- CWE: CWE-307
- Services: Mail Core, Webmail, Admin Panel
- Severity: High
- Attack scenario: repeated failures against one mailbox from many IPs could keep generating persistent limiter rows without changing the mailbox state, and admins had no safe UI to inspect/clear limiter rows or unlock affected users.
- Fix: `internal/session/sql_limiter.go` now locks active user mailbox records after 10 account-scoped failures; `internal/admin/handlers.go` adds authenticated unlock and login-rate-limit management APIs; `internal/admin/sql_store.go` clears account/pair limiter rows when unlocking; `internal/admin/static.go` adds Security view panels for locked accounts and native login protection state.
- Verification: `TestSQLLoginLimiterUsesSamePenaltySchedule`, `TestAccountEmailFromLimiterKeyOnlyAcceptsAccountEmailKeys`, `TestListLoginRateLimitsReturnsFriendlyRows`, `TestListLoginRateLimitsReturnsEmptyArray`, `TestUnlockUserActivatesUserAndClearsLimiterRows`, `TestAdminIndexIncludesNativeLoginProtectionUI`, full `go test ./...`, and dev-server smoke tests showing 10 Dovecot auth-policy failures lock a mailbox and the admin login-rate-limits API returns a JSON array through HTTPS.

### F-018: Admin Panel Had No MFA Enforcement Path
- CWE: CWE-308 / CWE-287
- Services: Admin Panel
- Severity: High
- Attack scenario: a leaked admin password could create a full admin session and allow tenant, DNS, TLS, routing, and security-policy changes.
- Fix: `internal/admin/mfa.go` adds a password-plus-MFA login challenge flow; local TOTP enrollment uses standard otpauth URLs and QR PNGs; ProIdentity Auth service-provider integration can create push/hardware-key auth requests and poll approval status with `X-API-Key`; ProIdentity Auth takes precedence over local TOTP when enabled. `internal/db/migrations/0016_admin_mfa.sql` persists MFA settings and short-lived challenges; `internal/admin/static.go` adds System tabs for Admin MFA and ProIdentity Auth.
- Verification: `TestAdminLoginRequiresLocalTOTPWhenEnabled`, `TestAdminLoginUsesProIdentityAuthWhenConfigured`, `TestAdminTOTPEnrollmentReturnsQRAndVerifyEnablesMFA`, `TestProIdentityAuthSettingsDoNotExposeAPIKey`, `TestAdminIndexIncludesMFASettingsAndProIdentityAuthTab`, full `go test ./...`, and dev-server HTTPS checks for the MFA UI/settings API.

### F-019: Autoconfig Endpoints Had No Native Rate Limit
- CWE: CWE-307 / CWE-770
- Services: Autoconfig, Admin Panel
- Severity: Medium
- Attack scenario: an attacker could enumerate or hammer Thunderbird/Outlook discovery endpoints by domain and client IP without entering the persistent limiter.
- Fix: `internal/admin/handlers.go` adds a dedicated discovery limiter with client-IP and domain keys, backed by SQL in `cmd/webadmin/main.go` when MariaDB is configured.
- Verification: `TestDiscoveryEndpointsUseDedicatedRateLimiterWithTrustedClientIP`, `TestDiscoveryEndpointReturns429WhenRateLimited`, and `go test ./internal/admin ./cmd/webadmin`.

### F-020: High-Risk Admin Operations Did Not Require Fresh Step-Up MFA
- CWE: CWE-287 / CWE-306
- Services: Admin Panel, Server Manager boundary
- Severity: High
- Attack scenario: a stolen active admin session could immediately change DNS/TLS settings, queue live config apply, clear lockouts, unlock users, or reset user MFA.
- Fix: `internal/session/manager.go` adds recent step-up state; `internal/admin/mfa.go` adds `/api/v1/session/step-up` and `/api/v1/session/step-up/verify`; high-risk admin handlers now require a fresh step-up; `internal/admin/static.go` retries protected UI actions after a built-in MFA modal.
- Verification: `TestDangerousAdminOperationRequiresFreshStepUp`, `TestAdminStepUpFlowAllowsDangerousOperation`, `TestManagerMarksRecentStepUpOnlyWithCSRF`, and full `go test ./...`.

### F-021: Security Alert Hooks Were Missing For High-Signal Events
- CWE: CWE-778
- Services: Admin Panel, Webmail
- Severity: Medium
- Attack scenario: suspicious actions such as an admin sign-in from a new IP or a bulk recipient send could blend into normal audit activity.
- Fix: admin login now records `security.alert.admin_new_ip` when the client IP is unseen for that admin; webmail records `security.alert.bulk_send` when a message reaches the bulk recipient threshold; the audit presenter categorizes these as security alerts.
- Verification: `TestAdminLoginFromNewClientIPRecordsSecurityAlert`, `TestAdminLoginFromKnownClientIPDoesNotRecordSecurityAlert`, `TestSendEndpointRecordsBulkSendSecurityAlert`, and `TestAuditEndpointReturnsReadableSecurityAlerts`.

### F-022: Root Oneshot Jobs Had Weak Systemd Sandboxing
- CWE: CWE-250 / CWE-266
- Services: Server Manager, Backuper
- Severity: High
- Attack scenario: a compromised root-level backup/TLS/config-apply job had broader kernel, device, namespace, and filesystem exposure than necessary.
- Fix: production and dev systemd units for backup, TLS worker, and config apply now include stricter sandboxing: kernel/control-group protection, private devices, SUID restrictions, namespace restrictions, realtime restriction, locked personality, and native syscall architecture.
- Verification: `TestRootOneshotUnitsUseSystemdSandboxing` and `go test ./deploy`.

### F-023: Live Restore Needed A Stronger Confirmation Gate
- CWE: CWE-284 / CWE-306
- Services: Backuper, Server Manager
- Severity: High
- Attack scenario: an admin or automation mistake could run `mailctl restore --live --apply` against the live system without a restore-specific confirmation phrase.
- Fix: `cmd/mailctl/main.go` requires `--confirm-live-restore="RESTORE <archive filename>"` for live apply restores.
- Verification: `TestLiveRestoreConfirmationUsesArchiveFilename` and `go test ./cmd/mailctl ./internal/backup`.

### F-024: Dependency/CVE Audit Was Not Repeatable
- CWE: CWE-1104
- Services: All
- Severity: Medium
- Attack scenario: releases could proceed without a standard dependency inventory and Go vulnerability reachability check.
- Fix: `scripts/security-check.ps1` runs `go test ./...`, `go list -m all`, and `govulncheck`; `README.md` and `deploy/PRODUCTION_SETUP.md` include it in the build path; `.github/workflows/security-check.yml` enforces the same script on push, pull request, and manual CI runs.
- Verification: `powershell -ExecutionPolicy Bypass -File .\scripts\security-check.ps1` completed with no reachable vulnerabilities; `TestProductionRunbookIncludesSecurityCheckScript` verifies the script and workflow.

### F-025: Outbound Attachment Metadata Was Under-Normalized
- CWE: CWE-93 / CWE-20
- Services: Webmail
- Severity: Medium
- Attack scenario: malicious attachment filenames or content types could try to inject MIME headers or disguise filenames with bidi/control characters.
- Fix: `internal/webmail/handlers.go` strips filename control/bidi characters, normalizes length by rune, rejects CR/LF/null-bearing content types, parses MIME types server-side, and falls back to content sniffing or `application/octet-stream`.
- Verification: `TestAttachmentMetadataIsSanitized` and webmail attachment tests.

### F-026: Multiple-Account Auth Spray Alerts Were Missing
- CWE: CWE-307 / CWE-778
- Services: Webmail, Groupware, Mail Core auth policy, Admin Panel
- Severity: Medium
- Attack scenario: one IP could try passwords across many accounts and only appear as individual failed logins instead of a coordinated spray attempt.
- Fix: `internal/session/sql_limiter.go` now detects distinct account failures from one client IP within a one-minute window and writes a throttled `security.alert.auth_spray` audit event through the SQL-backed limiter used by webmail, DAV, and Dovecot auth policy.
- Verification: `TestParsePairLimiterKeyValidatesShapeAndControlCharacters`, `TestAuditEndpointReturnsReadableSecurityAlerts`, and focused `go test ./internal/session ./internal/admin`.

### F-027: Hardware-Key MFA Could Not Be Used For Admin Step-Up
- CWE: CWE-287 / CWE-306
- Services: Admin Panel
- Severity: High
- Attack scenario: deployments using native hardware keys for admin MFA had login protection, but dangerous-operation step-up could not complete with the same hardware key.
- Fix: `internal/admin/webauthn.go` adds `/api/v1/session/step-up/webauthn` and shared assertion verification; `internal/admin/static.go` launches browser WebAuthn from the admin step-up modal and marks the session step-up after successful hardware-key verification.
- Verification: `TestAdminIndexIncludesMFASettingsAndProIdentityAuthTab`, `TestDangerousAdminOperationRequiresFreshStepUp`, `TestAdminStepUpFlowAllowsDangerousOperation`, and focused `go test ./internal/admin`.

### F-028: Manual Backup Runs Did Not Create Security-Visible Audit Events
- CWE: CWE-778
- Services: Backuper, Admin Panel audit
- Severity: Medium
- Attack scenario: a manual backup run could happen outside the timer path and only leave files/logs, making unexpected data-copy activity harder to review from the admin audit UI.
- Fix: `cmd/mailctl/main.go` records `backup.completed` for every successful backup when DB access is configured and additionally records `security.alert.backup_manual` when the run was not started by the timer-marked systemd service; production/dev backup services now set `PROIDENTITY_BACKUP_SCHEDULED=1`.
- Verification: `TestBackupAuditActionsAlertOnlyForManualRuns`, `TestProductionSetupKeepsBackupsRootOnly`, `TestAuditEndpointReturnsReadableSecurityAlerts`, and focused `go test ./cmd/mailctl ./deploy ./internal/admin`.

### F-029: Root Timer Jobs Ran Mail Tooling Directly As Root
- CWE: CWE-250 / CWE-269
- Services: Backuper, Server Manager boundary, TLS worker, Config apply
- Severity: Critical
- Attack scenario: if a root timer/path unit or its environment was abused, it launched large application binaries directly as root for backup, TLS, and config apply operations.
- Fix: production and dev unit templates now run the timer/path services as the `proidentity` user and call `/usr/bin/sudo -n -E /opt/proidentity-mail/bin/proidentity-rootctl <subcommand>`; the helper is owned by `root:proidentity`, installed `0750`, and sudoers allows only the fixed `backup`, `tls-worker`, `config-apply`, and `sync-proxy` subcommands while preserving `PROIDENTITY_*` environment variables.
- Verification: `TestPrivilegedJobsUseScopedRootHelper`, `TestRootOneshotUnitsUseSystemdSandboxing`, and focused `go test ./deploy`.

### F-030: Dependency Exception Handling Was Not Documented
- CWE: CWE-1104
- Services: All
- Severity: Low
- Attack scenario: an unreachable dependency finding could be ignored without an owner, expiry, or remediation plan.
- Fix: `docs/security/DEPENDENCY_EXCEPTION_PROCESS.md` defines release-blocking rules, maximum exception lifetimes, required exception fields, and the default policy that reachable High/Critical dependency vulnerabilities may not ship.
- Verification: documentation review and `scripts/security-check.ps1` CI wiring.

## Open Findings

### O-001: Protocol Abuse Protection Still Needs Connection-Level And Account-Lockout Depth
- Services: SMTP AUTH, IMAP, POP3, Webmail, Admin Panel, Autoconfig
- Severity: High
- Current state: web/admin use persistent SQL login counters, Dovecot auth-policy feeds IMAP/POP/SMTP AUTH failures into the same persistent limiter, progressive penalties are implemented, account lockout is automatic, admins can unlock users and clear limiter rows, Postfix/Dovecot connection/rate caps are rendered, autoconfig is rate limited, and alert hooks exist for admin new-IP login, webmail bulk sends, multi-account auth-spray attempts from one IP, and manual backup runs.
- Planned fix: add server-manager business-hours alerts once those operations are exposed as long-running app-managed jobs.

### O-002: Future Restore/Server-Manager APIs Need Step-Up Coverage
- Services: Admin Panel
- Severity: High
- Current state: admin login MFA exists with local TOTP, ProIdentity Auth push/TOTP, and native WebAuthn hardware-key provider integration. Fresh step-up MFA supports local TOTP, ProIdentity Auth, and native WebAuthn for config apply, DNS apply, Cloudflare token changes, TLS certificate requests/settings, user unlock/MFA reset/delete, clearing rate-limit rows, tenant/domain deletion, and mail server behavior changes.
- Planned fix: extend step-up to any future restore/server-manager HTTP operations.

### O-003: Server Manager And Backuper Still Run Root-Level Jobs
- Services: Server Manager, Backuper
- Severity: Critical
- Current state: scheduled backup/TLS/config-apply job control now runs as `proidentity` and escalates only through the allowlisted `proidentity-rootctl` sudo helper; root execution still exists inside the fixed helper subcommands because filesystem, service reload, and backup operations require privileged OS access.
- Planned fix: continue shrinking each helper subcommand into smaller purpose-specific helpers and move restore into an MFA-confirmed operation token flow before any web/API exposure.

### O-004: Restore Needs Stronger Human Confirmation For Live Mode
- Services: Backuper, Server Manager
- Severity: High
- Current state: `mailctl restore --live --apply` verifies archive integrity, supports dry-run, and now requires a restore-specific confirmation phrase.
- Planned fix: require MFA-backed operation tokens once restore is exposed through the future admin/server-manager API.

## Current Built-In Protection Coverage

- SMTP AUTH brute force: persistent Dovecot auth-policy SQL limiter, progressive schedule, and Postfix AUTH/connection-rate caps implemented.
- IMAP brute force: persistent Dovecot auth-policy SQL limiter, progressive schedule, and Dovecot user/IP connection caps implemented.
- POP3 brute force: persistent Dovecot auth-policy SQL limiter, progressive schedule, and Dovecot user/IP connection caps implemented.
- Webmail brute force: persistent SQL limiter for IP/account/IP+account with progressive schedule implemented.
- Admin panel brute force: persistent SQL limiter for IP/account/IP+account with stricter progressive schedule implemented.
- Admin MFA: login MFA implemented with local TOTP QR enrollment and ProIdentity Auth push/hardware-key provider integration; dangerous-operation step-up is implemented for the current admin HTTP surface.
- Account lockout: implemented for active user mailboxes after 10 account failures, with admin unlock and limiter cleanup.
- IP block persistence: persistent IP limiter keys and progressive penalties implemented; admin limiter-row visibility and manual clear controls implemented.
- Admin anomaly alerts: audit log exists; alert hooks cover admin sign-in from a new IP, webmail bulk sends, multi-account auth sprays from one client IP, and manual backup runs.

## Deployment Recommendation

Current state: Conditional for dev/pre-production only.

Do not publish as a customer production product until all Critical and High open findings above are closed, especially continued privilege shrinking for server-manager/backuper and future restore/server-manager operation-token coverage.
