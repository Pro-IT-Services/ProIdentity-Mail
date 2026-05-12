# Production Mail Platform Design

Date: 2026-05-06
Status: Draft for review

## Goal

Build a production-first, multi-tenant mail and groupware platform. The project uses proven mail infrastructure for protocol-heavy and abuse-facing services, while custom Go services provide the product control plane, webadmin, webmail, calendar, contacts, policy management, and future client APIs.

The platform must support real public internet email for multiple tenants, domains, sites, and users. Security, auditability, deliverability, and operational recovery are primary design goals from the first version.

## Recommended Foundation

The first production foundation uses:

- Postfix for SMTP inbound, outbound, queueing, and authenticated submission.
- Dovecot for IMAP, POP3, LMTP delivery, mailbox access, Sieve, and ManageSieve.
- Rspamd with Redis for spam filtering, SPF, DKIM, DMARC, ARC, greylisting, reputation, rate limits, Bayesian learning, neural learning, and custom rules.
- ClamAV clamd for antivirus scanning before final mailbox delivery.
- MariaDB for tenants, domains, users, aliases, credentials, policy, configuration state, audit logs, calendar data, contact data, and web application state.
- Redis for hot cache, sessions, rate limits, Rspamd state, async jobs, and recent mailbox summaries.
- Go services for webadmin, webmail, calendar, contacts, CardDAV, CalDAV, autoconfig, API, config generation, and later JMAP and ActiveSync gateways.

This avoids implementing SMTP, IMAP, and POP3 from scratch in the first version. Those protocols carry decades of client quirks, abuse cases, delivery edge cases, and security risks. Go remains the main product language, but the mail engines are battle-tested.

## Major Components

### Go Control Plane

The control plane owns business and platform state:

- Tenants and tenant-level settings.
- Domains and DNS/security status.
- Users, groups, aliases, catch-all addresses, forwarding, and quotas.
- Authentication policy, password policy, app passwords, MFA-ready account model.
- DKIM key lifecycle.
- Spam policy, quarantine policy, allow/block lists, and custom filter settings.
- Audit logs for administrative and security-sensitive activity.
- Service configuration rendering for Postfix, Dovecot, Rspamd, and supporting daemons.
- Health checks and deployment validation.

The control plane exposes an internal API used by webadmin, webmail, background workers, and service adapters.

### Webadmin

The webadmin is a Go web application for platform operators and tenant admins. It should support:

- Tenant, site, domain, and user management.
- Alias, group, mailbox, and quota management.
- DNS setup guidance for MX, SPF, DKIM, DMARC, MTA-STS, TLS-RPT, autoconfig, autodiscover, CalDAV, and CardDAV.
- Security policy management.
- Spam and malware quarantine policy.
- Audit log review.
- Service health and configuration validation.

### Webmail And Groupware

The webmail is a Go groupware web application, not only a mail reader. It includes:

- Mail UI backed by Dovecot IMAP and Postfix submission.
- Calendar UI backed by the Go calendar service.
- Contacts UI backed by the Go contacts service.
- Spam/not-spam training actions connected to Rspamd learning.
- Malware and quarantine folder views.
- User filter management through Sieve/ManageSieve.

Calendar and contacts are first-class services with their own schema, permissions, sync state, and APIs. The webmail UI is one client of those services.

### Calendar And Contacts

The Go calendar service provides calendar storage, sharing, invitations, reminders, and CalDAV access.

The Go contacts service provides address books, contact groups, sharing, and CardDAV access.

Phones and desktop clients can add the platform as:

- IMAP/SMTP for mail.
- CardDAV for contacts.
- CalDAV for calendars.

ActiveSync can later provide mail, contacts, and calendar in a single mobile account profile.

### Future JMAP And ActiveSync

Dovecot does not provide native JMAP, CalDAV, or CardDAV as the core design assumption. JMAP should be a Go gateway over the platform mail, calendar, and contact APIs if we choose to support it.

ActiveSync is possible, but it is a large later phase. It should be designed as a Go protocol gateway over the same mail, calendar, and contact services rather than as a separate data silo.

## Mail Flow

Inbound mail flow:

1. Remote sender connects to Postfix.
2. Postfix applies connection limits, TLS policy, recipient validation, and MTA policy.
3. Rspamd evaluates SPF, DKIM, DMARC, ARC, reputation, greylisting, content, custom rules, and spam score.
4. ClamAV scans message content and attachments.
5. Policy engine decides reject, defer, quarantine, mark, or deliver.
6. Clean mail is delivered to Dovecot through LMTP.
7. Dovecot applies Sieve rules and stores the message in the mailbox.
8. Go workers update metadata/cache and audit/security events where needed.

Outbound mail flow:

1. User sends through webmail or SMTP submission.
2. Authentication is validated through Dovecot/control-plane-backed credentials.
3. Postfix applies outbound rate limits, policy, and DKIM signing path.
4. Rspamd can scan outbound mail for compromised-account behavior.
5. Postfix queues and delivers mail to remote MX hosts.
6. Delivery status and abuse signals are logged for admin review.

