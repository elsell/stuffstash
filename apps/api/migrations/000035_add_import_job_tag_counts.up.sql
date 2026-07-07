ALTER TABLE import_jobs
  ADD COLUMN tags integer NOT NULL DEFAULT 0,
  ADD COLUMN tags_created integer NOT NULL DEFAULT 0,
  ADD COLUMN tags_existing integer NOT NULL DEFAULT 0;
