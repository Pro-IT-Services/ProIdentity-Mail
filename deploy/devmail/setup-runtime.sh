#!/usr/bin/env bash
set -euo pipefail

id proidentity >/dev/null 2>&1 || useradd --system --home /opt/proidentity-mail --shell /usr/sbin/nologin proidentity

mkdir -p /etc/proidentity-mail/generated /opt/proidentity-mail/bin
install -m 0755 /tmp/webadmin /opt/proidentity-mail/bin/webadmin
install -m 0755 /tmp/mailctl /opt/proidentity-mail/bin/mailctl
install -m 0644 /tmp/proidentity-devmail/proidentity-webadmin.service /etc/systemd/system/proidentity-webadmin.service
install -m 0644 /tmp/proidentity-devmail/proidentity-mailctl.service /etc/systemd/system/proidentity-mailctl.service

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
PROIDENTITY_HTTP_ADDR=0.0.0.0:8080
PROIDENTITY_DB_NAME=proidentity_mail
PROIDENTITY_DB_USER=proidentity_mail
PROIDENTITY_DB_PASSWORD=${db_password}
PROIDENTITY_DB_DSN='proidentity_mail:${db_password}@tcp(127.0.0.1:3306)/proidentity_mail?parseTime=true&multiStatements=true'
PROIDENTITY_CONFIG_DIR=/etc/proidentity-mail/generated
PROIDENTITY_MAIL_HOSTNAME=mail.local
EOF
  chmod 0640 /etc/proidentity-mail/proidentity-mail.env
fi

chown -R proidentity:proidentity /etc/proidentity-mail /opt/proidentity-mail
chmod 0750 /etc/proidentity-mail /etc/proidentity-mail/generated /opt/proidentity-mail /opt/proidentity-mail/bin

systemctl daemon-reload
echo "runtime setup complete"
