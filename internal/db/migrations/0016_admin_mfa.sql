CREATE TABLE admin_mfa_settings (
  id TINYINT UNSIGNED NOT NULL PRIMARY KEY,
  local_totp_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  local_totp_secret VARCHAR(128) NOT NULL DEFAULT '',
  local_totp_pending_secret VARCHAR(128) NOT NULL DEFAULT '',
  proidentity_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  proidentity_base_url VARCHAR(512) NOT NULL DEFAULT '',
  proidentity_api_key TEXT NULL,
  proidentity_user_email VARCHAR(320) NOT NULL DEFAULT '',
  proidentity_timeout_seconds INT NOT NULL DEFAULT 90,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT admin_mfa_settings_singleton CHECK (id = 1)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO admin_mfa_settings(id, proidentity_timeout_seconds)
VALUES (1, 90);

CREATE TABLE admin_mfa_challenges (
  token CHAR(64) NOT NULL PRIMARY KEY,
  username VARCHAR(255) NOT NULL,
  provider VARCHAR(32) NOT NULL,
  request_id VARCHAR(128) NOT NULL DEFAULT '',
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY admin_mfa_challenges_expires_idx (expires_at),
  KEY admin_mfa_challenges_request_idx (request_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
