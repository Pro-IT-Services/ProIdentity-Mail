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
systemctl enable --now proidentity-webadmin
systemctl start proidentity-mailctl
```
