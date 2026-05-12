CREATE TABLE IF NOT EXISTS webmail_content_trust (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  tenant_id bigint unsigned NOT NULL,
  user_id bigint unsigned NOT NULL,
  scope enum('sender','domain') NOT NULL,
  value varchar(255) NOT NULL,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  UNIQUE KEY webmail_content_trust_user_scope_value_idx (user_id, scope, value),
  KEY webmail_content_trust_tenant_id_idx (tenant_id),
  KEY webmail_content_trust_scope_value_idx (scope, value),
  CONSTRAINT webmail_content_trust_tenant_id_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id),
  CONSTRAINT webmail_content_trust_user_id_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
