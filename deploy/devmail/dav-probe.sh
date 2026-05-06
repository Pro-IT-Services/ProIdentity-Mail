#!/usr/bin/env bash
set -euo pipefail

domain="${1:-external-1778096561.local}"
tenant_id="${2:-3}"
domain_id="${3:-2}"
local_part="davprobe$(date +%s)"
password="$(openssl rand -hex 12)"
email="${local_part}@${domain}"
admin_user="$(grep '^PROIDENTITY_ADMIN_USERNAME=' /etc/proidentity-mail/proidentity-mail.env | cut -d= -f2-)"
admin_password="$(grep '^PROIDENTITY_ADMIN_PASSWORD=' /etc/proidentity-mail/proidentity-mail.env | cut -d= -f2-)"

curl -sS -m 5 \
  -X POST \
  -u "${admin_user}:${admin_password}" \
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

echo "== vCard PUT/GET =="
vcard="BEGIN:VCARD
VERSION:4.0
UID:${local_part}
FN:DAV Probe
EMAIL:${email}
END:VCARD
"
printf '%s' "${vcard}" >/tmp/proidentity-dav-contact.vcf
contact_code="$(curl -sS -m 5 -o /tmp/proidentity-dav-contact-put.out -w "%{http_code}" -u "${email}:${password}" -X PUT --data-binary @/tmp/proidentity-dav-contact.vcf "http://127.0.0.1:8081/dav/addressbooks/${email}/default/${local_part}.vcf")"
echo "put_status=${contact_code}"
if [[ "${contact_code}" != "201" ]]; then
  cat /tmp/proidentity-dav-contact-put.out
  exit 1
fi
curl -sS -m 5 -u "${email}:${password}" "http://127.0.0.1:8081/dav/addressbooks/${email}/default/${local_part}.vcf" | tee /tmp/proidentity-dav-contact-get.out | sed -n '1,8p'
grep -q "EMAIL:${email}" /tmp/proidentity-dav-contact-get.out

echo "== iCalendar PUT/GET =="
ics="BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:${local_part}-event
SUMMARY:DAV Probe
END:VEVENT
END:VCALENDAR
"
printf '%s' "${ics}" >/tmp/proidentity-dav-event.ics
event_code="$(curl -sS -m 5 -o /tmp/proidentity-dav-event-put.out -w "%{http_code}" -u "${email}:${password}" -X PUT --data-binary @/tmp/proidentity-dav-event.ics "http://127.0.0.1:8081/dav/calendars/${email}/default/${local_part}.ics")"
echo "put_status=${event_code}"
if [[ "${event_code}" != "201" ]]; then
  cat /tmp/proidentity-dav-event-put.out
  exit 1
fi
curl -sS -m 5 -u "${email}:${password}" "http://127.0.0.1:8081/dav/calendars/${email}/default/${local_part}.ics" | tee /tmp/proidentity-dav-event-get.out | sed -n '1,10p'
grep -q "SUMMARY:DAV Probe" /tmp/proidentity-dav-event-get.out
