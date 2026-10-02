CREATE TABLE archive_jobs (
 id TEXT PRIMARY KEY,
 tenant_id TEXT NOT NULL REFERENCES tenants(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
 source_inventory_id TEXT NOT NULL,
 principal_id TEXT NOT NULL,
 request_key TEXT NOT NULL,
 request_json TEXT NOT NULL,
 state_json TEXT NOT NULL,
 state TEXT NOT NULL CHECK (state IN ('queued','running','awaiting_approval','ready','failed','cancelled','expired')),
 revision BIGINT NOT NULL CHECK (revision > 0),
 created_at TIMESTAMPTZ NOT NULL,
 updated_at TIMESTAMPTZ NOT NULL,
 expires_at TIMESTAMPTZ NOT NULL,
 lease_until TIMESTAMPTZ NOT NULL
);
CREATE UNIQUE INDEX idx_archive_job_request ON archive_jobs(tenant_id, principal_id, request_key);
CREATE INDEX idx_archive_job_scope ON archive_jobs(tenant_id, source_inventory_id);
CREATE INDEX idx_archive_job_queue ON archive_jobs(state, lease_until);
CREATE INDEX idx_archive_jobs_expires_at ON archive_jobs(expires_at);
