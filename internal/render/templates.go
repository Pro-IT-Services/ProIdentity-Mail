package render

const postfixMainTemplate = `
myhostname = {{ .Hostname }}
myorigin = $myhostname
inet_interfaces = all
inet_protocols = ipv4
smtpd_tls_security_level = may
smtp_tls_security_level = may
smtpd_sasl_auth_enable = yes
smtpd_sasl_type = dovecot
smtpd_sasl_path = private/auth
virtual_transport = lmtp:unix:private/dovecot-lmtp
smtpd_milters = inet:127.0.0.1:11332
non_smtpd_milters = inet:127.0.0.1:11332
milter_protocol = 6
milter_default_action = tempfail
`

const dovecotSQLTemplate = `
driver = mysql
connect = host=127.0.0.1 dbname={{ .Database }} user={{ .User }} password={{ .Password }}
password_query = SELECT local_part AS user, password_hash AS password FROM users WHERE local_part = '%n' AND status = 'active'
`

const rspamdLocalTemplate = `
redis {
  servers = "127.0.0.1";
}
`
