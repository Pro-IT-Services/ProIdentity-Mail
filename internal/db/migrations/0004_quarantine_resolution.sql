ALTER TABLE quarantine_events
  ADD COLUMN status enum('held','released','deleted') NOT NULL DEFAULT 'held' AFTER symbols_json,
  ADD COLUMN resolved_at timestamp NULL AFTER status,
  ADD COLUMN resolution_note varchar(255) NULL AFTER resolved_at,
  ADD KEY quarantine_events_status_created_idx (status, created_at);
