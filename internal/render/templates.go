package render

const postfixMainTemplate = `
myhostname = {{ .Hostname }}
myorigin = $myhostname
compatibility_level = 3.6
inet_interfaces = all
inet_protocols = ipv4
smtpd_tls_security_level = may
smtp_tls_security_level = may
smtpd_tls_cert_file = {{ .TLSCertFile }}
smtpd_tls_key_file = {{ .TLSKeyFile }}
{{- if .SNIEnabled }}
tls_server_sni_maps = hash:{{ .SNIMapPath }}
{{- end }}
smtpd_tls_protocols = !SSLv2,!SSLv3,!TLSv1,!TLSv1.1
smtp_tls_protocols = !SSLv2,!SSLv3,!TLSv1,!TLSv1.1
smtpd_tls_mandatory_protocols = !SSLv2,!SSLv3,!TLSv1,!TLSv1.1
smtp_tls_mandatory_protocols = !SSLv2,!SSLv3,!TLSv1,!TLSv1.1
smtpd_tls_exclude_ciphers = aNULL,eNULL,EXPORT,DES,RC4,MD5,PSK,SRP,DSS,3DES
smtp_tls_exclude_ciphers = aNULL,eNULL,EXPORT,DES,RC4,MD5,PSK,SRP,DSS,3DES
smtpd_tls_auth_only = yes
smtpd_sasl_auth_enable = no
smtpd_sasl_type = dovecot
smtpd_sasl_path = private/auth
smtpd_sasl_security_options = noanonymous
smtpd_sasl_authenticated_header = no
disable_vrfy_command = yes
smtpd_delay_reject = yes
smtpd_helo_required = yes
anvil_rate_time_unit = 60s
smtpd_client_connection_count_limit = 20
smtpd_client_connection_rate_limit = 30
smtpd_client_auth_rate_limit = 10
smtpd_client_new_tls_session_rate_limit = 20
smtpd_client_event_limit_exceptions = $mynetworks
smtpd_timeout = 60s
smtpd_starttls_timeout = 60s
message_size_limit = 52428800
mailbox_size_limit = 0
smtpd_sender_login_maps = mysql:/etc/postfix/proidentity/sender-login-maps.cf
smtpd_relay_restrictions = permit_mynetworks,permit_sasl_authenticated,reject_unauth_destination
smtpd_client_restrictions = permit_mynetworks,reject_rbl_client zen.spamhaus.org=127.0.0.[2..11]
smtpd_recipient_restrictions = permit_mynetworks,reject_unauth_destination
smtpd_sender_restrictions = reject_non_fqdn_sender,reject_unknown_sender_domain
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
  -o smtpd_sasl_auth_enable=no
  -o smtpd_data_restrictions=reject_unauth_pipelining
submission inet n       -       y       -       -       smtpd
  -o syslog_name=postfix/submission
  -o smtpd_tls_security_level=encrypt
  -o smtpd_sasl_auth_enable=yes
  -o smtpd_tls_auth_only=yes
  -o smtpd_sasl_authenticated_header=no
  -o smtpd_sender_restrictions=reject_authenticated_sender_login_mismatch,permit_sasl_authenticated,reject
  -o smtpd_recipient_restrictions=permit_sasl_authenticated,reject
  -o smtpd_relay_restrictions=permit_sasl_authenticated,reject
  -o smtpd_client_restrictions=permit_sasl_authenticated,reject
  -o smtpd_data_restrictions=reject_unauth_pipelining
  -o smtpd_client_connection_count_limit=10
  -o smtpd_client_connection_rate_limit=20
  -o smtpd_client_auth_rate_limit=10
  -o smtpd_client_new_tls_session_rate_limit=20
smtps     inet  n       -       y       -       -       smtpd
  -o syslog_name=postfix/smtps
  -o smtpd_tls_wrappermode=yes
  -o smtpd_sasl_auth_enable=yes
  -o smtpd_tls_auth_only=yes
  -o smtpd_sasl_authenticated_header=no
  -o smtpd_sender_restrictions=reject_authenticated_sender_login_mismatch,permit_sasl_authenticated,reject
  -o smtpd_recipient_restrictions=permit_sasl_authenticated,reject
  -o smtpd_relay_restrictions=permit_sasl_authenticated,reject
  -o smtpd_client_restrictions=permit_sasl_authenticated,reject
  -o smtpd_data_restrictions=reject_unauth_pipelining
  -o smtpd_client_connection_count_limit=10
  -o smtpd_client_connection_rate_limit=20
  -o smtpd_client_auth_rate_limit=10
  -o smtpd_client_new_tls_session_rate_limit=20
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
  query = SELECT user, password, nopassword FROM (SELECT CONCAT(u.local_part, '@', d.name) AS user, u.password_hash AS password, NULL AS nopassword, 1 AS priority FROM users u JOIN domains d ON d.id = u.primary_domain_id LEFT JOIN mail_server_settings ms ON ms.id = 1 LEFT JOIN user_mfa_settings mfa ON mfa.user_id = u.id WHERE CONCAT(u.local_part, '@', d.name) = '%{user}' AND u.mailbox_type = 'user' AND u.status = 'active' AND d.status IN ('pending','active') AND COALESCE(ms.force_mailbox_mfa, 0) = 0 AND COALESCE(mfa.totp_enabled, 0) = 0 UNION ALL SELECT CONCAT(u.local_part, '@', d.name) AS user, NULL AS password, 'Y' AS nopassword, 0 AS priority FROM users u JOIN domains d ON d.id = u.primary_domain_id JOIN user_app_passwords ap ON ap.user_id = u.id WHERE CONCAT(u.local_part, '@', d.name) = '%{user}' AND u.mailbox_type = 'user' AND u.status = 'active' AND d.status IN ('pending','active') AND ap.status = 'active' AND ap.secret_sha256 = SHA2('%{password}', 256)) auth_rows ORDER BY priority LIMIT 1
}

userdb sql {
  query = SELECT CONCAT('/var/vmail/', d.name, '/', u.local_part) AS home, 'maildir' AS mail_driver, CONCAT('/var/vmail/', d.name, '/', u.local_part, '/Maildir') AS mail_path, 5000 AS uid, 5000 AS gid FROM users u JOIN domains d ON d.id = u.primary_domain_id WHERE CONCAT(u.local_part, '@', d.name) = '%{user}' AND u.status = 'active' AND d.status IN ('pending','active')
  iterate_query = SELECT CONCAT(u.local_part, '@', d.name) AS user FROM users u JOIN domains d ON d.id = u.primary_domain_id WHERE u.mailbox_type = 'user' AND u.status = 'active' AND d.status IN ('pending','active')
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
auth_failure_delay = 2 secs
auth_internal_failure_delay = 2 secs
auth_verbose = yes
mail_max_userip_connections = 20
ssl = required
ssl_server_cert_file = {{ .TLSCertFile }}
ssl_server_key_file = {{ .TLSKeyFile }}
sieve_script personal {
  driver = file
  path = ~/sieve
  active_path = ~/.dovecot.sieve
}
{{- if .AuthPolicy.Enabled }}
auth_policy_server_url = {{ .AuthPolicy.ServerURL }}
auth_policy_server_api_header = {{ .AuthPolicy.APIHeader }}
auth_policy_hash_nonce = {{ .AuthPolicy.Nonce }}
auth_policy_check_before_auth = yes
auth_policy_check_after_auth = yes
auth_policy_report_after_auth = yes
auth_policy_reject_on_fail = yes
auth_policy_request_attributes {
  login = %{requested_username}
  remote = %{remote_ip}
  protocol = %{protocol}
  session_id = %{session}
  fail_type = %{fail_type}
  success = %{success}
}
{{- end }}
{{- range .SNIHosts }}

local_name {{ .Hostname }} {
  ssl_server_cert_file = {{ .TLSCertFile }}
  ssl_server_key_file = {{ .TLSKeyFile }}
}
{{- end }}

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

service imap-login {
  restart_request_count = 1
  process_min_avail = 0
  process_limit = 256
}

service pop3-login {
  restart_request_count = 1
  process_min_avail = 0
  process_limit = 128
}

service managesieve-login {
  restart_request_count = 1
  process_min_avail = 0
  process_limit = 64
}

protocol lmtp {
  auth_username_format = %{user|lower}
  mail_plugins {
    sieve = yes
  }
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

const rspamdTenantSettingsTemplate = `
{{- range .Domains }}
{{ .RuleName }} {
  priority = high;
  rcpt = "@{{ .Domain }}";
  symbols [
    "{{ .SymbolName }}"
  ]
  apply {
    actions {
{{- if eq .SpamAction "reject" }}
      reject = 6.0;
      "add header" = null;
      "quarantine" = null;
{{- else if eq .SpamAction "quarantine" }}
      "quarantine" = 6.0;
      reject = 999.0;
      "add header" = null;
{{- else }}
      "add header" = 6.0;
      reject = 999.0;
      "quarantine" = null;
{{- end }}
    }
    subject = "[SPAM] %s";
  }
}
{{ end -}}
`

const rspamdForceActionsTemplate = `
rules {
{{- range .Domains }}
  {{ .MalwareRuleName }} {
    action = "{{ .MalwareAction }}";
    expression = "CLAM_VIRUS & {{ .SymbolName }}";
    message = "Rejected due to malware policy";
    honor_action = ["reject"];
  }
{{ end -}}
}
`

const nginxProxyTemplate = `
map $http_upgrade $connection_upgrade {
  default upgrade;
  '' close;
}

