ALTER TABLE admin_mfa_settings
  ADD COLUMN proidentity_totp_enabled BOOLEAN NOT NULL DEFAULT FALSE;
