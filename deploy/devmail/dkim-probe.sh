#!/usr/bin/env bash
set -euo pipefail

user="${1:-marko@external-1778096561.local}"
password="${2:-secret123456}"
mailbox_root="/var/vmail/${user#*@}/${user%@*}/Maildir/new"

swaks \
  --server 127.0.0.1 \
  --port 587 \
  --tls \
  --auth PLAIN \
  --auth-user "${user}" \
  --auth-password "${password}" \
  --from "${user}" \
  --to "${user}" \
  --header "Subject: DKIM signing probe" \
  --body "hello dkim" \
  --timeout 20

sleep 2
latest="$(find "${mailbox_root}" -type f -printf '%T@ %p\n' | sort -n | tail -1 | cut -d' ' -f2-)"
echo "latest=${latest}"
grep -E '^(DKIM-Signature|Authentication-Results|X-Spamd-Result|X-Rspamd)' "${latest}" || true
if ! grep -q '^DKIM-Signature:' "${latest}"; then
  echo "missing DKIM-Signature header" >&2
  sed -n '1,80p' "${latest}" >&2
  exit 1
fi
postqueue -p
