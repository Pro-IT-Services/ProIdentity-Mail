ALTER TABLE mail_server_settings
  ADD COLUMN mailbox_mfa_enabled tinyint(1) NOT NULL DEFAULT 1 AFTER default_language,
  ADD COLUMN force_mailbox_mfa tinyint(1) NOT NULL DEFAULT 0 AFTER mailbox_mfa_enabled;

CREATE TABLE user_mfa_settings (
  user_id bigint unsigned NOT NULL PRIMARY KEY,
  tenant_id bigint unsigned NOT NULL,
  totp_enabled tinyint(1) NOT NULL DEFAULT 0,
  totp_secret varchar(128) NOT NULL DEFAULT '',
  pending_totp_secret varchar(128) NOT NULL DEFAULT '',
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  KEY user_mfa_settings_tenant_idx (tenant_id),
  CONSTRAINT user_mfa_settings_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT user_mfa_settings_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE TABLE user_mfa_challenges (
  token varchar(128) NOT NULL PRIMARY KEY,
  tenant_id bigint unsigned NOT NULL,
  user_id bigint unsigned NOT NULL,
  purpose enum('login','setup') NOT NULL,
  expires_at timestamp NOT NULL,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  KEY user_mfa_challenges_user_idx (user_id),
  KEY user_mfa_challenges_expires_idx (expires_at),
  CONSTRAINT user_mfa_challenges_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT user_mfa_challenges_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE TABLE user_app_passwords (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  tenant_id bigint unsigned NOT NULL,
  user_id bigint unsigned NOT NULL,
  name varchar(190) NOT NULL,
  secret_sha256 char(64) NOT NULL,
  protocols varchar(120) NOT NULL DEFAULT 'imap,smtp,pop3,dav',
  status enum('active','revoked') NOT NULL DEFAULT 'active',
  last_used_at timestamp NULL,
  last_used_protocol varchar(20) NOT NULL DEFAULT '',
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  revoked_at timestamp NULL,
  UNIQUE KEY user_app_passwords_secret_unique (secret_sha256),
  KEY user_app_passwords_user_idx (user_id, status),
  KEY user_app_passwords_tenant_idx (tenant_id),
  CONSTRAINT user_app_passwords_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT user_app_passwords_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE TABLE tenant_admins (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  tenant_id bigint unsigned NOT NULL,
  user_id bigint unsigned NOT NULL,
  role enum('tenant_admin','read_only') NOT NULL DEFAULT 'tenant_admin',
  status enum('active','disabled') NOT NULL DEFAULT 'active',
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  UNIQUE KEY tenant_admins_tenant_user_unique (tenant_id, user_id),
  KEY tenant_admins_user_idx (user_id),
  CONSTRAINT tenant_admins_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
  CONSTRAINT tenant_admins_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
