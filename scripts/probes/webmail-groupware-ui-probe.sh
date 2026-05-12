#!/bin/sh
set -eu

base_admin="${PROIDENTITY_ADMIN_URL:-http://127.0.0.1:8080}"
base_webmail="${PROIDENTITY_WEBMAIL_URL:-http://127.0.0.1:8082}"
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

curl -sS -u "$admin_user:$admin_password" "$base_admin/api/v1/domains" >"$tmpdir/domains.json"
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
local_part="guiprobe${stamp}"
email="${local_part}@${DOMAIN_NAME}"
password="ProbePassword${stamp}!"

python3 - "$TENANT_ID" "$DOMAIN_ID" "$local_part" "$password" >"$tmpdir/user.json" <<'PY'
import json
import sys

print(json.dumps({
    "tenant_id": int(sys.argv[1]),
    "primary_domain_id": int(sys.argv[2]),
    "local_part": sys.argv[3],
    "display_name": "Webmail UI Probe",
    "password": sys.argv[4],
}))
PY
create_user_status="$(curl -sS -u "$admin_user:$admin_password" -H 'Content-Type: application/json' --data-binary @"$tmpdir/user.json" -o "$tmpdir/user.out" -w '%{http_code}' "$base_admin/api/v1/users")"

contact_create="$(curl -sS -u "$email:$password" -H 'Content-Type: application/json' -d '{"name":"Ada Lovelace","email":"ada@example.net"}' -o "$tmpdir/contact-create.json" -w '%{http_code}' "$base_webmail/api/v1/contacts")"
contact_id="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("id",""))' "$tmpdir/contact-create.json")"
contact_update="$(curl -sS -u "$email:$password" -H 'Content-Type: application/json' -X PUT -d '{"name":"Ada Byron","email":"ada@lovelace.example"}' -o "$tmpdir/contact-update.json" -w '%{http_code}' "$base_webmail/api/v1/contacts/${contact_id}")"
contact_delete="$(curl -sS -u "$email:$password" -X DELETE -o "$tmpdir/contact-delete.out" -w '%{http_code}' "$base_webmail/api/v1/contacts/${contact_id}")"

event_create="$(curl -sS -u "$email:$password" -H 'Content-Type: application/json' -d '{"title":"Planning","starts_at":"2026-05-07T10:00:00Z","ends_at":"2026-05-07T11:00:00Z"}' -o "$tmpdir/event-create.json" -w '%{http_code}' "$base_webmail/api/v1/calendar")"
event_id="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("id",""))' "$tmpdir/event-create.json")"
event_update="$(curl -sS -u "$email:$password" -H 'Content-Type: application/json' -X PUT -d '{"title":"Planning updated","starts_at":"2026-05-07T12:00:00Z","ends_at":"2026-05-07T13:00:00Z"}' -o "$tmpdir/event-update.json" -w '%{http_code}' "$base_webmail/api/v1/calendar/${event_id}")"
event_delete="$(curl -sS -u "$email:$password" -X DELETE -o "$tmpdir/event-delete.out" -w '%{http_code}' "$base_webmail/api/v1/calendar/${event_id}")"

index_status="$(curl -sS -o "$tmpdir/index.html" -w '%{http_code}' "$base_webmail/")"
ui_has_forms="$(grep -E -c 'contact-modal|event-modal|Phone contact source|Phone calendar source' "$tmpdir/index.html")"

printf 'create_user_status=%s contact=%s/%s/%s event=%s/%s/%s index_status=%s ui_markers=%s\n' "$create_user_status" "$contact_create" "$contact_update" "$contact_delete" "$event_create" "$event_update" "$event_delete" "$index_status" "$ui_has_forms"

test "$create_user_status" = "201"
test "$contact_create" = "201"
test "$contact_update" = "200"
test "$contact_delete" = "204"
test "$event_create" = "201"
test "$event_update" = "200"
test "$event_delete" = "204"
test "$index_status" = "200"
test "$ui_has_forms" -ge 1
