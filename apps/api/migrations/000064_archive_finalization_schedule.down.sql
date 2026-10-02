DROP INDEX idx_archive_jobs_next_attempt_at;
ALTER TABLE archive_jobs DROP COLUMN next_attempt_at;
