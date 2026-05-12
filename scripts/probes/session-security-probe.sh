#!/bin/sh
set -eu

base_admin="${PROIDENTITY_ADMIN_URL:-http://127.0.0.1:8080}"
base_webmail="${PROIDENTITY_WEBMAIL_URL:-http://127.0.0.1:8082}"
env_file="${PROIDENTITY_ENV_FILE:-/etc/proidentity-mail/proidentity-mail.env}"
mailbox_email="${PROIDENTITY_PROBE_MAILBOX:-alice@example.test}"
mailbox_password="${PROIDENTITY_PROBE_PASSWORD:-Password123!ChangeMe}"

tmpdir="$(mktemp -d)"
cleanup() {
	rm -rf "$tmpdir"
}
trap cleanup EXIT

read_env_value() {
	key="$1"
	grep "^${key}=" "$env_file" | tail -n 1 | cut -d= -f2-
}

json_payload() {
	kind="$1"
	subject="$2"
	password="$3"
	KIND="$kind" SUBJECT="$subject" PASSWORD="$password" python3 - <<'PY'
import json
import os

kind = os.environ["KIND"]
subject_key = "email" if kind == "webmail" else "username"
print(json.dumps({subject_key: os.environ["SUBJECT"], "password": os.environ["PASSWORD"]}))
PY
}

extract_csrf() {
	file="$1"
	python3 - "$file" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as handle:
    print(json.load(handle).get("csrf_token", ""))
PY
}

admin_user="$(read_env_value PROIDENTITY_ADMIN_USERNAME)"
admin_password="$(read_env_value PROIDENTITY_ADMIN_PASSWORD)"

json_payload admin "$admin_user" "$admin_password" >"$tmpdir/admin-login.json"
admin_status="$(curl -sS -H 'Content-Type: application/json' --data-binary @"$tmpdir/admin-login.json" -c "$tmpdir/admin.cookies" -o "$tmpdir/admin-session.json" -w '%{http_code}' "$base_admin/api/v1/session")"
admin_csrf=""
admin_no_csrf="skip"
admin_with_csrf="skip"
if [ "$admin_status" = "200" ]; then
	admin_csrf="$(extract_csrf "$tmpdir/admin-session.json")"
	admin_no_csrf="$(curl -sS -H 'Content-Type: application/json' -b "$tmpdir/admin.cookies" -d '{}' -o "$tmpdir/admin-no-csrf.out" -w '%{http_code}' "$base_admin/api/v1/tenants")"
	admin_with_csrf="$(curl -sS -H 'Content-Type: application/json' -H "X-CSRF-Token: $admin_csrf" -b "$tmpdir/admin.cookies" -d '{}' -o "$tmpdir/admin-with-csrf.out" -w '%{http_code}' "$base_admin/api/v1/tenants")"
fi

json_payload webmail "$mailbox_email" "$mailbox_password" >"$tmpdir/webmail-login.json"
webmail_status="$(curl -sS -H 'Content-Type: application/json' --data-binary @"$tmpdir/webmail-login.json" -c "$tmpdir/webmail.cookies" -o "$tmpdir/webmail-session.json" -w '%{http_code}' "$base_webmail/api/v1/session")"
webmail_csrf=""
webmail_no_csrf="skip"
webmail_with_csrf="skip"
if [ "$webmail_status" = "200" ]; then
	webmail_csrf="$(extract_csrf "$tmpdir/webmail-session.json")"
	webmail_no_csrf="$(curl -sS -H 'Content-Type: application/json' -b "$tmpdir/webmail.cookies" -d '{}' -o "$tmpdir/webmail-no-csrf.out" -w '%{http_code}' "$base_webmail/api/v1/send")"
	webmail_with_csrf="$(curl -sS -H 'Content-Type: application/json' -H "X-CSRF-Token: $webmail_csrf" -b "$tmpdir/webmail.cookies" -d '{}' -o "$tmpdir/webmail-with-csrf.out" -w '%{http_code}' "$base_webmail/api/v1/send")"
fi

printf 'admin_login_status=%s admin_csrf_present=%s admin_no_csrf_status=%s admin_with_csrf_status=%s\n' "$admin_status" "$([ -n "$admin_csrf" ] && echo yes || echo no)" "$admin_no_csrf" "$admin_with_csrf"
printf 'webmail_login_status=%s webmail_csrf_present=%s webmail_no_csrf_status=%s webmail_with_csrf_status=%s\n' "$webmail_status" "$([ -n "$webmail_csrf" ] && echo yes || echo no)" "$webmail_no_csrf" "$webmail_with_csrf"

test "$admin_status" = "200"
test -n "$admin_csrf"
test "$admin_no_csrf" = "403"
test "$admin_with_csrf" = "400"
test "$webmail_status" = "200"
test -n "$webmail_csrf"
test "$webmail_no_csrf" = "403"
test "$webmail_with_csrf" = "400"
