CREATE TABLE login_rate_limits (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  service varchar(40) NOT NULL,
  limiter_key varchar(512) NOT NULL,
  failure_count int unsigned NOT NULL DEFAULT 0,
  first_failed_at timestamp NULL DEFAULT NULL,
  last_failed_at timestamp NULL DEFAULT NULL,
  locked_until timestamp NULL DEFAULT NULL,
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  UNIQUE KEY login_rate_limits_service_key_unique (service, limiter_key),
  KEY login_rate_limits_locked_until_idx (locked_until),
  KEY login_rate_limits_last_failed_idx (last_failed_at)
);
