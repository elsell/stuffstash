ALTER TABLE import_jobs
  DROP COLUMN IF EXISTS tags_existing,
  DROP COLUMN IF EXISTS tags_created,
  DROP COLUMN IF EXISTS tags;
