CREATE TABLE undoable_operations (
  id varchar(26) PRIMARY KEY,
  tenant_id varchar(26) NOT NULL,
  inventory_id varchar(26) NOT NULL,
  principal_id varchar(255) NOT NULL,
  source varchar(64) NOT NULL,
  target_type varchar(64) NOT NULL,
  target_id varchar(255) NOT NULL,
  original_action varchar(96) NOT NULL,
  status varchar(32) NOT NULL,
  created_at timestamptz NOT NULL,
  last_applied_at timestamptz,
  before_asset jsonb,
  after_asset jsonb NOT NULL,
  undo_audit_record_id varchar(26),
  redo_audit_record_id varchar(26),
  CONSTRAINT undoable_operations_tenant_id_fkey
    FOREIGN KEY (tenant_id)
    REFERENCES tenants(id)
    ON DELETE RESTRICT,
  CONSTRAINT undoable_operations_inventory_id_fkey
    FOREIGN KEY (inventory_id)
    REFERENCES inventories(id)
    ON DELETE RESTRICT,
  CONSTRAINT undoable_operations_undo_audit_record_id_fkey
    FOREIGN KEY (undo_audit_record_id)
    REFERENCES audit_records(id)
    ON DELETE RESTRICT,
  CONSTRAINT undoable_operations_redo_audit_record_id_fkey
    FOREIGN KEY (redo_audit_record_id)
    REFERENCES audit_records(id)
    ON DELETE RESTRICT,
  CONSTRAINT chk_undoable_operations_source CHECK (source IN ('api','system')),
  CONSTRAINT chk_undoable_operations_target_type CHECK (target_type IN ('asset')),
  CONSTRAINT chk_undoable_operations_original_action CHECK (original_action IN (
    'asset.created',
    'asset.updated',
    'asset.moved',
    'asset.archived',
    'asset.restored'
  )),
  CONSTRAINT chk_undoable_operations_status CHECK (status IN ('available','undone','redone')),
  CONSTRAINT chk_undoable_operations_before_asset_object CHECK (before_asset IS NULL OR jsonb_typeof(before_asset) = 'object'),
  CONSTRAINT chk_undoable_operations_after_asset_object CHECK (jsonb_typeof(after_asset) = 'object')
);

CREATE INDEX idx_undoable_operations_scope
  ON undoable_operations(tenant_id, inventory_id, created_at, id);

CREATE INDEX idx_undoable_operations_target
  ON undoable_operations(target_type, target_id, created_at, id);

ALTER TABLE audit_records DROP CONSTRAINT chk_audit_records_action;
ALTER TABLE audit_records ADD CONSTRAINT chk_audit_records_action CHECK (action IN (
    'tenant.created',
    'tenant.viewed',
    'tenant.updated',
    'tenant.archived',
    'tenant.restored',
    'tenant.deleted',
    'inventory.created',
    'inventory.viewed',
    'inventory.listed',
    'inventory.updated',
    'inventory.archived',
    'inventory.restored',
    'inventory.deleted',
    'inventory_access.granted',
    'inventory_access_grant.viewed',
    'inventory_access_grant.listed',
    'inventory_access.revoked',
    'inventory_invitation.created',
    'inventory_invitation.viewed',
    'inventory_invitation.listed',
    'inventory_invitation.accepted',
    'inventory_invitation.expiration_updated',
    'inventory_invitation.revoked',
    'inventory_invitation.cancelled',
    'inventory_invitation.deleted',
    'custom_asset_type.created',
    'custom_asset_type.viewed',
    'custom_asset_type.listed',
    'custom_asset_type.updated',
    'custom_asset_type.archived',
    'custom_asset_type.restored',
    'custom_asset_type.deleted',
    'custom_field_definition.created',
    'custom_field_definition.viewed',
    'custom_field_definition.listed',
    'custom_field_definition.updated',
    'custom_field_definition.archived',
    'custom_field_definition.restored',
    'custom_field_definition.deleted',
    'asset.created',
    'asset.viewed',
    'asset.listed',
    'asset.updated',
    'asset.moved',
    'asset.archived',
    'asset.restored',
    'asset.deleted',
    'attachment.created',
    'attachment.viewed',
    'attachment.listed',
    'attachment.content_downloaded',
    'attachment.archived',
    'attachment.restored',
    'attachment.deleted',
    'audit_record.listed',
    'undoable_operation.undone',
    'undoable_operation.redone'
));

ALTER TABLE audit_records DROP CONSTRAINT chk_audit_records_target_type;
ALTER TABLE audit_records ADD CONSTRAINT chk_audit_records_target_type CHECK (
  target_type IN ('tenant','inventory','inventory_access_grant','inventory_invitation','custom_asset_type','custom_field_definition','asset','attachment','audit_record','undoable_operation')
);
