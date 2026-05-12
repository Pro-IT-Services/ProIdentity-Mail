#!/bin/sh
set -eu

admin_host="${PROIDENTITY_ADMIN_HOSTNAME:-admin.mail.local}"
webmail_host="${PROIDENTITY_WEBMAIL_HOSTNAME:-mail.local}"
dav_host="${PROIDENTITY_DAV_HOSTNAME:-dav.mail.local}"

admin_status="$(curl -sS -H "Host: ${admin_host}" -o /tmp/proxy-admin.out -w '%{http_code}' http://127.0.0.1/)"
webmail_status="$(curl -sS -H "Host: ${webmail_host}" -o /tmp/proxy-webmail.out -w '%{http_code}' http://127.0.0.1/healthz)"
dav_status="$(curl -sS -H "Host: ${dav_host}" -o /tmp/proxy-dav.out -w '%{http_code}' http://127.0.0.1/healthz)"

printf 'admin_status=%s webmail_status=%s dav_status=%s\n' "$admin_status" "$webmail_status" "$dav_status"

test "$admin_status" = "200"
test "$webmail_status" = "200"
test "$dav_status" = "200"
