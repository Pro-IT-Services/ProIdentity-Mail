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

mkdir -p /var/vmail /etc/postfix/proidentity /etc/postfix/proidentity/tls-sni /etc/proidentity-mail/backups
mkdir -p /var/lib/rspamd/dkim
chown -R vmail:vmail /var/vmail
chmod 0750 /var/vmail
chown _rspamd:_rspamd /var/lib/rspamd/dkim
chmod 0750 /var/lib/rspamd/dkim

sql_escape() {
  printf "%s" "$1" | sed "s/'/''/g"
}

while IFS=$'\t' read -r domain_id domain_name selector; do
  [[ -n "${domain_id}" && -n "${domain_name}" && -n "${selector}" ]] || continue
  key_path="/var/lib/rspamd/dkim/${domain_name}.${selector}.key"
  if [[ ! -f "${key_path}" ]]; then
    dns_txt="$(rspamadm dkim_keygen -d "${domain_name}" -s "${selector}" -b 2048 -k "${key_path}" -o dns)"
    chown _rspamd:_rspamd "${key_path}"
    chmod 0640 "${key_path}"
    escaped_dns="$(sql_escape "${dns_txt}")"
    escaped_key="$(sql_escape "${key_path}")"
    escaped_selector="$(sql_escape "${selector}")"
    mariadb -D "${PROIDENTITY_DB_NAME}" -e "INSERT INTO dkim_keys(domain_id, selector, key_path, public_dns_txt, status) VALUES (${domain_id}, '${escaped_selector}', '${escaped_key}', '${escaped_dns}', 'active') ON DUPLICATE KEY UPDATE key_path=VALUES(key_path), public_dns_txt=VALUES(public_dns_txt), status='active'"
  fi
done < <(mariadb --batch --skip-column-names -D "${PROIDENTITY_DB_NAME}" -e "SELECT d.id, d.name, d.dkim_selector FROM domains d LEFT JOIN dkim_keys k ON k.domain_id = d.id AND k.selector = d.dkim_selector AND k.status = 'active' WHERE d.status IN ('pending','active') AND k.id IS NULL")

/opt/proidentity-mail/bin/mailctl render

backup_dir="/etc/proidentity-mail/backups/$(date +%Y%m%d-%H%M%S)"
mkdir -p "${backup_dir}"
cp -a /etc/postfix/main.cf "${backup_dir}/postfix-main.cf" 2>/dev/null || true
cp -a /etc/postfix/master.cf "${backup_dir}/postfix-master.cf" 2>/dev/null || true
cp -a /etc/postfix/proidentity/tls-sni-map "${backup_dir}/postfix-tls-sni-map" 2>/dev/null || true
cp -a /etc/dovecot/conf.d/99-proidentity.conf "${backup_dir}/dovecot-99-proidentity.conf" 2>/dev/null || true
cp -a /etc/dovecot/proidentity-sql.conf.ext "${backup_dir}/dovecot-proidentity-sql.conf.ext" 2>/dev/null || true
cp -a /etc/rspamd/local.d/settings.conf "${backup_dir}/rspamd-settings.conf" 2>/dev/null || true
cp -a /etc/rspamd/local.d/force_actions.conf "${backup_dir}/rspamd-force_actions.conf" 2>/dev/null || true

install -m 0644 /etc/proidentity-mail/generated/postfix-main.cf /etc/postfix/main.cf
install -m 0644 /etc/proidentity-mail/generated/postfix-master.cf /etc/postfix/master.cf
install -m 0640 -o root -g postfix /etc/proidentity-mail/generated/virtual-mailbox-domains.cf /etc/postfix/proidentity/virtual-mailbox-domains.cf
install -m 0640 -o root -g postfix /etc/proidentity-mail/generated/virtual-mailbox-maps.cf /etc/postfix/proidentity/virtual-mailbox-maps.cf
install -m 0640 -o root -g postfix /etc/proidentity-mail/generated/virtual-alias-maps.cf /etc/postfix/proidentity/virtual-alias-maps.cf
install -m 0640 -o root -g postfix /etc/proidentity-mail/generated/sender-login-maps.cf /etc/postfix/proidentity/sender-login-maps.cf

sni_map="/etc/postfix/proidentity/tls-sni-map"
: > "${sni_map}"
chown root:postfix "${sni_map}"
chmod 0640 "${sni_map}"
if [[ -f /etc/proidentity-mail/generated/tls-sni-source-map ]]; then
  while IFS=$'\t' read -r hostname cert_file key_file chain_file; do
    [[ -n "${hostname}" && -n "${cert_file}" && -n "${key_file}" && -n "${chain_file}" ]] || continue
    if [[ -s "${cert_file}" && -s "${key_file}" ]]; then
      install -d -m 0700 "$(dirname "${chain_file}")"
      cat "${key_file}" "${cert_file}" > "${chain_file}"
      chmod 0600 "${chain_file}"
      printf "%s\t%s\n" "${hostname}" "${chain_file}" >> "${sni_map}"
    else
      echo "warning: skipping SNI host ${hostname}; missing cert/key" >&2
    fi
  done < /etc/proidentity-mail/generated/tls-sni-source-map
fi
postmap -F "hash:${sni_map}"
chown root:postfix "${sni_map}" "${sni_map}.db" 2>/dev/null || true
chmod 0640 "${sni_map}" "${sni_map}.db" 2>/dev/null || true

install -m 0644 /etc/proidentity-mail/generated/dovecot-proidentity.conf /etc/dovecot/conf.d/99-proidentity.conf
install -m 0640 -o root -g dovecot /etc/proidentity-mail/generated/dovecot-sql.conf.ext /etc/dovecot/proidentity-sql.conf.ext
sed -i 's/^!include auth-system.conf.ext/#!include auth-system.conf.ext/' /etc/dovecot/conf.d/10-auth.conf

mkdir -p /etc/rspamd/local.d
install -m 0644 /etc/proidentity-mail/generated/rspamd-local.d-redis.conf /etc/rspamd/local.d/redis.conf
install -m 0644 /etc/proidentity-mail/generated/rspamd-local.d-antivirus.conf /etc/rspamd/local.d/antivirus.conf
install -m 0644 /etc/proidentity-mail/generated/rspamd-local.d-dkim_signing.conf /etc/rspamd/local.d/dkim_signing.conf
install -m 0644 /etc/proidentity-mail/generated/rspamd-local.d-actions.conf /etc/rspamd/local.d/actions.conf
install -m 0644 /etc/proidentity-mail/generated/rspamd-local.d-milter_headers.conf /etc/rspamd/local.d/milter_headers.conf
install -m 0644 /etc/proidentity-mail/generated/rspamd-local.d-settings.conf /etc/rspamd/local.d/settings.conf
install -m 0644 /etc/proidentity-mail/generated/rspamd-local.d-force_actions.conf /etc/rspamd/local.d/force_actions.conf
find /etc/rspamd/local.d -maxdepth 1 -type f -name '*.conf' -exec chmod 0644 {} \;

postfix check
doveconf >/dev/null
rspamadm configtest

systemctl restart postfix dovecot rspamd
systemctl reload clamav-daemon || systemctl restart clamav-daemon || true

echo "mail config applied; backup: ${backup_dir}"
