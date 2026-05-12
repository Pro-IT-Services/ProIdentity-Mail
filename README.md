# ProIdentity Mail

ProIdentity Mail is a production-first, multi-tenant mail and groupware platform
with a Go control plane and proven mail components underneath.

It is designed for operators who want one platform for mail hosting, webmail,
admin management, DNS automation, TLS certificates, spam filtering, antivirus,
CalDAV/CardDAV, and security policy.

## Current Status

This project is in active pre-production development. It is suitable for lab,
test, and controlled pilot environments while the security hardening and release
process are finalized.

## Main Components

- Postfix for SMTP, submission, relay policy, and outbound delivery.
- Dovecot for IMAP, POP3, LMTP, Sieve, and ManageSieve.
- Rspamd for spam filtering, DKIM/SPF/DMARC checks, and spam training.
- ClamAV for malware scanning before delivery.
- MariaDB for tenant, domain, mailbox, policy, audit, and groupware state.
- Redis for caching, sessions, and Rspamd state.
- Go services for web admin, webmail, groupware/DAV, automation, drift checks,
  backup orchestration, and secure control-plane operations.

## Features

- Multi-tenant domains, users, aliases, shared mailboxes, catch-all mailboxes,
  quotas, permissions, and tenant-scoped administration.
- Webmail with folders, drag and drop, spam/not-spam training, drafts,
  attachments, signatures, trusted sender controls, calendar, and contacts.
- Admin panel for tenants, domains, DNS records, Cloudflare automation,
  certificates, security policy, quarantine, audit logs, drift checks, backups,
  service settings, and 2FA policy.
- CalDAV/CardDAV endpoints for phones and desktop clients.
- Autoconfig and autodiscover endpoints for mail client setup.
- Built-in rate limiting, lockout, audit trails, session security, CSRF
  protection, secure cookies, 2FA, app passwords, and admin step-up checks.
- Binary release installer for servers that should not compile from source.

## Binary Install from GitHub Release

Download the installer script on a fresh Debian server and point it at a GitHub
release. The installer downloads the correct binary archive, verifies checksums
when available, and runs the full setup script.

```bash
curl -fsSL https://github.com/Pro-IT-Services/ProIdentity-Mail/raw/main/deploy/install-from-release.sh \
  -o /tmp/install-proidentity-mail.sh
chmod +x /tmp/install-proidentity-mail.sh

sudo /tmp/install-proidentity-mail.sh \
  --github-repo Pro-IT-Services/ProIdentity-Mail \
  --version v0.1.1 \
  -- \
  --public-ipv4 203.0.113.10 \
  --mail-hostname mail.example.com \
  --admin-hostname madmin.example.com \
  --webmail-hostname webmail.example.com \
  --dav-hostname webmail.example.com \
  --autoconfig-hostname autoconfig.example.com \
  --autodiscover-hostname autodiscover.example.com \
  --tls-mode letsencrypt-dns-cloudflare
```

All omitted passwords, database credentials, auth tokens, nonces, and backup
encryption keys are generated automatically and written to the root-only
install summary on the server.

TLS can be enabled during install with `--tls-mode letsencrypt-http` when the
public DNS records point directly to the server and port 80 is reachable. Use
`--tls-mode letsencrypt-dns-cloudflare` for Cloudflare DNS-01 automation, or
`--tls-mode custom-cert` to install an existing certificate and key.

You can also install from a direct archive URL:

```bash
sudo /tmp/install-proidentity-mail.sh \
  --release-url https://example.com/proidentity-mail_v0.1.1_linux_x64.tar.gz \
  --checksum-url https://example.com/SHA256SUMS \
  -- \
  --mail-hostname mail.example.com \
  --admin-hostname madmin.example.com \
  --webmail-hostname webmail.example.com
```

## Building Release Artifacts

Release artifacts are created by GitHub Actions when a tag such as `v0.1.1` is
pushed. Locally, you can create the same archives with:

```bash
bash scripts/build-release.sh v0.1.1
```

This produces Linux archives for:

- `x64` (`linux/amd64`)
- `x86` (`linux/386`)
- `arm` (`linux/arm/v7`)
- `arm64` (`linux/arm64`)

Archive names use this format:

- `proidentity-mail_<version>_linux_x64.tar.gz`
- `proidentity-mail_<version>_linux_x86.tar.gz`
- `proidentity-mail_<version>_linux_arm.tar.gz`
- `proidentity-mail_<version>_linux_arm64.tar.gz`

Each archive contains:

- `webadmin`
- `webmail`
- `groupware`
- `mailctl`
- `apply-mail-config`
- `proidentity-production-setup.sh`
- license and release metadata

## Development Commands

```bash
go test ./...
```

On Windows:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\security-check.ps1
```

## License

ProIdentity Mail uses a dual licensing model.

Self-hosted, internal, development, testing, and non-commercial community use is
available under the community grant described in `LICENSE`, based on AGPLv3
network source-sharing obligations.

Commercial hosted service use requires a separate commercial license. This
includes paid SaaS, managed mail hosting, MSP/reseller offerings, cloud
marketplace images, white-label hosting, and commercial third-party email
platform services.

This license model is source-available, not OSI-approved open source, because it
reserves commercial hosted service use for a paid license.

See `LICENSE` and `COMMERCIAL-LICENSE.md`.
