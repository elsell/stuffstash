CREATE TABLE import_jobs (
    id VARCHAR(26) PRIMARY KEY,
    tenant_id VARCHAR(26) NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    inventory_id VARCHAR(26) NOT NULL REFERENCES inventories(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    actor_id VARCHAR(128) NOT NULL,
    status VARCHAR(64) NOT NULL,
    source_type VARCHAR(64) NOT NULL,
    source_name VARCHAR(128) NOT NULL,
    source_base_url VARCHAR(2048) NOT NULL,
    source_version VARCHAR(128) NOT NULL,
    source_image_import VARCHAR(64) NOT NULL,
    source_fingerprint VARCHAR(128) NOT NULL,
    fields INTEGER NOT NULL DEFAULT 0,
    locations INTEGER NOT NULL DEFAULT 0,
    assets INTEGER NOT NULL DEFAULT 0,
    attachments INTEGER NOT NULL DEFAULT 0,
    warnings INTEGER NOT NULL DEFAULT 0,
    errors INTEGER NOT NULL DEFAULT 0,
    fields_created INTEGER NOT NULL DEFAULT 0,
    fields_existing INTEGER NOT NULL DEFAULT 0,
    locations_created INTEGER NOT NULL DEFAULT 0,
    assets_created INTEGER NOT NULL DEFAULT 0,
    assets_skipped INTEGER NOT NULL DEFAULT 0,
    attachments_created INTEGER NOT NULL DEFAULT 0,
    attachments_skipped INTEGER NOT NULL DEFAULT 0,
    records_discarded INTEGER NOT NULL DEFAULT 0,
    source_links_discarded INTEGER NOT NULL DEFAULT 0,
    preview_json BYTEA NULL,
    progress_phase VARCHAR(64) NOT NULL,
    progress_done INTEGER NOT NULL DEFAULT 0,
    progress_total INTEGER NOT NULL DEFAULT 0,
    progress_message VARCHAR(512) NOT NULL DEFAULT '',
    progress_updated_at TIMESTAMP NULL,
    progress_history_json BYTEA NULL,
    cancellation_mode VARCHAR(64) NOT NULL DEFAULT '',
    messages_json BYTEA NOT NULL,
    history_removed_at TIMESTAMP NULL,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_import_jobs_tenant_id ON import_jobs (tenant_id);
CREATE INDEX idx_import_jobs_inventory_id ON import_jobs (inventory_id);
CREATE INDEX idx_import_jobs_actor_id ON import_jobs (actor_id);
CREATE INDEX idx_import_jobs_status ON import_jobs (status);
CREATE INDEX idx_import_jobs_inventory_created ON import_jobs (tenant_id, inventory_id, created_at DESC);

CREATE TABLE import_job_sources (
    job_id VARCHAR(26) PRIMARY KEY REFERENCES import_jobs(id) ON UPDATE CASCADE ON DELETE CASCADE,
    tenant_id VARCHAR(26) NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    inventory_id VARCHAR(26) NOT NULL REFERENCES inventories(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    key_id VARCHAR(128) NOT NULL,
    algorithm VARCHAR(64) NOT NULL,
    nonce BYTEA NOT NULL,
    ciphertext BYTEA NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_import_job_sources_tenant_id ON import_job_sources (tenant_id);
CREATE INDEX idx_import_job_sources_inventory_id ON import_job_sources (inventory_id);
CREATE INDEX idx_import_job_sources_expires_at ON import_job_sources (expires_at);

CREATE TABLE import_source_links (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(26) NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    inventory_id VARCHAR(26) NOT NULL REFERENCES inventories(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    job_id VARCHAR(26) NOT NULL REFERENCES import_jobs(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    source_type VARCHAR(64) NOT NULL,
    source_instance_key VARCHAR(2048) NOT NULL,
    source_entity_type VARCHAR(64) NOT NULL,
    source_entity_id VARCHAR(512) NOT NULL,
    resource_type VARCHAR(64) NOT NULL,
    resource_id VARCHAR(64) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    CONSTRAINT uq_import_source_links_source UNIQUE (
        tenant_id,
        inventory_id,
        source_type,
        source_instance_key,
        source_entity_type,
        source_entity_id
    )
);

CREATE INDEX idx_import_source_links_tenant_id ON import_source_links (tenant_id);
CREATE INDEX idx_import_source_links_inventory_id ON import_source_links (inventory_id);
CREATE INDEX idx_import_source_links_job_id ON import_source_links (job_id);
CREATE INDEX idx_import_source_links_resource_id ON import_source_links (resource_id);

CREATE TABLE import_job_resources (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(26) NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    inventory_id VARCHAR(26) NOT NULL REFERENCES inventories(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    job_id VARCHAR(26) NOT NULL REFERENCES import_jobs(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    resource_type VARCHAR(64) NOT NULL,
    resource_id VARCHAR(64) NOT NULL,
    resource_owner_id VARCHAR(64) NOT NULL DEFAULT '',
    source_type VARCHAR(64) NOT NULL,
    source_instance_key VARCHAR(2048) NOT NULL,
    source_entity_type VARCHAR(64) NOT NULL,
    source_entity_id VARCHAR(512) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    CONSTRAINT uq_import_job_resources_resource UNIQUE (resource_type, resource_id)
);

CREATE INDEX idx_import_job_resources_tenant_id ON import_job_resources (tenant_id);
CREATE INDEX idx_import_job_resources_inventory_id ON import_job_resources (inventory_id);
CREATE INDEX idx_import_job_resources_job ON import_job_resources (tenant_id, inventory_id, job_id);
