CREATE TABLE dkim_keys (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  domain_id bigint unsigned NOT NULL,
  selector varchar(63) NOT NULL,
  key_path varchar(500) NOT NULL,
  public_dns_txt text NOT NULL,
  status enum('active','retired') NOT NULL DEFAULT 'active',
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  UNIQUE KEY dkim_keys_domain_selector_unique (domain_id, selector),
  CONSTRAINT dkim_keys_domain_id_fk FOREIGN KEY (domain_id) REFERENCES domains(id)
);

CREATE TABLE quarantine_events (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  tenant_id bigint unsigned NOT NULL,
  user_id bigint unsigned NULL,
  domain_id bigint unsigned NULL,
  message_id varchar(255) NULL,
  sender varchar(320) NULL,
  recipient varchar(320) NOT NULL,
  verdict enum('spam','malware','phishing','policy') NOT NULL,
  action enum('reject','quarantine','mark') NOT NULL,
  scanner varchar(80) NOT NULL,
  symbols_json json NOT NULL,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  KEY quarantine_events_tenant_created_idx (tenant_id, created_at),
  KEY quarantine_events_user_created_idx (user_id, created_at),
  CONSTRAINT quarantine_events_tenant_id_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id),
  CONSTRAINT quarantine_events_user_id_fk FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT quarantine_events_domain_id_fk FOREIGN KEY (domain_id) REFERENCES domains(id)
);
