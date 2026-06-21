CREATE TABLE attachments (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    inventory_id TEXT NOT NULL REFERENCES inventories(id) ON UPDATE CASCADE ON DELETE CASCADE,
    asset_id TEXT NOT NULL REFERENCES assets(id) ON UPDATE CASCADE ON DELETE CASCADE,
    storage_key TEXT NOT NULL UNIQUE,
    file_name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    sha256 TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT chk_attachments_content_type CHECK (content_type IN ('image/jpeg', 'image/png', 'application/pdf')),
    CONSTRAINT chk_attachments_size_bytes CHECK (size_bytes > 0),
    CONSTRAINT chk_attachments_sha256 CHECK (length(sha256) = 64)
);

CREATE INDEX idx_attachments_asset_page ON attachments(tenant_id, inventory_id, asset_id, id);

ALTER TABLE audit_records DROP CONSTRAINT chk_audit_records_action;
ALTER TABLE audit_records ADD CONSTRAINT chk_audit_records_action CHECK (action IN (
    'tenant.created',
    'inventory.created',
    'inventory_access.granted',
    'custom_asset_type.created',
    'custom_asset_type.updated',
    'custom_field_definition.created',
    'custom_field_definition.updated',
    'asset.created',
    'asset.updated',
    'asset.moved',
    'asset.archived',
    'asset.restored',
    'attachment.created'
));
