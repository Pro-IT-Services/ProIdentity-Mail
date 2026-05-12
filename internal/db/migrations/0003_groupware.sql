CREATE TABLE calendars (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  tenant_id bigint unsigned NOT NULL,
  user_id bigint unsigned NOT NULL,
  slug varchar(120) NOT NULL,
  display_name varchar(190) NOT NULL,
  color varchar(32) NOT NULL DEFAULT '#0f766e',
  timezone varchar(80) NOT NULL DEFAULT 'UTC',
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  UNIQUE KEY calendars_user_slug_unique (user_id, slug),
  KEY calendars_tenant_id_idx (tenant_id),
  CONSTRAINT calendars_tenant_id_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id),
  CONSTRAINT calendars_user_id_fk FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE calendar_objects (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  calendar_id bigint unsigned NOT NULL,
  uid varchar(255) NOT NULL,
  href varchar(255) NOT NULL,
  etag varchar(128) NOT NULL,
  icalendar longtext NOT NULL,
  starts_at datetime NULL,
  ends_at datetime NULL,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  UNIQUE KEY calendar_objects_calendar_href_unique (calendar_id, href),
  UNIQUE KEY calendar_objects_calendar_uid_unique (calendar_id, uid),
  KEY calendar_objects_time_idx (starts_at, ends_at),
  CONSTRAINT calendar_objects_calendar_id_fk FOREIGN KEY (calendar_id) REFERENCES calendars(id)
);

CREATE TABLE address_books (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  tenant_id bigint unsigned NOT NULL,
  user_id bigint unsigned NOT NULL,
  slug varchar(120) NOT NULL,
  display_name varchar(190) NOT NULL,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  UNIQUE KEY address_books_user_slug_unique (user_id, slug),
  KEY address_books_tenant_id_idx (tenant_id),
  CONSTRAINT address_books_tenant_id_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id),
  CONSTRAINT address_books_user_id_fk FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE contact_objects (
  id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  address_book_id bigint unsigned NOT NULL,
  uid varchar(255) NOT NULL,
  href varchar(255) NOT NULL,
  etag varchar(128) NOT NULL,
  vcard longtext NOT NULL,
  full_name varchar(255) NULL,
  email varchar(320) NULL,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  updated_at timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  UNIQUE KEY contact_objects_book_href_unique (address_book_id, href),
  UNIQUE KEY contact_objects_book_uid_unique (address_book_id, uid),
  KEY contact_objects_email_idx (email),
  CONSTRAINT contact_objects_address_book_id_fk FOREIGN KEY (address_book_id) REFERENCES address_books(id)
);
