#!/usr/bin/env bash
set -euo pipefail

if ! command -v sudo >/dev/null 2>&1 && command -v apt-get >/dev/null 2>&1; then
  apt-get update
  DEBIAN_FRONTEND=noninteractive apt-get install -y sudo
fi

id proidentity >/dev/null 2>&1 || useradd --system --home /opt/proidentity-mail --shell /usr/sbin/nologin proidentity

mkdir -p /etc/proidentity-mail/generated /opt/proidentity-mail/bin /var/backups/proidentity-mail
install -m 0755 /tmp/webadmin /opt/proidentity-mail/bin/webadmin
install -m 0755 /tmp/mailctl /opt/proidentity-mail/bin/mailctl
install -m 0755 /tmp/groupware /opt/proidentity-mail/bin/groupware
install -m 0755 /tmp/webmail /opt/proidentity-mail/bin/webmail
install -m 0644 /tmp/proidentity-devmail/proidentity-webadmin.service /etc/systemd/system/proidentity-webadmin.service
install -m 0644 /tmp/proidentity-devmail/proidentity-groupware.service /etc/systemd/system/proidentity-groupware.service
install -m 0644 /tmp/proidentity-devmail/proidentity-webmail.service /etc/systemd/system/proidentity-webmail.service
install -m 0644 /tmp/proidentity-devmail/proidentity-mailctl.service /etc/systemd/system/proidentity-mailctl.service
install -m 0644 /tmp/proidentity-devmail/proidentity-backup.service /etc/systemd/system/proidentity-backup.service
install -m 0644 /tmp/proidentity-devmail/proidentity-backup.timer /etc/systemd/system/proidentity-backup.timer
install -m 0644 /tmp/proidentity-devmail/proidentity-config-apply.service /etc/systemd/system/proidentity-config-apply.service
install -m 0644 /tmp/proidentity-devmail/proidentity-config-apply.path /etc/systemd/system/proidentity-config-apply.path
install -m 0755 /tmp/proidentity-devmail/apply-mail-config.sh /opt/proidentity-mail/bin/apply-mail-config
install -m 0750 /tmp/proidentity-devmail/proidentity-rootctl /opt/proidentity-mail/bin/proidentity-rootctl

if [[ ! -f /etc/proidentity-mail/proidentity-mail.env ]]; then
  db_password="$(openssl rand -hex 24)"
  mariadb <<SQL
CREATE DATABASE IF NOT EXISTS proidentity_mail CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'proidentity_mail'@'127.0.0.1' IDENTIFIED BY '${db_password}';
ALTER USER 'proidentity_mail'@'127.0.0.1' IDENTIFIED BY '${db_password}';
GRANT ALL PRIVILEGES ON proidentity_mail.* TO 'proidentity_mail'@'127.0.0.1';
FLUSH PRIVILEGES;
SQL

  cat > /etc/proidentity-mail/proidentity-mail.env <<EOF
PROIDENTITY_HTTP_ADDR=127.0.0.1:8080
PROIDENTITY_GROUPWARE_ADDR=127.0.0.1:8081
PROIDENTITY_WEBMAIL_ADDR=127.0.0.1:8082
PROIDENTITY_DB_NAME=proidentity_mail
PROIDENTITY_DB_USER=proidentity_mail
PROIDENTITY_DB_PASSWORD=${db_password}
PROIDENTITY_DB_DSN='proidentity_mail:${db_password}@tcp(127.0.0.1:3306)/proidentity_mail?parseTime=true&multiStatements=true'
PROIDENTITY_BACKUP_ENCRYPTION_KEY=$(openssl rand -hex 32)
PROIDENTITY_AUTH_POLICY_TOKEN=$(openssl rand -hex 32)
PROIDENTITY_AUTH_POLICY_NONCE=$(openssl rand -hex 32)
PROIDENTITY_CONFIG_DIR=/etc/proidentity-mail/generated
PROIDENTITY_MAILCTL_PATH=/opt/proidentity-mail/bin/mailctl
PROIDENTITY_CONFIG_APPLY_REQUEST_PATH=/etc/proidentity-mail/apply-request
PROIDENTITY_MAIL_HOSTNAME=mail.local
PROIDENTITY_ADMIN_USERNAME=admin
PROIDENTITY_ADMIN_PASSWORD=$(openssl rand -hex 18)
EOF
  chmod 0640 /etc/proidentity-mail/proidentity-mail.env
fi

