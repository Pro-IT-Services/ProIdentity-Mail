#!/usr/bin/env bash
set -euo pipefail

recipient="${1:-marko@external-1778096561.local}"

echo "== GTUBE spam probe =="
swaks \
  --server 127.0.0.1 \
  --port 25 \
  --from sender@example.net \
  --to "${recipient}" \
  --header "Subject: GTUBE spam probe" \
  --body 'XJS*C4JDBQADN1.NSBN3*2IDNEN*GTUBE-STANDARD-ANTI-UBE-TEST-EMAIL*C.34X' \
  --timeout 20 || true

echo "== EICAR antivirus probe =="
printf '%s' 'WDVPIVAlQEFQWzRcUFpYNTQoUF4pN0NDKTd9JEVJQ0FSLVNUQU5EQVJELUFOVElWSVJVUy1URVNULUZJTEUhJEgrSCo=' | base64 -d > /tmp/proidentity-eicar.com
clamdscan /tmp/proidentity-eicar.com || true
swaks \
  --server 127.0.0.1 \
  --port 25 \
  --from sender@example.net \
  --to "${recipient}" \
  --header "Subject: EICAR antivirus probe" \
  --body 'EICAR attachment probe' \
  --attach @/tmp/proidentity-eicar.com \
  --timeout 20 || true

sleep 2
echo "== Queue =="
postqueue -p
