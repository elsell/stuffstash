CREATE TABLE inventory_access_invitations (
  id varchar(64) PRIMARY KEY,
  tenant_id varchar(26) NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  inventory_id varchar(26) NOT NULL REFERENCES inventories(id) ON UPDATE CASCADE ON DELETE CASCADE,
  email varchar(320) NOT NULL,
  token_hash varchar(128) NOT NULL,
  relationship varchar(32) NOT NULL,
  status varchar(32) NOT NULL,
  inviter_principal_id varchar(128) NOT NULL,
  accepted_principal_id varchar(128) NOT NULL DEFAULT '',
  expires_at timestamptz NOT NULL,
  accepted_at timestamptz,
  revoked_at timestamptz,
  created_at timestamptz,
  updated_at timestamptz,
  CONSTRAINT chk_inventory_access_invitations_relationship CHECK (relationship IN ('viewer','editor')),
  CONSTRAINT chk_inventory_access_invitations_status CHECK (status IN ('pending','accepted','revoked'))
);

CREATE UNIQUE INDEX idx_inventory_access_invitations_pending
  ON inventory_access_invitations(tenant_id, inventory_id, email, relationship)
  WHERE status = 'pending';

CREATE INDEX idx_inventory_access_invitations_inventory
  ON inventory_access_invitations(tenant_id, inventory_id);

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
