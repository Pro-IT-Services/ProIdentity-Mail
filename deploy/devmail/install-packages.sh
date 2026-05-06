#!/usr/bin/env bash
set -euo pipefail

apt-get update
DEBIAN_FRONTEND=noninteractive apt-get install -y \
  ca-certificates \
  curl \
  mariadb-server \
  redis-server \
  postfix \
  postfix-mysql \
  dovecot-core \
  dovecot-imapd \
  dovecot-pop3d \
  dovecot-lmtpd \
  dovecot-mysql \
  dovecot-sieve \
  dovecot-managesieved \
  rspamd \
  clamav-daemon \
  clamav-freshclam

systemctl enable mariadb redis-server postfix dovecot rspamd clamav-daemon
