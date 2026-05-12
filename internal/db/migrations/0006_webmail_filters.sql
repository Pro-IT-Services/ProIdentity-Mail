CREATE TABLE IF NOT EXISTS mail_filters (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  tenant_id bigint unsigned NOT NULL,
  user_id bigint unsigned NOT NULL,
  name varchar(190) NOT NULL,
  field varchar(32) NOT NULL DEFAULT 'subject',
  operator varchar(32) NOT NULL DEFAULT 'contains',
  value varchar(255) NOT NULL,
  action varchar(32) NOT NULL DEFAULT 'move',
  folder varchar(190) NULL,
  enabled boolean NOT NULL DEFAULT true,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  PRIMARY KEY (id),
  KEY idx_mail_filters_user (user_id, enabled),
  CONSTRAINT fk_mail_filters_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
  CONSTRAINT fk_mail_filters_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
