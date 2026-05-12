CREATE TABLE cloudflare_domain_configs (
  domain_id bigint unsigned NOT NULL PRIMARY KEY,
  zone_id varchar(64) NOT NULL DEFAULT '',
  zone_name varchar(253) NOT NULL DEFAULT '',
  api_token text NOT NULL,
  status enum('not_configured','configured','checked','error') NOT NULL DEFAULT 'not_configured',
  last_checked_at timestamp NULL,
  last_error varchar(500) NULL,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  CONSTRAINT cloudflare_domain_configs_domain_id_fk FOREIGN KEY (domain_id) REFERENCES domains(id)
);

CREATE TABLE dns_provision_backups (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  domain_id bigint unsigned NOT NULL,
  provider varchar(40) NOT NULL,
  zone_id varchar(64) NOT NULL,
  mode enum('check','apply') NOT NULL,
  replace_existing tinyint(1) NOT NULL DEFAULT 0,
  plan_json json NOT NULL,
  backup_json json NOT NULL,
  result_json json NOT NULL,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  KEY dns_provision_backups_domain_created_idx (domain_id, created_at),
  CONSTRAINT dns_provision_backups_domain_id_fk FOREIGN KEY (domain_id) REFERENCES domains(id)
);
