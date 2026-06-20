ALTER TABLE audit_records DROP CONSTRAINT chk_audit_records_action;
ALTER TABLE audit_records ADD CONSTRAINT chk_audit_records_action CHECK (action IN (
    'tenant.created',
    'inventory.created',
    'inventory_access.granted',
    'inventory_access.revoked',
    'inventory_invitation.created',
    'inventory_invitation.accepted',
    'inventory_invitation.revoked',
    'custom_asset_type.created',
    'custom_asset_type.updated',
    'custom_asset_type.archived',
    'custom_field_definition.created',
    'custom_field_definition.updated',
    'asset.created',
    'asset.updated',
    'asset.moved',
    'asset.archived',
    'asset.restored',
    'attachment.created'
));

ALTER TABLE audit_records DROP CONSTRAINT chk_audit_records_target_type;
ALTER TABLE audit_records ADD CONSTRAINT chk_audit_records_target_type CHECK (
  target_type IN ('tenant','inventory','inventory_access_grant','inventory_invitation','custom_asset_type','custom_field_definition','asset')
);

ALTER TABLE inventory_access_invitations DROP CONSTRAINT chk_inventory_access_invitations_status;
ALTER TABLE inventory_access_invitations ADD CONSTRAINT chk_inventory_access_invitations_status CHECK (status IN ('pending','accepted','revoked'));

ALTER TABLE audit_records
  ADD CONSTRAINT audit_records_tenant_id_fkey
  FOREIGN KEY (tenant_id)
  REFERENCES tenants(id)
  ON UPDATE CASCADE
  ON DELETE RESTRICT;

ALTER TABLE audit_records
  ADD CONSTRAINT audit_records_inventory_id_fkey
  FOREIGN KEY (inventory_id)
  REFERENCES inventories(id)
  ON UPDATE CASCADE
  ON DELETE RESTRICT;

DROP INDEX IF EXISTS idx_attachments_active_page;
DROP INDEX IF EXISTS idx_custom_field_definitions_lifecycle_state;
DROP INDEX IF EXISTS idx_inventories_active_tenant_id;
DROP INDEX IF EXISTS idx_tenants_lifecycle_state;

ALTER TABLE attachments DROP CONSTRAINT IF EXISTS chk_attachments_lifecycle_state;
ALTER TABLE attachments DROP COLUMN IF EXISTS lifecycle_state;

ALTER TABLE custom_field_definitions DROP CONSTRAINT IF EXISTS chk_custom_field_definitions_lifecycle_state;
ALTER TABLE custom_field_definitions DROP COLUMN IF EXISTS lifecycle_state;

ALTER TABLE inventories DROP CONSTRAINT IF EXISTS chk_inventories_lifecycle_state;
ALTER TABLE inventories DROP COLUMN IF EXISTS lifecycle_state;

ALTER TABLE tenants DROP CONSTRAINT IF EXISTS chk_tenants_lifecycle_state;
ALTER TABLE tenants DROP COLUMN IF EXISTS lifecycle_state;
