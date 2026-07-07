ALTER TABLE import_jobs
  DROP COLUMN IF EXISTS source_allow_insecure_tls,
  DROP COLUMN IF EXISTS source_allow_private_network;
