ALTER TABLE import_jobs
  ADD COLUMN cancellation_request_id VARCHAR(128) NOT NULL DEFAULT '';
