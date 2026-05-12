# DevMail Deployment

Target:

- Host: `root@192.168.254.125`
- OS: Debian GNU/Linux 13 trixie

## Package Install

```bash
cd /opt/proidentity-mail
bash deploy/devmail/install-packages.sh
```

## Runtime User

```bash
bash deploy/devmail/setup-runtime.sh
```

## Database

```sql
CREATE DATABASE proidentity_mail CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'proidentity_mail'@'127.0.0.1' IDENTIFIED BY 'change-me';
GRANT ALL PRIVILEGES ON proidentity_mail.* TO 'proidentity_mail'@'127.0.0.1';
FLUSH PRIVILEGES;
```

## Services

Copy `proidentity-mail.env.example` to `/etc/proidentity-mail/proidentity-mail.env`, edit secrets, install units into `/etc/systemd/system`, then run:

```bash
systemctl daemon-reload
systemctl enable --now proidentity-webadmin proidentity-groupware
systemctl enable --now proidentity-backup.timer
systemctl start proidentity-mailctl
```

## Backups

The backup timer runs daily at 02:15 with a randomized delay. Defaults keep 7 daily, 4 weekly, and 12 monthly backup archives.

```bash
/opt/proidentity-mail/bin/mailctl backup --output-dir /var/backups/proidentity-mail --prune-after
/opt/proidentity-mail/bin/mailctl backup-prune --dir /var/backups/proidentity-mail
```

## Apply Mail Daemon Config

```bash
/opt/proidentity-mail/bin/apply-mail-config
```
