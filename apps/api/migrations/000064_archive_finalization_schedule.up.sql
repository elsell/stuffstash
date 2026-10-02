ALTER TABLE archive_jobs ADD COLUMN next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT '0001-01-01 00:00:00+00';
CREATE INDEX idx_archive_jobs_next_attempt_at ON archive_jobs(next_attempt_at);