map $http_x_forwarded_proto $proidentity_forwarded_proto {
  default $scheme;
  "~^https?$" $http_x_forwarded_proto;
  "" $scheme;
}

{{- if .TrustProxyHeaders }}
{{- range .TrustedProxyCIDRs }}
set_real_ip_from {{ . }};
{{- end }}
real_ip_header {{ .RealIPHeader }};
real_ip_recursive on;
{{- end }}

server {
  listen 80 default_server;
  server_name _;
  return 444;
}

server {
  listen 80;
  server_name {{ .AdminHostname }};
  location ^~ /.well-known/acme-challenge/ {
    root {{ .ACMEWebroot }};
    default_type "text/plain";
  }
  {{- if and .TLSEnabled .ForceHTTPS }}
  location / {
    return 301 https://$host$request_uri;
  }
  {{- else }}
  include /etc/nginx/proidentity/proxy-common.conf;
  location ^~ /internal/ {
    return 404;
  }
  location / {
    proxy_pass http://127.0.0.1:8080;
  }
  {{- end }}
}

server {
  listen 80;
  server_name {{ .WebmailHostname }};
  location ^~ /.well-known/acme-challenge/ {
    root {{ .ACMEWebroot }};
    default_type "text/plain";
  }
  {{- if and .TLSEnabled .ForceHTTPS }}
  location / {
    return 301 https://$host$request_uri;
  }
  {{- else }}
  include /etc/nginx/proidentity/proxy-common.conf;
  location / {
    proxy_pass http://127.0.0.1:8082;
  }
  location /dav/ {
    proxy_pass http://127.0.0.1:8081;
  }
  location /.well-known/caldav {
    proxy_pass http://127.0.0.1:8080;
  }
  location /.well-known/carddav {
    proxy_pass http://127.0.0.1:8080;
  }
  {{- end }}
}

