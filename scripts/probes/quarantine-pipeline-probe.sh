#!/bin/sh
set -eu

base_admin="${PROIDENTITY_ADMIN_URL:-http://127.0.0.1:8080}"
env_file="${PROIDENTITY_ENV_FILE:-/etc/proidentity-mail/proidentity-mail.env}"

tmpdir="$(mktemp -d)"
cleanup() {
	rm -rf "$tmpdir"
}
trap cleanup EXIT

set -a
. "$env_file"
set +a

admin_user="${PROIDENTITY_ADMIN_USERNAME}"
admin_password="${PROIDENTITY_ADMIN_PASSWORD}"

curl_admin() {
	method="$1"
	path="$2"
	body="${3:-}"
	if [ -n "$body" ]; then
		curl -sS -u "$admin_user:$admin_password" -H 'Content-Type: application/json' -X "$method" -d "$body" "$base_admin$path"
	else
		curl -sS -u "$admin_user:$admin_password" -H 'Content-Type: application/json' -X "$method" "$base_admin$path"
	fi
}

curl_admin GET /api/v1/domains >"$tmpdir/domains.json"
python3 - "$tmpdir/domains.json" >"$tmpdir/domain.env" <<'PY'
import json
import sys

domains = json.load(open(sys.argv[1], encoding="utf-8"))
if not domains:
    raise SystemExit("no domains available")
domain = domains[0]
print(f"DOMAIN_ID={domain['id']}")
print(f"TENANT_ID={domain['tenant_id']}")
print(f"DOMAIN_NAME={domain['name']}")
PY
. "$tmpdir/domain.env"

stamp="$(date +%s)"
local_part="qprobe${stamp}"
recipient="${local_part}@${DOMAIN_NAME}"
password="ProbePassword${stamp}!"

python3 - "$TENANT_ID" "$DOMAIN_ID" "$local_part" "$password" >"$tmpdir/user.json" <<'PY'
import json
import sys

print(json.dumps({
    "tenant_id": int(sys.argv[1]),
    "primary_domain_id": int(sys.argv[2]),
    "local_part": sys.argv[3],
    "display_name": "Quarantine Probe",
    "password": sys.argv[4],
}))
PY
create_user_status="$(curl -sS -u "$admin_user:$admin_password" -H 'Content-Type: application/json' --data-binary @"$tmpdir/user.json" -o "$tmpdir/user.out" -w '%{http_code}' "$base_admin/api/v1/users")"

cat >"$tmpdir/message.eml" <<EOF
From: scanner-probe@example.net
To: ${recipient}
Subject: Quarantine release probe
Message-ID: <qprobe-${stamp}@example.net>

probe body
EOF

if command -v runuser >/dev/null 2>&1 && [ "$(id -u)" = "0" ]; then
	quarantine_output="$(runuser -u proidentity --preserve-environment -- /opt/proidentity-mail/bin/mailctl quarantine-message -recipient "$recipient" -sender scanner-probe@example.net -message-id "qprobe-${stamp}" -verdict malware -action quarantine -scanner Probe -symbols '{"TEST":1}' <"$tmpdir/message.eml")"
else
	quarantine_output="$(/opt/proidentity-mail/bin/mailctl quarantine-message -recipient "$recipient" -sender scanner-probe@example.net -message-id "qprobe-${stamp}" -verdict malware -action quarantine -scanner Probe -symbols '{"TEST":1}' <"$tmpdir/message.eml")"
fi
event_id="$(printf '%s\n' "$quarantine_output" | sed -n 's/.*id=\([0-9][0-9]*\).*/\1/p')"
if [ -z "$event_id" ]; then
	echo "missing_event_id"
	exit 1
fi

release_status="$(curl -sS -u "$admin_user:$admin_password" -H 'Content-Type: application/json' -d '{"resolution_note":"probe release"}' -o "$tmpdir/release.out" -w '%{http_code}' "$base_admin/api/v1/quarantine/${event_id}/release")"
mail_count=0
for _ in 1 2 3 4 5 6 7 8 9 10; do
	mail_count="$(find "/var/vmail/${DOMAIN_NAME}/${local_part}/Maildir/new" -type f 2>/dev/null | wc -l)"
	if [ "$mail_count" -ge 1 ]; then
		break
	fi
	sleep 1
done

printf 'create_user_status=%s quarantine_event_id=%s release_status=%s delivered_count=%s\n' "$create_user_status" "$event_id" "$release_status" "$mail_count"

test "$create_user_status" = "201"
test "$release_status" = "200"
test "$mail_count" -ge 1
