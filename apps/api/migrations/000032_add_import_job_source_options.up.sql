ALTER TABLE import_jobs
  ADD COLUMN source_allow_private_network BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN source_allow_insecure_tls BOOLEAN NOT NULL DEFAULT FALSE;