{{- if .MailServerName }}
server {
  listen 80;
  server_name {{ .MailServerName }};
  location ^~ /.well-known/acme-challenge/ {
    root {{ .ACMEWebroot }};
    default_type "text/plain";
  }
  {{- if and .TLSEnabled .ForceHTTPS }}
  location / {
    return 301 https://$host$request_uri;
  }
  {{- else }}
  include /etc/nginx/proidentity/proxy-common.conf;
  location /dav/ {
    proxy_pass http://127.0.0.1:8081;
  }
  location = /.well-known/caldav {
    proxy_pass http://127.0.0.1:8080;
  }
  location = /.well-known/carddav {
    proxy_pass http://127.0.0.1:8080;
  }
  location = /.well-known/autoconfig/mail/config-v1.1.xml {
    proxy_pass http://127.0.0.1:8080;
  }
  location = /mail/config-v1.1.xml {
    proxy_pass http://127.0.0.1:8080;
  }
  location = /autodiscover/autodiscover.xml {
    proxy_pass http://127.0.0.1:8080;
  }
  location = /Autodiscover/Autodiscover.xml {
    proxy_pass http://127.0.0.1:8080;
  }
  location / {
    proxy_pass http://127.0.0.1:8082;
  }
  {{- end }}
}
{{- end }}

