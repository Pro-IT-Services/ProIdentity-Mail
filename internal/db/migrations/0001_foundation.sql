CREATE TABLE tenants (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name varchar(190) NOT NULL,
  slug varchar(120) NOT NULL,
  status enum('active','suspended') NOT NULL DEFAULT 'active',
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  UNIQUE KEY tenants_slug_unique (slug)
);

CREATE TABLE domains (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  tenant_id bigint unsigned NOT NULL,
  name varchar(253) NOT NULL,
  status enum('pending','active','disabled') NOT NULL DEFAULT 'active',
  dkim_selector varchar(63) NOT NULL DEFAULT 'mail',
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  UNIQUE KEY domains_name_unique (name),
  KEY domains_tenant_id_idx (tenant_id),
  CONSTRAINT domains_tenant_id_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

CREATE TABLE users (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  tenant_id bigint unsigned NOT NULL,
  primary_domain_id bigint unsigned NOT NULL,
  local_part varchar(128) NOT NULL,
  display_name varchar(190) NOT NULL,
  password_hash varchar(255) NOT NULL,
  status enum('active','locked','disabled') NOT NULL DEFAULT 'active',
  quota_bytes bigint unsigned NOT NULL DEFAULT 10737418240,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  UNIQUE KEY users_mailbox_unique (primary_domain_id, local_part),
  KEY users_tenant_id_idx (tenant_id),
  CONSTRAINT users_tenant_id_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id),
  CONSTRAINT users_primary_domain_id_fk FOREIGN KEY (primary_domain_id) REFERENCES domains(id)
);

CREATE TABLE aliases (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  tenant_id bigint unsigned NOT NULL,
  domain_id bigint unsigned NOT NULL,
  source_local_part varchar(128) NOT NULL,
  destination varchar(320) NOT NULL,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  UNIQUE KEY aliases_source_destination_unique (domain_id, source_local_part, destination),
  KEY aliases_tenant_id_idx (tenant_id),
  CONSTRAINT aliases_tenant_id_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id),
  CONSTRAINT aliases_domain_id_fk FOREIGN KEY (domain_id) REFERENCES domains(id)
);

CREATE TABLE tenant_policies (
  tenant_id bigint unsigned NOT NULL PRIMARY KEY,
  spam_action enum('mark','quarantine','reject') NOT NULL DEFAULT 'mark',
  malware_action enum('quarantine','reject') NOT NULL DEFAULT 'quarantine',
  require_tls_for_auth tinyint(1) NOT NULL DEFAULT 1,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  CONSTRAINT tenant_policies_tenant_id_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

CREATE TABLE audit_events (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  tenant_id bigint unsigned NULL,
  actor_type varchar(40) NOT NULL,
  actor_id bigint unsigned NULL,
  action varchar(120) NOT NULL,
  target_type varchar(80) NOT NULL,
  target_id varchar(120) NOT NULL,
  metadata_json json NOT NULL,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  KEY audit_events_tenant_id_created_idx (tenant_id, created_at),
  KEY audit_events_action_created_idx (action, created_at)
);
