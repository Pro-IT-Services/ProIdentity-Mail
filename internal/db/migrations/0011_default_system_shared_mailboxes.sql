INSERT IGNORE INTO users(tenant_id, primary_domain_id, local_part, display_name, mailbox_type, password_hash, status, quota_bytes)
SELECT tenant_id, id, 'postmaster', 'Postmaster', 'shared', '', 'active', 1073741824
FROM domains;

INSERT IGNORE INTO users(tenant_id, primary_domain_id, local_part, display_name, mailbox_type, password_hash, status, quota_bytes)
SELECT tenant_id, id, 'abuse', 'Abuse Desk', 'shared', '', 'active', 1073741824
FROM domains;

INSERT IGNORE INTO users(tenant_id, primary_domain_id, local_part, display_name, mailbox_type, password_hash, status, quota_bytes)
SELECT tenant_id, id, 'dmarc', 'DMARC Reports', 'shared', '', 'active', 1073741824
FROM domains;

INSERT IGNORE INTO users(tenant_id, primary_domain_id, local_part, display_name, mailbox_type, password_hash, status, quota_bytes)
SELECT tenant_id, id, 'tlsrpt', 'TLS Reports', 'shared', '', 'active', 1073741824
FROM domains;
