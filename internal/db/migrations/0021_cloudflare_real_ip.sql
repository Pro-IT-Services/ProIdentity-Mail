ALTER TABLE mail_server_settings
  ADD COLUMN cloudflare_real_ip_enabled tinyint(1) NOT NULL DEFAULT 0 AFTER force_mailbox_mfa;
