CREATE TABLE IF NOT EXISTS webmail_user_settings (
  user_id bigint unsigned NOT NULL PRIMARY KEY,
  tenant_id bigint unsigned NOT NULL,
  first_name varchar(120) NOT NULL DEFAULT '',
  last_name varchar(120) NOT NULL DEFAULT '',
  signature_html text NULL,
  signature_auto_add tinyint(1) NOT NULL DEFAULT 0,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  KEY webmail_user_settings_tenant_id_idx (tenant_id),
  CONSTRAINT webmail_user_settings_tenant_id_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id),
  CONSTRAINT webmail_user_settings_user_id_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
