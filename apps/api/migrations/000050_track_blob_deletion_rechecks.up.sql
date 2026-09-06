ALTER TABLE blob_deletion_events ADD COLUMN rechecked_at TIMESTAMPTZ;
ALTER TABLE blob_deletion_events ADD COLUMN recheck_failed BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX idx_blob_deletion_rechecks ON blob_deletion_events (rechecked_at, processed_at) WHERE processed_at IS NOT NULL;
