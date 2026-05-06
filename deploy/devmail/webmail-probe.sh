#!/usr/bin/env bash
set -euo pipefail

domain="${1:-external-1778096561.local}"
tenant_id="${2:-3}"
domain_id="${3:-2}"
local_part="webmailprobe$(date +%s)"
password="$(openssl rand -hex 12)"
email="${local_part}@${domain}"

curl -sS -m 5 \
  -X POST \
  -H "Content-Type: application/json" \
  --data "{\"tenant_id\":${tenant_id},\"primary_domain_id\":${domain_id},\"local_part\":\"${local_part}\",\"display_name\":\"Webmail Probe\",\"password\":\"${password}\"}" \
  http://127.0.0.1:8080/api/v1/users >/tmp/proidentity-webmail-user.json

swaks \
  --server 127.0.0.1 \
  --port 25 \
  --from sender@example.net \
  --to "${email}" \
  --header "Subject: Webmail probe" \
  --body "hello webmail" \
  --timeout 20 >/tmp/proidentity-webmail-swaks.log

sleep 2
body="$(curl -sS -m 5 -u "${email}:${password}" "http://127.0.0.1:8082/api/v1/messages?limit=5")"
printf '%s\n' "${body}" | sed -n '1,12p'
if ! printf '%s' "${body}" | grep -q "Webmail probe"; then
  echo "missing delivered probe message" >&2
  cat /tmp/proidentity-webmail-swaks.log >&2
  exit 1
fi
