package render

const postfixMainTemplate = `
myhostname = {{ .Hostname }}
myorigin = $myhostname
compatibility_level = 3.6
inet_interfaces = all
inet_protocols = ipv4
smtpd_tls_security_level = may
smtp_tls_security_level = may
smtpd_tls_cert_file = /etc/ssl/certs/ssl-cert-snakeoil.pem
smtpd_tls_key_file = /etc/ssl/private/ssl-cert-snakeoil.key
smtpd_tls_auth_only = yes
smtpd_sasl_auth_enable = yes
smtpd_sasl_type = dovecot
smtpd_sasl_path = private/auth
smtpd_relay_restrictions = permit_mynetworks,permit_sasl_authenticated,defer_unauth_destination
smtpd_recipient_restrictions = permit_mynetworks,reject_unauth_destination
virtual_transport = lmtp:unix:private/dovecot-lmtp
virtual_mailbox_domains = mysql:/etc/postfix/proidentity/virtual-mailbox-domains.cf
virtual_mailbox_maps = mysql:/etc/postfix/proidentity/virtual-mailbox-maps.cf
virtual_alias_maps = mysql:/etc/postfix/proidentity/virtual-alias-maps.cf
smtpd_milters = inet:127.0.0.1:11332
non_smtpd_milters = inet:127.0.0.1:11332
milter_protocol = 6
milter_default_action = tempfail
milter_mail_macros = i {auth_type} {auth_authen} {auth_author} {mail_addr}
`

const postfixMasterTemplate = `
smtp      inet  n       -       y       -       -       smtpd
submission inet n       -       y       -       -       smtpd
  -o syslog_name=postfix/submission
  -o smtpd_tls_security_level=encrypt
  -o smtpd_sasl_auth_enable=yes
  -o smtpd_tls_auth_only=yes
  -o smtpd_recipient_restrictions=permit_sasl_authenticated,reject
pickup    unix  n       -       y       60      1       pickup
cleanup   unix  n       -       y       -       0       cleanup
qmgr      unix  n       -       n       300     1       qmgr
tlsmgr    unix  -       -       y       1000?   1       tlsmgr
rewrite   unix  -       -       y       -       -       trivial-rewrite
bounce    unix  -       -       y       -       0       bounce
defer     unix  -       -       y       -       0       bounce
trace     unix  -       -       y       -       0       bounce
verify    unix  -       -       y       -       1       verify
flush     unix  n       -       y       1000?   0       flush
proxymap  unix  -       -       n       -       -       proxymap
proxywrite unix -       -       n       -       1       proxymap
smtp      unix  -       -       y       -       -       smtp
relay     unix  -       -       y       -       -       smtp
showq     unix  n       -       y       -       -       showq
error     unix  -       -       y       -       -       error
retry     unix  -       -       y       -       -       error
discard   unix  -       -       y       -       -       discard
local     unix  -       n       n       -       -       local
virtual   unix  -       n       n       -       -       virtual
lmtp      unix  -       -       y       -       -       lmtp
anvil     unix  -       -       y       -       1       anvil
scache    unix  -       -       y       -       1       scache
`

const dovecotSQLTemplate = `
sql_driver = mysql

mysql 127.0.0.1 {
  user = {{ .User }}
  password = {{ .Password }}
  dbname = {{ .Database }}
}

passdb sql {
  default_password_scheme = BLF-CRYPT
  query = SELECT CONCAT(u.local_part, '@', d.name) AS user, u.password_hash AS password FROM users u JOIN domains d ON d.id = u.primary_domain_id WHERE CONCAT(u.local_part, '@', d.name) = '%{user}' AND u.status = 'active' AND d.status IN ('pending','active')
}

userdb sql {
  query = SELECT CONCAT('/var/vmail/', d.name, '/', u.local_part) AS home, 'maildir' AS mail_driver, CONCAT('/var/vmail/', d.name, '/', u.local_part, '/Maildir') AS mail_path, 5000 AS uid, 5000 AS gid FROM users u JOIN domains d ON d.id = u.primary_domain_id WHERE CONCAT(u.local_part, '@', d.name) = '%{user}' AND u.status = 'active' AND d.status IN ('pending','active')
  iterate_query = SELECT CONCAT(u.local_part, '@', d.name) AS user FROM users u JOIN domains d ON d.id = u.primary_domain_id WHERE u.status = 'active' AND d.status IN ('pending','active')
}
`

const dovecotLocalTemplate = `
protocols = imap pop3 lmtp sieve
mail_driver = maildir
mail_home = /var/vmail/%{user|domain}/%{user|username}
mail_path = /var/vmail/%{user|domain}/%{user|username}/Maildir
mail_inbox_path = /var/vmail/%{user|domain}/%{user|username}/Maildir
mail_uid = vmail
mail_gid = vmail
first_valid_uid = 5000
last_valid_uid = 5000
auth_mechanisms = plain login
ssl = required

!include /etc/dovecot/proidentity-sql.conf.ext

service auth {
  unix_listener /var/spool/postfix/private/auth {
    mode = 0660
    user = postfix
    group = postfix
  }
}

service lmtp {
  unix_listener /var/spool/postfix/private/dovecot-lmtp {
    mode = 0600
    user = postfix
    group = postfix
  }
}

protocol lmtp {
  auth_username_format = %{user|lower}
}
`

const postfixMySQLTemplate = `
user = {{ .User }}
password = {{ .Password }}
hosts = 127.0.0.1
dbname = {{ .Database }}
query = {{ .Query }}
`

const rspamdLocalTemplate = `
servers = "127.0.0.1";
`

const rspamdAntivirusTemplate = `
clamav {
  type = "clamav";
  servers = "/run/clamav/clamd.ctl";
  action = "reject";
  symbol = "CLAM_VIRUS";
  scan_mime_parts = true;
  scan_text_mime = true;
  scan_image_mime = true;
  max_size = 50000000;
  log_clean = false;
}
`

const rspamdDKIMSigningTemplate = `
enabled = true;
sign_authenticated = true;
sign_local = true;
sign_inbound = false;
allow_hdrfrom_mismatch = false;
allow_username_mismatch = false;
use_domain = "header";
use_esld = false;
try_fallback = false;
{{- if .Domains }}
domain {
{{- range .Domains }}
  {{ .Domain }} {
    selector = "{{ .Selector }}";
    path = "{{ .KeyPath }}";
  }
{{- end }}
}
{{- end }}
`

const rspamdActionsTemplate = `
reject = 15;
add_header = 6;
greylist = 4;
subject = "[SPAM] %s";
`

const rspamdMilterHeadersTemplate = `
use = ["x-spamd-result", "x-rspamd-server", "x-rspamd-queue-id", "authentication-results"];
authenticated_headers = ["authentication-results"];
`