if ! grep -q '^PROIDENTITY_GROUPWARE_ADDR=' /etc/proidentity-mail/proidentity-mail.env; then
  printf '\nPROIDENTITY_GROUPWARE_ADDR=127.0.0.1:8081\n' >> /etc/proidentity-mail/proidentity-mail.env
fi
if ! grep -q '^PROIDENTITY_WEBMAIL_ADDR=' /etc/proidentity-mail/proidentity-mail.env; then
  printf '\nPROIDENTITY_WEBMAIL_ADDR=127.0.0.1:8082\n' >> /etc/proidentity-mail/proidentity-mail.env
fi
if ! grep -q '^PROIDENTITY_ADMIN_USERNAME=' /etc/proidentity-mail/proidentity-mail.env; then
  printf '\nPROIDENTITY_ADMIN_USERNAME=admin\n' >> /etc/proidentity-mail/proidentity-mail.env
fi
if ! grep -q '^PROIDENTITY_ADMIN_PASSWORD=' /etc/proidentity-mail/proidentity-mail.env; then
  printf '\nPROIDENTITY_ADMIN_PASSWORD=%s\n' "$(openssl rand -hex 18)" >> /etc/proidentity-mail/proidentity-mail.env
fi
if ! grep -q '^PROIDENTITY_BACKUP_ENCRYPTION_KEY=' /etc/proidentity-mail/proidentity-mail.env; then
  printf '\nPROIDENTITY_BACKUP_ENCRYPTION_KEY=%s\n' "$(openssl rand -hex 32)" >> /etc/proidentity-mail/proidentity-mail.env
fi
if ! grep -q '^PROIDENTITY_AUTH_POLICY_TOKEN=' /etc/proidentity-mail/proidentity-mail.env; then
  printf '\nPROIDENTITY_AUTH_POLICY_TOKEN=%s\n' "$(openssl rand -hex 32)" >> /etc/proidentity-mail/proidentity-mail.env
fi
if ! grep -q '^PROIDENTITY_AUTH_POLICY_NONCE=' /etc/proidentity-mail/proidentity-mail.env; then
  printf '\nPROIDENTITY_AUTH_POLICY_NONCE=%s\n' "$(openssl rand -hex 32)" >> /etc/proidentity-mail/proidentity-mail.env
fi
if ! grep -q '^PROIDENTITY_MAILCTL_PATH=' /etc/proidentity-mail/proidentity-mail.env; then
  printf '\nPROIDENTITY_MAILCTL_PATH=/opt/proidentity-mail/bin/mailctl\n' >> /etc/proidentity-mail/proidentity-mail.env
fi
if ! grep -q '^PROIDENTITY_CONFIG_APPLY_REQUEST_PATH=' /etc/proidentity-mail/proidentity-mail.env; then
  printf '\nPROIDENTITY_CONFIG_APPLY_REQUEST_PATH=/etc/proidentity-mail/apply-request\n' >> /etc/proidentity-mail/proidentity-mail.env
fi

usermod -a -G vmail proidentity || true
usermod -a -G postfix proidentity || true
usermod -a -G dovecot proidentity || true
chown -R proidentity:proidentity /etc/proidentity-mail /opt/proidentity-mail
chown root:proidentity /opt/proidentity-mail/bin/proidentity-rootctl
chown root:root /var/backups/proidentity-mail
chmod 0750 /etc/proidentity-mail /etc/proidentity-mail/generated
chmod 0755 /opt/proidentity-mail /opt/proidentity-mail/bin
chmod 0750 /opt/proidentity-mail/bin/proidentity-rootctl
chmod 0750 /var/backups/proidentity-mail
cat > /etc/sudoers.d/proidentity-mail-rootctl <<'EOF'
Defaults:proidentity env_keep += "PROIDENTITY_*"
proidentity ALL=(root) NOPASSWD: /opt/proidentity-mail/bin/proidentity-rootctl backup
proidentity ALL=(root) NOPASSWD: /opt/proidentity-mail/bin/proidentity-rootctl tls-worker
proidentity ALL=(root) NOPASSWD: /opt/proidentity-mail/bin/proidentity-rootctl config-apply
proidentity ALL=(root) NOPASSWD: /opt/proidentity-mail/bin/proidentity-rootctl sync-proxy
EOF
chmod 0440 /etc/sudoers.d/proidentity-mail-rootctl
visudo -cf /etc/sudoers.d/proidentity-mail-rootctl >/dev/null

systemctl daemon-reload
systemctl enable --now proidentity-backup.timer proidentity-config-apply.path
echo "runtime setup complete"
