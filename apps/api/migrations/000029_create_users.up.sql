CREATE TABLE users (
  id varchar(128) PRIMARY KEY,
  email varchar(320) NOT NULL DEFAULT '',
  created_at timestamptz,
  updated_at timestamptz
);
