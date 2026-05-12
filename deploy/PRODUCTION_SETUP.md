# ProIdentity Mail Production Setup

This runbook records the production deployment flow and points to the reusable
bootstrap script.

## Example Server

- Host: `root@203.0.113.10`
- OS: Debian GNU/Linux 13 (trixie)
- Mail hostname used: `mail.example.com`
- Admin hostname used: `madmin.example.com`
- Webmail/DAV hostname used: `webmail.example.com`
- Initial TLS mode: `none`

`TLS_MODE=none` is useful for first boot, lab installs, or when TLS is
terminated by an upstream proxy. Direct public deployments can use
`letsencrypt-http` when DNS points straight to this server and ports 80/443 are
reachable. Deployments behind Cloudflare proxy or another front proxy should use
`letsencrypt-dns-cloudflare`, `behind-proxy`, or `custom-cert` depending on who
owns the public certificate.

## TLS Modes

- `none`: HTTP only. Good for first boot, internal testing, or temporary setup.
- `behind-proxy`: HTTPS is handled by an upstream proxy. Internal Nginx serves
  the app and trusts configured proxy headers when enabled.
- `letsencrypt-http`: uses Let's Encrypt HTTP-01 through local Nginx. Use this
  when public A/AAAA records resolve directly to the server public IP and port
  80 reaches this server.
- `letsencrypt-dns-cloudflare`: uses Let's Encrypt DNS-01 with a Cloudflare API
  token. Use this when Cloudflare proxy is enabled, the server is behind another
  proxy, or HTTP-01 cannot reliably reach the origin.
- `custom-cert`: installs an existing certificate/key pair and applies it to
  web HTTPS plus Postfix/Dovecot TLS inventory.

## Recommended: Install From Binary Release

Use this path for normal server setup. The target server downloads the published
release archive, verifies checksums when available, and runs the same full
bootstrap script without compiling Go code on the server.
The release wrapper itself only downloads and extracts the archive; the
extracted `proidentity-production-setup.sh` performs the package installation
and full server configuration.

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
  --tls-mode none
```

For more release installer options, including GitLab and direct archive URLs,
see [BINARY_RELEASE_INSTALL.md](BINARY_RELEASE_INSTALL.md).

## Alternative: Build Local Artifacts

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

## Upload Local Artifacts

```powershell
ssh -o BatchMode=yes root@203.0.113.10 "rm -rf /tmp/proidentity-release; mkdir -p /tmp/proidentity-release"
scp -o BatchMode=yes .\bin\webadmin-linux-amd64 .\bin\webmail-linux-amd64 .\bin\groupware-linux-amd64 .\bin\mailctl-linux-amd64 .\deploy\devmail\apply-mail-config.sh .\deploy\proidentity-production-setup.sh root@203.0.113.10:/tmp/proidentity-release/
ssh -o BatchMode=yes root@203.0.113.10 "cd /tmp/proidentity-release; cp webadmin-linux-amd64 webadmin; cp webmail-linux-amd64 webmail; cp groupware-linux-amd64 groupware; cp mailctl-linux-amd64 mailctl; cp apply-mail-config.sh apply-mail-config; chmod +x webadmin webmail groupware mailctl apply-mail-config proidentity-production-setup.sh; bash -n proidentity-production-setup.sh"
```

## Run Bootstrap Script With Local Artifacts

```bash
bash /tmp/proidentity-release/proidentity-production-setup.sh \
  --artifact-dir /tmp/proidentity-release \
  --public-ipv4 203.0.113.10 \
  --mail-hostname mail.example.com \
  --admin-hostname madmin.example.com \
  --webmail-hostname webmail.example.com \
  --dav-hostname webmail.example.com \
  --autoconfig-hostname autoconfig.example.com \
  --autodiscover-hostname autodiscover.example.com \
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
curl -H 'Host: madmin.example.com' http://127.0.0.1/
curl -H 'Host: webmail.example.com' http://127.0.0.1/
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
