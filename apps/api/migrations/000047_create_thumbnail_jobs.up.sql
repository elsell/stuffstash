CREATE TABLE thumbnail_jobs (
    attachment_id TEXT NOT NULL REFERENCES attachments(id) ON DELETE CASCADE,
    revision INTEGER NOT NULL,
    tenant_id TEXT NOT NULL,
    inventory_id TEXT NOT NULL,
    asset_id TEXT NOT NULL,
    storage_key TEXT NOT NULL,
    sha256 TEXT NOT NULL,
    priority TEXT NOT NULL,
    status TEXT NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    failure TEXT NOT NULL DEFAULT '',
    claim_id TEXT NOT NULL DEFAULT '',
    claimed_until TIMESTAMPTZ,
    next_attempt_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (attachment_id, revision)
);
CREATE INDEX idx_thumbnail_jobs_pending ON thumbnail_jobs (status, next_attempt_at);
