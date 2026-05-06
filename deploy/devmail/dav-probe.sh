#!/usr/bin/env bash
set -euo pipefail

domain="${1:-external-1778096561.local}"
tenant_id="${2:-3}"
domain_id="${3:-2}"
local_part="davprobe$(date +%s)"
password="$(openssl rand -hex 12)"
email="${local_part}@${domain}"

curl -sS -m 5 \
  -X POST \
  -H "Content-Type: application/json" \
  --data "{\"tenant_id\":${tenant_id},\"primary_domain_id\":${domain_id},\"local_part\":\"${local_part}\",\"display_name\":\"DAV Probe\",\"password\":\"${password}\"}" \
  http://127.0.0.1:8080/api/v1/users >/tmp/proidentity-dav-user.json

echo "== unauthenticated PROPFIND =="
unauth_code="$(curl -sS -m 5 -o /tmp/proidentity-dav-unauth.out -w "%{http_code}" -X PROPFIND "http://127.0.0.1:8081/dav/principals/${email}/")"
echo "status=${unauth_code}"
if [[ "${unauth_code}" != "401" ]]; then
  cat /tmp/proidentity-dav-unauth.out
  exit 1
fi

echo "== authenticated PROPFIND =="
auth_body="$(curl -sS -m 5 -u "${email}:${password}" -X PROPFIND "http://127.0.0.1:8081/dav/principals/${email}/")"
printf '%s\n' "${auth_body}" | sed -n '1,24p'
if ! printf '%s' "${auth_body}" | grep -q "/dav/calendars/${email}/"; then
  echo "missing calendar home set" >&2
  exit 1
fi
if ! printf '%s' "${auth_body}" | grep -q "/dav/addressbooks/${email}/"; then
  echo "missing addressbook home set" >&2
  exit 1
fi
