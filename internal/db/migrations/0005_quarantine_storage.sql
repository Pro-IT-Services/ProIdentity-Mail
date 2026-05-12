ALTER TABLE quarantine_events
  ADD COLUMN storage_path varchar(700) NULL AFTER symbols_json,
  ADD COLUMN size_bytes bigint unsigned NOT NULL DEFAULT 0 AFTER storage_path,
  ADD COLUMN sha256 char(64) NULL AFTER size_bytes,
  ADD KEY quarantine_events_sha256_idx (sha256);