## Malware Handling

Malware detection happens before normal mailbox delivery. The first version should not silently delete suspected malware. Instead:

- Preserve evidence and audit metadata.
- Place infected or suspicious messages in a restricted user-visible malware/quarantine folder when policy allows.
- Strip or block direct attachment access unless explicitly released by an admin or policy.
- Notify users and/or admins according to tenant policy.
- Keep immutable scan verdicts and scanner version metadata.

False positives must be recoverable by an authorized admin workflow.

## Spam Handling And Learning

Rspamd is the primary spam engine. The platform should support:

- Global, tenant, domain, and user spam thresholds.
- Spam, probable spam, and quarantine actions.
- Mark as spam and mark as not spam from webmail.
- Per-user and per-domain allow/block lists.
- Custom rules through managed Rspamd maps and policy templates.
- Redis-backed Bayesian and neural learning.
- Audit logs for admin-level spam policy changes.

## Storage Model

The first production version should not store full mailboxes primarily in MariaDB.

Primary mail storage:

- Dovecot-managed mailbox storage on filesystem initially.
- A path toward object storage or Dovecot Pro storage can be evaluated later.

MariaDB stores:

- Tenants, domains, users, aliases, credentials, quotas, policies, audit logs.
- Calendar and contact data.
- Mail metadata references only when needed by Go services.

Redis stores:

- Sessions and ephemeral application state.
- Recent mailbox cache for webmail acceleration.
- Recent 200-300 message summaries per active mailbox/folder where useful.
- Message metadata, previews, unread counts, folder summaries, and security verdict summaries.
- Optional short-lived full-message cache with strict size and TTL limits.

Redis should not be the source of truth for mail content.

## Multi-Tenant And Multi-Site Model

Tenant is the top-level customer or organization object. A tenant can own:

- Multiple domains.
- Multiple users and groups.
- Multiple sites.
- Tenant-specific branding.
- Tenant-specific security, spam, quota, and retention policies.

The first implementation can deploy as a single-server system on the development host, but the data model and service boundaries must not prevent future multi-site routing.

Initial deployment target:

- Host: DevMail
- SSH: root@192.168.254.125
- OS: Debian GNU/Linux 13 trixie
- Architecture: x86_64

## Security Baseline

The platform must start with conservative production defaults:

- TLS required for authenticated submission, IMAP, POP3, webadmin, webmail, CalDAV, and CardDAV.
- Strong password hashing for local credentials.
- App-password-ready credential model for mail clients.
- Admin MFA-ready account model.
- Least-privilege Linux users for services.
- Strict file permissions for DKIM keys, mail storage, config files, and secrets.
- No plaintext secret logging.
- Immutable audit events for security-sensitive actions.
- Rate limits for login, submission, SMTP sessions, and APIs.
- Account lockout or risk scoring for repeated failures.
- CSRF protection for browser apps.
- Secure cookie flags.
- Input validation at every public boundary.
- Config rendering with validation before reload.
- Safe service reload and rollback path.

## Autoconfig And Discovery

The platform should provide:

- Thunderbird autoconfig.
- Apple-friendly CalDAV/CardDAV discovery through well-known endpoints.
- Microsoft autodiscover later for ActiveSync.
- DNS helper output for tenant domains.
- MTA-STS and TLS-RPT support.
- DMARC reporting support path.

## First Implementation Slice

The first implementation slice should produce a working single-server foundation:

1. Go monorepo scaffold with separate commands for API, webadmin, worker, config renderer, and service health checks.
2. MariaDB schema and migrations for tenants, domains, users, aliases, credentials, policies, and audit logs.
3. Initial Go admin API.
4. Initial webadmin login and tenant/domain/user creation.
5. Config rendering templates for Postfix, Dovecot, and Rspamd.
6. Deployment scripts for the Debian 13 development host.
7. Baseline Postfix, Dovecot, Rspamd, ClamAV, MariaDB, and Redis integration.
8. End-to-end mail flow for one tenant, one domain, and one user.

Calendar, contacts, CardDAV, CalDAV, webmail, JMAP, and ActiveSync should follow after the mail foundation is working.

## Open Decisions

- Whether the first mailbox storage format should be Maildir or Dovecot sdbox/mdbox.
- Whether to begin with Dovecot Community Edition only or preserve an easy path to Dovecot Pro.
- Which Go web framework and frontend approach to use for webadmin and webmail.
- Whether deployment should use native systemd services first or containers for supporting infrastructure.
- Exact domain used for first live mail tests.

## Acceptance Criteria For The Foundation

- Admin can create a tenant, domain, and user.
- Generated service configs pass validation before reload.
- Postfix receives mail for a configured domain.
- Rspamd and ClamAV scan mail before delivery.
- Dovecot stores mail and exposes IMAP/POP3.
- Authenticated submission can send outbound mail.
- Spam/not-spam actions can train Rspamd.
- Audit logs record security-sensitive admin actions.
- Redis caches recent mailbox summaries without becoming source of truth.
- The deployment can be recreated on Debian 13 from documented steps.
