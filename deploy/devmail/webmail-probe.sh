#!/usr/bin/env bash
set -euo pipefail

domain="${1:-external-1778096561.local}"
tenant_id="${2:-3}"
domain_id="${3:-2}"
local_part="webmailprobe$(date +%s)"
password="$(openssl rand -hex 12)"
email="${local_part}@${domain}"
admin_user="$(grep '^PROIDENTITY_ADMIN_USERNAME=' /etc/proidentity-mail/proidentity-mail.env | cut -d= -f2-)"
admin_password="$(grep '^PROIDENTITY_ADMIN_PASSWORD=' /etc/proidentity-mail/proidentity-mail.env | cut -d= -f2-)"

curl -sS -m 5 \
  -X POST \
  -u "${admin_user}:${admin_password}" \
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

message_id="$(printf '%s' "${body}" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p' | head -1)"
detail="$(curl -sS -m 5 -u "${email}:${password}" "http://127.0.0.1:8082/api/v1/messages/${message_id}")"
printf '%s\n' "${detail}" | sed -n '1,12p'
if ! printf '%s' "${detail}" | grep -q "hello webmail"; then
  echo "missing full message body" >&2
  exit 1
fi

report_code="$(curl -sS -m 5 -u "${email}:${password}" -H "Content-Type: application/json" --data '{"verdict":"spam"}' -o /tmp/proidentity-webmail-report.out -w "%{http_code}" "http://127.0.0.1:8082/api/v1/messages/${message_id}/report")"
echo "report_status=${report_code}"
if [[ "${report_code}" != "202" ]]; then
  cat /tmp/proidentity-webmail-report.out
  exit 1
fi
spam_list="$(curl -sS -m 5 -u "${email}:${password}" "http://127.0.0.1:8082/api/v1/messages?folder=spam&limit=5")"
if ! printf '%s' "${spam_list}" | grep -q "Webmail probe"; then
  echo "reported message did not move to spam folder" >&2
  exit 1
fi

send_body="{\"to\":[\"${email}\"],\"subject\":\"Webmail API send\",\"body\":\"hello from compose api\"}"
send_code="$(curl -sS -m 5 -u "${email}:${password}" -H "Content-Type: application/json" --data "${send_body}" -o /tmp/proidentity-webmail-send.out -w "%{http_code}" http://127.0.0.1:8082/api/v1/send)"
echo "send_status=${send_code}"
if [[ "${send_code}" != "202" ]]; then
  cat /tmp/proidentity-webmail-send.out
  exit 1
fi
sleep 2
sent_list="$(curl -sS -m 5 -u "${email}:${password}" "http://127.0.0.1:8082/api/v1/messages?limit=10")"
printf '%s\n' "${sent_list}" | sed -n '1,12p'
if ! printf '%s' "${sent_list}" | grep -q "Webmail API send"; then
  echo "sent message did not arrive" >&2
  exit 1
fi