{{- if .DiscoveryServerNames }}
server {
  listen 80;
  server_name {{ .DiscoveryServerNames }};
  location ^~ /.well-known/acme-challenge/ {
    root {{ .ACMEWebroot }};
    default_type "text/plain";
  }
  {{- if and .TLSEnabled .ForceHTTPS }}
  location / {
    return 301 https://$host$request_uri;
  }
  {{- else }}
  include /etc/nginx/proidentity/proxy-common.conf;
  location = /.well-known/autoconfig/mail/config-v1.1.xml {
    proxy_pass http://127.0.0.1:8080;
  }
  location = /mail/config-v1.1.xml {
    proxy_pass http://127.0.0.1:8080;
  }
  location = /autodiscover/autodiscover.xml {
    proxy_pass http://127.0.0.1:8080;
  }
  location = /Autodiscover/Autodiscover.xml {
    proxy_pass http://127.0.0.1:8080;
  }
  location / {
    return 404;
  }
  {{- end }}
}
{{- end }}

{{- if .SeparateDAVHost }}
server {
  listen 80;
  server_name {{ .DAVHostname }};
  location ^~ /.well-known/acme-challenge/ {
    root {{ .ACMEWebroot }};
    default_type "text/plain";
  }
  {{- if and .TLSEnabled .ForceHTTPS }}
  location / {
    return 301 https://$host$request_uri;
  }
  {{- else }}
  include /etc/nginx/proidentity/proxy-common.conf;
  location / {
    proxy_pass http://127.0.0.1:8081;
  }
  {{- end }}
}
{{- end }}

{{- if .TLSEnabled }}
server {
  listen 443 ssl default_server;
  http2 on;
  server_name _;
  ssl_certificate {{ .CertPath }};
  ssl_certificate_key {{ .KeyPath }};
  return 444;
}

server {
  listen 443 ssl;
  http2 on;
  server_name {{ .AdminHostname }};
  ssl_certificate {{ .CertPath }};
  ssl_certificate_key {{ .KeyPath }};
  add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
  include /etc/nginx/proidentity/proxy-common.conf;
  location ^~ /internal/ {
    return 404;
  }
  location / {
    proxy_pass http://127.0.0.1:8080;
  }
}

server {
  listen 443 ssl;
  http2 on;
  server_name {{ .WebmailHostname }};
  ssl_certificate {{ .CertPath }};
  ssl_certificate_key {{ .KeyPath }};
  add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
  include /etc/nginx/proidentity/proxy-common.conf;
  location / {
    proxy_pass http://127.0.0.1:8082;
  }
  location /dav/ {
    proxy_pass http://127.0.0.1:8081;
  }
  location /.well-known/caldav {
    proxy_pass http://127.0.0.1:8080;
  }
  location /.well-known/carddav {
    proxy_pass http://127.0.0.1:8080;
  }
}

