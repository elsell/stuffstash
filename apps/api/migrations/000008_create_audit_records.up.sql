CREATE TABLE audit_records (
  id varchar(26) PRIMARY KEY,
  tenant_id varchar(26) NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  inventory_id varchar(26) REFERENCES inventories(id) ON UPDATE CASCADE ON DELETE CASCADE,
  principal_id varchar(128) NOT NULL,
  action varchar(80) NOT NULL,
  source varchar(40) NOT NULL,
  target_type varchar(80) NOT NULL,
  target_id varchar(180) NOT NULL,
  occurred_at timestamptz NOT NULL,
  request_id varchar(128) NOT NULL DEFAULT '',
  metadata jsonb NOT NULL DEFAULT '{}',
  created_at timestamptz,
  updated_at timestamptz,
  CONSTRAINT chk_audit_records_action CHECK (
    action IN (
      'tenant.created',
      'inventory.created',
      'inventory_access.granted',
      'custom_field_definition.created',
      'asset.created',
      'asset.updated',
      'asset.moved'
    )
  ),
  CONSTRAINT chk_audit_records_source CHECK (
    source IN ('api','conversation','mcp','import','background_job','system')
  ),
  CONSTRAINT chk_audit_records_target_type CHECK (
    target_type IN ('tenant','inventory','inventory_access_grant','custom_field_definition','asset')
  ),
  CONSTRAINT chk_audit_records_metadata_object CHECK (jsonb_typeof(metadata) = 'object')
);

CREATE INDEX idx_audit_records_tenant_id
  ON audit_records(tenant_id, id);

CREATE INDEX idx_audit_records_inventory_id
  ON audit_records(tenant_id, inventory_id, id)
  WHERE inventory_id IS NOT NULL;

CREATE INDEX idx_audit_records_principal_id
  ON audit_records(principal_id);

CREATE INDEX idx_audit_records_action
  ON audit_records(action);

CREATE INDEX idx_audit_records_target
  ON audit_records(target_type, target_id);

CREATE INDEX idx_audit_records_occurred_at
  ON audit_records(occurred_at);

CREATE INDEX idx_audit_records_request_id
  ON audit_records(request_id);
