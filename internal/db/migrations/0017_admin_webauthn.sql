ALTER TABLE admin_mfa_settings
  ADD COLUMN native_webauthn_enabled BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE admin_webauthn_credentials (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  credential_id VARBINARY(512) NOT NULL,
  name VARCHAR(160) NOT NULL,
  credential_json LONGBLOB NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_used_at TIMESTAMP NULL DEFAULT NULL,
  UNIQUE KEY admin_webauthn_credentials_credential_id_unique (credential_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE admin_webauthn_sessions (
  token CHAR(64) NOT NULL PRIMARY KEY,
  ceremony VARCHAR(32) NOT NULL,
  session_json LONGBLOB NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY admin_webauthn_sessions_expires_idx (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