{{- if .MailServerName }}
server {
  listen 443 ssl;
  http2 on;
  server_name {{ .MailServerName }};
  ssl_certificate {{ .CertPath }};
  ssl_certificate_key {{ .KeyPath }};
  add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
  include /etc/nginx/proidentity/proxy-common.conf;
  location /dav/ {
    proxy_pass http://127.0.0.1:8081;
  }
  location = /.well-known/caldav {
    proxy_pass http://127.0.0.1:8080;
  }
  location = /.well-known/carddav {
    proxy_pass http://127.0.0.1:8080;
  }
  location = /.well-known/autoconfig/mail/config-v1.1.xml {
    proxy_pass http://127.0.0.1:8080;
  }
  location = /mail/config-v1.1.xml {
    proxy_pass http://127.0.0.1:8080;
  }
  location = /autodiscover/autodiscover.xml {
    proxy_pass http://127.0.0.1:8080;
  }
  location = /Autodiscover/Autodiscover.xml {
    proxy_pass http://127.0.0.1:8080;
  }
  location / {
    proxy_pass http://127.0.0.1:8082;
  }
}
{{- end }}

{{- if .DiscoveryServerNames }}
server {
  listen 443 ssl;
  http2 on;
  server_name {{ .DiscoveryServerNames }};
  ssl_certificate {{ .CertPath }};
  ssl_certificate_key {{ .KeyPath }};
  add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
  include /etc/nginx/proidentity/proxy-common.conf;
  location = /.well-known/autoconfig/mail/config-v1.1.xml {
    proxy_pass http://127.0.0.1:8080;
  }
  location = /mail/config-v1.1.xml {
    proxy_pass http://127.0.0.1:8080;
  }
  location = /autodiscover/autodiscover.xml {
    proxy_pass http://127.0.0.1:8080;
  }
  location = /Autodiscover/Autodiscover.xml {
    proxy_pass http://127.0.0.1:8080;
  }
  location / {
    return 404;
  }
}
{{- end }}

{{- if .SeparateDAVHost }}
server {
  listen 443 ssl;
  http2 on;
  server_name {{ .DAVHostname }};
  ssl_certificate {{ .CertPath }};
  ssl_certificate_key {{ .KeyPath }};
  add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
  include /etc/nginx/proidentity/proxy-common.conf;
  location / {
    proxy_pass http://127.0.0.1:8081;
  }
}
{{- end }}
{{- end }}
`

const nginxProxyCommonTemplate = `
proxy_http_version 1.1;
proxy_set_header Host $host;
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
proxy_set_header X-Forwarded-Host $host;
proxy_set_header X-Forwarded-Port $server_port;
proxy_set_header X-Forwarded-Proto $proidentity_forwarded_proto;
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection $connection_upgrade;
proxy_read_timeout 300s;
client_max_body_size 100m;
`

const certbotScriptTemplate = `
#!/bin/sh
set -eu

case "{{ .TLSMode }}" in
  letsencrypt-http)
    install -d -m 0755 {{ .ACMEWebroot }}
    certbot certonly --webroot -w {{ .ACMEWebroot }}{{ range .Hostnames }} -d {{ . }}{{ end }}
    ;;
  letsencrypt-dns-cloudflare)
    if [ -z "{{ .CloudflareCredentialsFile }}" ]; then
      echo "Cloudflare credentials file is required for letsencrypt-dns-cloudflare" >&2
      exit 2
    fi
    {{- if .CloudflareCertDomain }}
    if [ ! -s "{{ .CloudflareCredentialsFile }}" ]; then
      "{{ .MailctlPath }}" cloudflare-cert-credentials --domain "{{ .CloudflareCertDomain }}" --output "{{ .CloudflareCredentialsFile }}"
    fi
    {{- end }}
    certbot certonly --dns-cloudflare --dns-cloudflare-credentials {{ .CloudflareCredentialsFile }} --dns-cloudflare-propagation-seconds {{ .CloudflarePropagationSec }}{{ range .Hostnames }} -d {{ . }}{{ end }}
    ;;
  custom-cert|behind-proxy|none)
    echo "No certbot action for {{ .TLSMode }}"
    ;;
  *)
    echo "Unsupported TLS mode: {{ .TLSMode }}" >&2
    exit 2
    ;;
esac
`
