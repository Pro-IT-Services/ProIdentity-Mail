ALTER TABLE user_mfa_challenges
  ADD COLUMN provider varchar(32) NOT NULL DEFAULT 'totp' AFTER purpose,
  ADD COLUMN request_id varchar(190) NOT NULL DEFAULT '' AFTER provider;
