#!/usr/bin/env bash
set -euo pipefail

if [[ ! -f /etc/proidentity-mail/proidentity-mail.env ]]; then
  echo "/etc/proidentity-mail/proidentity-mail.env missing" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1091
. /etc/proidentity-mail/proidentity-mail.env
set +a

id vmail >/dev/null 2>&1 || useradd --system --uid 5000 --gid 5000 --home /var/vmail --shell /usr/sbin/nologin vmail 2>/dev/null || true
if ! getent group vmail >/dev/null; then
  groupadd --system --gid 5000 vmail
fi
if ! id vmail >/dev/null 2>&1; then
  useradd --system --uid 5000 --gid vmail --home /var/vmail --shell /usr/sbin/nologin vmail
fi

mkdir -p /var/vmail /etc/postfix/proidentity /etc/proidentity-mail/backups
chown -R vmail:vmail /var/vmail
chmod 0750 /var/vmail

/opt/proidentity-mail/bin/mailctl render

backup_dir="/etc/proidentity-mail/backups/$(date +%Y%m%d-%H%M%S)"
mkdir -p "${backup_dir}"
cp -a /etc/postfix/main.cf "${backup_dir}/postfix-main.cf" 2>/dev/null || true
cp -a /etc/postfix/master.cf "${backup_dir}/postfix-master.cf" 2>/dev/null || true
cp -a /etc/dovecot/conf.d/99-proidentity.conf "${backup_dir}/dovecot-99-proidentity.conf" 2>/dev/null || true
cp -a /etc/dovecot/proidentity-sql.conf.ext "${backup_dir}/dovecot-proidentity-sql.conf.ext" 2>/dev/null || true

install -m 0644 /etc/proidentity-mail/generated/postfix-main.cf /etc/postfix/main.cf
install -m 0644 /etc/proidentity-mail/generated/postfix-master.cf /etc/postfix/master.cf
install -m 0640 -o root -g postfix /etc/proidentity-mail/generated/virtual-mailbox-domains.cf /etc/postfix/proidentity/virtual-mailbox-domains.cf
install -m 0640 -o root -g postfix /etc/proidentity-mail/generated/virtual-mailbox-maps.cf /etc/postfix/proidentity/virtual-mailbox-maps.cf
install -m 0640 -o root -g postfix /etc/proidentity-mail/generated/virtual-alias-maps.cf /etc/postfix/proidentity/virtual-alias-maps.cf

install -m 0644 /etc/proidentity-mail/generated/dovecot-proidentity.conf /etc/dovecot/conf.d/99-proidentity.conf
install -m 0640 -o root -g dovecot /etc/proidentity-mail/generated/dovecot-sql.conf.ext /etc/dovecot/proidentity-sql.conf.ext
sed -i 's/^!include auth-system.conf.ext/#!include auth-system.conf.ext/' /etc/dovecot/conf.d/10-auth.conf

mkdir -p /etc/rspamd/local.d
install -m 0644 /etc/proidentity-mail/generated/rspamd-local.d-redis.conf /etc/rspamd/local.d/redis.conf

postfix check
doveconf >/dev/null
rspamadm configtest

systemctl restart postfix dovecot rspamd
systemctl reload clamav-daemon || systemctl restart clamav-daemon || true

echo "mail config applied; backup: ${backup_dir}"
