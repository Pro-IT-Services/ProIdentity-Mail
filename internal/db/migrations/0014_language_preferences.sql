ALTER TABLE mail_server_settings
  ADD COLUMN default_language varchar(8) NOT NULL DEFAULT 'en' AFTER sni_enabled;

ALTER TABLE webmail_user_settings
  ADD COLUMN language varchar(8) NOT NULL DEFAULT '' AFTER signature_auto_add;
