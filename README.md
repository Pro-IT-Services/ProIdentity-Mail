# ProIdentity Mail

Production-first multi-tenant mail and groupware platform.

The first foundation uses Go for the control plane and proven mail daemons for protocol-heavy services:

- Postfix for SMTP and submission
- Dovecot for IMAP, POP3, LMTP, Sieve, and ManageSieve
- Rspamd for spam filtering and mail authentication checks
- ClamAV for malware scanning
- MariaDB for platform state
- Redis for cache, sessions, and Rspamd state

## First Local Commands

```powershell
go test ./...
go run ./cmd/mailctl
go run ./cmd/webadmin
```

## Development Target

Initial deployment target:

- Host: DevMail
- SSH: root@192.168.254.125
- OS: Debian GNU/Linux 13 trixie
