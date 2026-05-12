ALTER TABLE domains
  MODIFY status enum('pending','active','disabled') NOT NULL DEFAULT 'active';
