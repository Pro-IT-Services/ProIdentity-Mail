ALTER TABLE users
  ADD COLUMN mailbox_type enum('user','shared') NOT NULL DEFAULT 'user' AFTER display_name;

CREATE TABLE catch_all_routes (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  tenant_id bigint unsigned NOT NULL,
  domain_id bigint unsigned NOT NULL,
  destination varchar(320) NOT NULL,
  status enum('active','disabled') NOT NULL DEFAULT 'active',
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  UNIQUE KEY catch_all_routes_domain_unique (domain_id),
  KEY catch_all_routes_tenant_id_idx (tenant_id),
  CONSTRAINT catch_all_routes_tenant_id_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id),
  CONSTRAINT catch_all_routes_domain_id_fk FOREIGN KEY (domain_id) REFERENCES domains(id)
);

CREATE TABLE shared_mailbox_permissions (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  tenant_id bigint unsigned NOT NULL,
  shared_mailbox_id bigint unsigned NOT NULL,
  user_id bigint unsigned NOT NULL,
  can_read tinyint(1) NOT NULL DEFAULT 1,
  can_send_as tinyint(1) NOT NULL DEFAULT 0,
  can_send_on_behalf tinyint(1) NOT NULL DEFAULT 0,
  can_manage tinyint(1) NOT NULL DEFAULT 0,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  UNIQUE KEY shared_permissions_unique (shared_mailbox_id, user_id),
  KEY shared_permissions_tenant_id_idx (tenant_id),
  KEY shared_permissions_user_id_idx (user_id),
  CONSTRAINT shared_permissions_tenant_id_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id),
  CONSTRAINT shared_permissions_shared_mailbox_id_fk FOREIGN KEY (shared_mailbox_id) REFERENCES users(id),
  CONSTRAINT shared_permissions_user_id_fk FOREIGN KEY (user_id) REFERENCES users(id)
);
