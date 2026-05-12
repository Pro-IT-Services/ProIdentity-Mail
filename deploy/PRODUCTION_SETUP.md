# ProIdentity Mail Production Setup

This runbook records the production deployment flow and points to the reusable
bootstrap script.

## Server Used

- Host: `root@46.225.87.251`
- OS: Debian GNU/Linux 13 (trixie)
- Mail hostname used: `mail.proidentity.cloud`
- Admin hostname used: `madmin.proidentity.cloud`
- Webmail/DAV hostname used: `webmail.proidentity.cloud`
- Initial TLS mode: `none`

`TLS_MODE=none` was used because the visible DNS records for the ProIdentity
hostnames were Cloudflare proxy addresses, not direct `46.225.87.251` A records.
The server is ready to switch later to `letsencrypt-dns-cloudflare` or
`custom-cert` after DNS/API credentials are configured.

## Build Steps

From the project worktree:

Use Go `1.26.3` or newer for release builds. Earlier `1.26.2` builds include
reachable standard-library vulnerabilities in `net/mail`, `net`, and HTTP/2.

```powershell
$env:GOCACHE=(Join-Path (Get-Location) '.gocache')
go version
go test -buildvcs=false ./...
powershell -ExecutionPolicy Bypass -File .\scripts\security-check.ps1
$env:GOOS='linux'
$env:GOARCH='amd64'
go build -buildvcs=false -o .\bin\webadmin-linux-amd64 .\cmd\webadmin
go build -buildvcs=false -o .\bin\webmail-linux-amd64 .\cmd\webmail
go build -buildvcs=false -o .\bin\groupware-linux-amd64 .\cmd\groupware
go build -buildvcs=false -o .\bin\mailctl-linux-amd64 .\cmd\mailctl
Remove-Item Env:\GOCACHE,Env:\GOOS,Env:\GOARCH -ErrorAction SilentlyContinue
```

## Upload Steps

```powershell
ssh -o BatchMode=yes root@46.225.87.251 "rm -rf /tmp/proidentity-release; mkdir -p /tmp/proidentity-release"
scp -o BatchMode=yes .\bin\webadmin-linux-amd64 .\bin\webmail-linux-amd64 .\bin\groupware-linux-amd64 .\bin\mailctl-linux-amd64 .\deploy\devmail\apply-mail-config.sh .\deploy\proidentity-production-setup.sh root@46.225.87.251:/tmp/proidentity-release/
ssh -o BatchMode=yes root@46.225.87.251 "cd /tmp/proidentity-release; cp webadmin-linux-amd64 webadmin; cp webmail-linux-amd64 webmail; cp groupware-linux-amd64 groupware; cp mailctl-linux-amd64 mailctl; cp apply-mail-config.sh apply-mail-config; chmod +x webadmin webmail groupware mailctl apply-mail-config proidentity-production-setup.sh; bash -n proidentity-production-setup.sh"
```

## Install Command Used

```bash
bash /tmp/proidentity-release/proidentity-production-setup.sh \
  --artifact-dir /tmp/proidentity-release \
  --public-ipv4 46.225.87.251 \
  --mail-hostname mail.proidentity.cloud \
  --admin-hostname madmin.proidentity.cloud \
  --webmail-hostname webmail.proidentity.cloud \
  --dav-hostname webmail.proidentity.cloud \
  --autoconfig-hostname autoconfig.proidentity.cloud \
  --autodiscover-hostname autodiscover.proidentity.cloud \
  --tls-mode none
```

The script installs OS packages, creates service users, generates missing
secrets, creates the MariaDB database/user, installs binaries, writes systemd
units, runs migrations, creates missing DKIM keys before rendering Rspamd
signing config, renders Postfix/Dovecot/Rspamd/Nginx config, syncs TLS
certificate inventory, starts services, and writes a root-only summary.

Any omitted password, token, nonce, or backup key is generated with OpenSSL and
stored on the server in:

```text
/root/proidentity-install-summary.txt
```

## Verification Commands

```bash
systemctl is-active mariadb redis-server postfix dovecot rspamd clamav-daemon nginx proidentity-webadmin proidentity-webmail proidentity-groupware proidentity-backup.timer proidentity-tls-worker.timer
curl -sS http://127.0.0.1:8080/healthz
curl -sS http://127.0.0.1:8081/healthz
curl -sS http://127.0.0.1:8082/healthz
curl -H 'Host: madmin.proidentity.cloud' http://127.0.0.1/
curl -H 'Host: webmail.proidentity.cloud' http://127.0.0.1/
/opt/proidentity-mail/bin/mailctl sync-tls-inventory
ss -ltnp
```

## Reusable Bootstrap Script

Use:

```text
deploy/proidentity-production-setup.sh
```

Important environment overrides:

- `ARTIFACT_DIR` or `--artifact-dir`: directory containing `webadmin`, `webmail`, `groupware`, `mailctl`, and `apply-mail-config`
- `PUBLIC_IPV4`, `PUBLIC_IPV6`
- `MAIL_HOSTNAME`, `ADMIN_HOSTNAME`, `WEBMAIL_HOSTNAME`, `DAV_HOSTNAME`
- `AUTOCONFIG_HOSTNAME`, `AUTODISCOVER_HOSTNAME`
- `TLS_MODE`: `none`, `behind-proxy`, `letsencrypt-http`, `letsencrypt-dns-cloudflare`, or `custom-cert`
- `ADMIN_USERNAME`, `ADMIN_PASSWORD`
- `DB_NAME`, `DB_USER`, `DB_PASSWORD`
- `CLOUDFLARE_CERT_DOMAIN`, `CLOUDFLARE_CREDENTIALS_FILE`
- `TLS_CERT_PATH`, `TLS_KEY_PATH`

If passwords/secrets are not supplied, the script generates them with OpenSSL
and records them in `/root/proidentity-install-summary.txt`.

Equivalent command-line options are supported for the same values:

```bash
bash deploy/proidentity-production-setup.sh --help
```
