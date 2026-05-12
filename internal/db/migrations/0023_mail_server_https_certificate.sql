ALTER TABLE mail_server_settings
  ADD COLUMN https_certificate_id bigint unsigned NULL AFTER force_https,
  ADD CONSTRAINT mail_server_settings_https_certificate_fk FOREIGN KEY (https_certificate_id) REFERENCES tls_certificates(id);
