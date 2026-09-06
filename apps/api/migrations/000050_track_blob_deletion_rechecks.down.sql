DROP INDEX IF EXISTS idx_blob_deletion_rechecks;
ALTER TABLE blob_deletion_events DROP COLUMN recheck_failed;
ALTER TABLE blob_deletion_events DROP COLUMN rechecked_at;
