CREATE TABLE print_jobs (
 id VARCHAR(26) PRIMARY KEY,
 tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
 inventory_id TEXT NOT NULL REFERENCES inventories(id) ON DELETE RESTRICT,
 printer_id TEXT NOT NULL REFERENCES printers(id) ON DELETE RESTRICT,
 requested_by TEXT NOT NULL,
 idempotency_key TEXT NOT NULL,
 status TEXT NOT NULL CHECK (status IN ('queued','claimed','printing','completed','failed','uncertain','canceled')),
 media_fingerprint TEXT NOT NULL,
 request_fingerprint TEXT NOT NULL,
 revision BIGINT NOT NULL CHECK (revision > 0),
 snapshot BYTEA NOT NULL,
 created_at TIMESTAMPTZ NOT NULL,
 updated_at TIMESTAMPTZ NOT NULL,
 CONSTRAINT idx_print_job_request UNIQUE (tenant_id,inventory_id,requested_by,idempotency_key)
);
CREATE INDEX idx_print_job_queue ON print_jobs(tenant_id,inventory_id,printer_id,status,created_at,id);
CREATE TABLE print_attempt_index (
 id TEXT PRIMARY KEY,
 tenant_id TEXT NOT NULL,
 inventory_id TEXT NOT NULL,
 connector_id TEXT NOT NULL REFERENCES print_connectors(id) ON DELETE RESTRICT,
 job_id VARCHAR(26) NOT NULL REFERENCES print_jobs(id) ON DELETE CASCADE
);
CREATE INDEX idx_print_attempt_owner ON print_attempt_index(tenant_id,inventory_id,connector_id,id);
