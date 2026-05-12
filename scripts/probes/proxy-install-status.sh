#!/bin/sh
set -eu

if certbot plugins 2>/tmp/certbot-plugins.err | grep -E 'dns-cloudflare|webroot' >/tmp/certbot-plugins.match; then
	printf 'certbot_plugins=%s\n' "$(wc -l </tmp/certbot-plugins.match)"
else
	echo 'certbot_plugins=0'
fi

test -s /etc/nginx/conf.d/proidentity.conf
test -s /etc/nginx/proidentity/proxy-common.conf
test -x /opt/proidentity-mail/bin/proidentity-issue-cert
nginx -t >/tmp/nginx-proidentity-test.out 2>&1
printf 'proxy_files=present nginx_config=ok cert_helper=installed\n'
