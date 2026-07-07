CREATE UNIQUE INDEX idx_assets_scope_identity
  ON assets(tenant_id, inventory_id, id);

CREATE TABLE asset_checkouts (
  id varchar(26) PRIMARY KEY,
  tenant_id varchar(26) NOT NULL
    REFERENCES tenants(id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT,
  inventory_id varchar(26) NOT NULL
    REFERENCES inventories(id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT,
  asset_id varchar(26) NOT NULL,
  state varchar(32) NOT NULL,
  checked_out_at timestamptz NOT NULL,
  checked_out_by_principal varchar(255) NOT NULL,
  checkout_details text NOT NULL DEFAULT '',
  returned_at timestamptz,
  returned_by_principal varchar(255) NOT NULL DEFAULT '',
  return_details text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT chk_asset_checkouts_state CHECK (state IN ('open','returned','undone')),
  CONSTRAINT chk_asset_checkouts_checkout_details_length CHECK (char_length(checkout_details) <= 1000),
  CONSTRAINT chk_asset_checkouts_return_details_length CHECK (char_length(return_details) <= 1000),
  CONSTRAINT chk_asset_checkouts_return_fields CHECK (
    (state = 'returned' AND returned_at IS NOT NULL AND returned_by_principal <> '')
    OR
    (state <> 'returned' AND returned_at IS NULL AND returned_by_principal = '' AND return_details = '')
  ),
  CONSTRAINT asset_checkouts_scoped_asset_fkey
    FOREIGN KEY (tenant_id, inventory_id, asset_id)
    REFERENCES assets(tenant_id, inventory_id, id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT
);

CREATE UNIQUE INDEX idx_asset_checkouts_one_open
  ON asset_checkouts(tenant_id, inventory_id, asset_id)
  WHERE state = 'open';

CREATE INDEX idx_asset_checkouts_history
  ON asset_checkouts(tenant_id, inventory_id, asset_id, checked_out_at DESC, id DESC);

CREATE INDEX idx_asset_checkouts_checked_out
  ON asset_checkouts(tenant_id, inventory_id, state, checked_out_at DESC, asset_id DESC);

ALTER TABLE undoable_operations
  ALTER COLUMN after_asset DROP NOT NULL,
  ADD COLUMN before_checkout jsonb,
  ADD COLUMN after_checkout jsonb;

ALTER TABLE undoable_operations DROP CONSTRAINT IF EXISTS chk_undoable_operations_source;
ALTER TABLE undoable_operations ADD CONSTRAINT chk_undoable_operations_source CHECK (
  source IN ('api','conversation','mcp','import','background_job','system')
);

ALTER TABLE undoable_operations DROP CONSTRAINT IF EXISTS chk_undoable_operations_original_action;
ALTER TABLE undoable_operations ADD CONSTRAINT chk_undoable_operations_original_action CHECK (original_action IN (
  'asset.created',
  'asset.updated',
  'asset.moved',
  'asset.archived',
  'asset.restored',
  'asset.checked_out',
  'asset.returned'
));

ALTER TABLE undoable_operations DROP CONSTRAINT IF EXISTS chk_undoable_operations_after_asset_object;
ALTER TABLE undoable_operations ADD CONSTRAINT chk_undoable_operations_after_asset_object CHECK (
  after_asset IS NULL OR jsonb_typeof(after_asset) = 'object'
);

ALTER TABLE undoable_operations ADD CONSTRAINT chk_undoable_operations_before_checkout_object CHECK (
  before_checkout IS NULL OR jsonb_typeof(before_checkout) = 'object'
);

ALTER TABLE undoable_operations ADD CONSTRAINT chk_undoable_operations_after_checkout_object CHECK (
  after_checkout IS NULL OR jsonb_typeof(after_checkout) = 'object'
);

ALTER TABLE undoable_operations ADD CONSTRAINT chk_undoable_operations_snapshot_kind CHECK (
  (after_asset IS NOT NULL AND after_checkout IS NULL AND before_checkout IS NULL)
  OR
  (after_asset IS NULL AND after_checkout IS NOT NULL AND before_asset IS NULL)
);

ALTER TABLE audit_records DROP CONSTRAINT chk_audit_records_action;
ALTER TABLE audit_records ADD CONSTRAINT chk_audit_records_action CHECK (action IN (
  'tenant.created',
  'tenant.viewed',
  'tenant.listed',
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
  'asset.checked_out',
  'asset.returned',
  'attachment.created',
  'attachment.viewed',
  'attachment.listed',
  'attachment.content_downloaded',
  'attachment.archived',
  'attachment.restored',
  'attachment.deleted',
  'audit_record.listed',
  'undoable_operation.undone',
  'undoable_operation.redone',
  'provider_profile.created',
  'provider_profile.viewed',
  'provider_profile.listed',
  'provider_profile.updated',
  'provider_profile.enabled',
  'provider_profile.disabled',
  'provider_profile.archived',
  'provider_profile.credential_replaced',
  'provider_profile.tested',
  'voice_provider_configuration.updated',
  'import_job.previewed',
  'import_job.started',
  'import_job.completed',
  'import_job.failed',
  'import_job.cancellation_requested',
  'import_job.cancelled',
  'import_job.history_removed',
  'import_job.credential_cleaned'
));

ALTER TABLE audit_records DROP CONSTRAINT chk_audit_records_target_type;
ALTER TABLE audit_records ADD CONSTRAINT chk_audit_records_target_type CHECK (
  target_type IN ('tenant','inventory','inventory_access_grant','inventory_invitation','custom_asset_type','custom_field_definition','asset','attachment','audit_record','undoable_operation','provider_profile','import_job')
);
