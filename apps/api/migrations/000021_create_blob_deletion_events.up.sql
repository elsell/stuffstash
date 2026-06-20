CREATE TABLE blob_deletion_events (
    id TEXT PRIMARY KEY,
    storage_key TEXT NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT '',
    claim_id TEXT NOT NULL DEFAULT '',
    claimed_until TIMESTAMPTZ,
    processed_at TIMESTAMPTZ,
    dead_lettered_at TIMESTAMPTZ,
    dead_letter_reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_blob_deletion_events_storage_key ON blob_deletion_events (storage_key);
CREATE INDEX idx_blob_deletion_events_claim_id ON blob_deletion_events (claim_id);
CREATE INDEX idx_blob_deletion_events_dead_lettered_at ON blob_deletion_events (dead_lettered_at);
CREATE INDEX idx_blob_deletion_events_created_at ON blob_deletion_events (created_at);
