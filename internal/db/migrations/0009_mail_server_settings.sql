CREATE TABLE mail_server_settings (
  id tinyint unsigned NOT NULL PRIMARY KEY DEFAULT 1,
  hostname_mode enum('shared','head-domain','per-domain') NOT NULL DEFAULT 'shared',
  mail_hostname varchar(253) NOT NULL DEFAULT '',
  head_tenant_id bigint unsigned NULL,
  head_domain_id bigint unsigned NULL,
  public_ipv4 varchar(45) NOT NULL DEFAULT '',
  public_ipv6 varchar(45) NOT NULL DEFAULT '',
  sni_enabled tinyint(1) NOT NULL DEFAULT 0,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  CONSTRAINT mail_server_settings_singleton CHECK (id = 1),
  CONSTRAINT mail_server_settings_head_tenant_fk FOREIGN KEY (head_tenant_id) REFERENCES tenants(id),
  CONSTRAINT mail_server_settings_head_domain_fk FOREIGN KEY (head_domain_id) REFERENCES domains(id)
);
