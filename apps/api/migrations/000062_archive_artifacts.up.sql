-- No job foreign key: sources are journaled before their job is created.
CREATE TABLE archive_artifacts (
 storage_key TEXT PRIMARY KEY,
 job_id TEXT NOT NULL,
 tenant_id TEXT NOT NULL REFERENCES tenants(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
 source_inventory_id TEXT NOT NULL,
 kind TEXT NOT NULL CHECK (kind IN ('source','plan','export','restored_media')),
 expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_archive_artifacts_job_id ON archive_artifacts(job_id);
CREATE INDEX idx_archive_artifacts_expires_at ON archive_artifacts(expires_at);
