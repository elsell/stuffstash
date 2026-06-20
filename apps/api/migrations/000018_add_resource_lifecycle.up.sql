ALTER TABLE tenants
  ADD COLUMN lifecycle_state varchar(32) NOT NULL DEFAULT 'active',
  ADD CONSTRAINT chk_tenants_lifecycle_state CHECK (lifecycle_state IN ('active','archived'));

ALTER TABLE inventories
  ADD COLUMN lifecycle_state varchar(32) NOT NULL DEFAULT 'active',
  ADD CONSTRAINT chk_inventories_lifecycle_state CHECK (lifecycle_state IN ('active','archived'));

ALTER TABLE custom_field_definitions
  ADD COLUMN lifecycle_state varchar(32) NOT NULL DEFAULT 'active',
  ADD CONSTRAINT chk_custom_field_definitions_lifecycle_state CHECK (lifecycle_state IN ('active','archived'));

ALTER TABLE attachments
  ADD COLUMN lifecycle_state varchar(32) NOT NULL DEFAULT 'active',
  ADD CONSTRAINT chk_attachments_lifecycle_state CHECK (lifecycle_state IN ('active','archived'));

CREATE INDEX idx_tenants_lifecycle_state
  ON tenants(lifecycle_state);

CREATE INDEX idx_inventories_active_tenant_id
  ON inventories(tenant_id, lifecycle_state, id);

CREATE INDEX idx_custom_field_definitions_lifecycle_state
  ON custom_field_definitions(tenant_id, lifecycle_state);

CREATE INDEX idx_attachments_active_page
  ON attachments(tenant_id, inventory_id, asset_id, lifecycle_state, id);

ALTER TABLE audit_records DROP CONSTRAINT IF EXISTS audit_records_tenant_id_fkey;
ALTER TABLE audit_records DROP CONSTRAINT IF EXISTS audit_records_inventory_id_fkey;

ALTER TABLE inventory_access_invitations DROP CONSTRAINT chk_inventory_access_invitations_status;
ALTER TABLE inventory_access_invitations ADD CONSTRAINT chk_inventory_access_invitations_status CHECK (status IN ('pending','accepted','revoked','cancelled'));

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
    'inventory_invitation.accepted',
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
    'audit_record.listed'
));

ALTER TABLE audit_records DROP CONSTRAINT chk_audit_records_target_type;
ALTER TABLE audit_records ADD CONSTRAINT chk_audit_records_target_type CHECK (
  target_type IN ('tenant','inventory','inventory_access_grant','inventory_invitation','custom_asset_type','custom_field_definition','asset','attachment','audit_record')
);
