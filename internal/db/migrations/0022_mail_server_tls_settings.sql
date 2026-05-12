ALTER TABLE mail_server_settings
  ADD COLUMN tls_mode enum('system','none','behind-proxy','letsencrypt-http','letsencrypt-dns-cloudflare','custom-cert') NOT NULL DEFAULT 'system' AFTER sni_enabled,
  ADD COLUMN force_https tinyint(1) NOT NULL DEFAULT 1 AFTER tls_mode;
