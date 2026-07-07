CREATE TABLE asset_tags (
  id varchar(26) PRIMARY KEY,
  tenant_id varchar(26) NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  inventory_id varchar(26) NOT NULL REFERENCES inventories(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  key varchar(80) NOT NULL,
  display_name varchar(80) NOT NULL,
  color varchar(7) NOT NULL DEFAULT '',
  lifecycle_state varchar(32) NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT chk_asset_tags_key CHECK (key ~ '^[a-z0-9][a-z0-9-]{0,79}$'),
  CONSTRAINT chk_asset_tags_color CHECK (color = '' OR color ~ '^#[0-9A-F]{6}$'),
  CONSTRAINT chk_asset_tags_lifecycle_state CHECK (lifecycle_state IN ('active','archived')),
  CONSTRAINT asset_tags_scope_key_unique UNIQUE (tenant_id, inventory_id, key)
);

CREATE INDEX idx_asset_tags_scope_id
  ON asset_tags(tenant_id, inventory_id, id);

CREATE TABLE asset_tag_assignments (
  tenant_id varchar(26) NOT NULL,
  inventory_id varchar(26) NOT NULL,
  asset_id varchar(26) NOT NULL REFERENCES assets(id) ON UPDATE CASCADE ON DELETE CASCADE,
  tag_id varchar(26) NOT NULL REFERENCES asset_tags(id) ON UPDATE CASCADE ON DELETE CASCADE,
  created_at timestamptz NOT NULL,
  PRIMARY KEY (tenant_id, inventory_id, asset_id, tag_id)
);

CREATE INDEX idx_asset_tag_assignments_asset
  ON asset_tag_assignments(tenant_id, inventory_id, asset_id);

CREATE INDEX idx_asset_tag_assignments_tag
  ON asset_tag_assignments(tag_id);

ALTER TABLE audit_records DROP CONSTRAINT chk_audit_records_action;
ALTER TABLE audit_records ADD CONSTRAINT chk_audit_records_action CHECK (action IN (
  'tenant.created','tenant.viewed','tenant.listed','tenant.updated','tenant.archived','tenant.restored','tenant.deleted',
  'inventory.created','inventory.viewed','inventory.listed','inventory.updated','inventory.archived','inventory.restored','inventory.deleted',
  'inventory_access.granted','inventory_access_grant.viewed','inventory_access_grant.listed','inventory_access.revoked',
  'inventory_invitation.created','inventory_invitation.viewed','inventory_invitation.listed','inventory_invitation.accepted','inventory_invitation.expiration_updated','inventory_invitation.revoked','inventory_invitation.cancelled','inventory_invitation.deleted',
  'custom_asset_type.created','custom_asset_type.viewed','custom_asset_type.listed','custom_asset_type.updated','custom_asset_type.archived','custom_asset_type.restored','custom_asset_type.deleted',
  'custom_field_definition.created','custom_field_definition.viewed','custom_field_definition.listed','custom_field_definition.updated','custom_field_definition.archived','custom_field_definition.restored','custom_field_definition.deleted',
  'asset.created','asset.viewed','asset.listed','asset.updated','asset.moved','asset.archived','asset.restored','asset.deleted','asset.checked_out','asset.returned',
  'asset_tag.created','asset_tag.listed','asset_tag.updated','asset_tag.archived',
  'attachment.created','attachment.viewed','attachment.listed','attachment.content_downloaded','attachment.archived','attachment.restored','attachment.deleted',
  'audit_record.listed','undoable_operation.undone','undoable_operation.redone',
  'provider_profile.created','provider_profile.viewed','provider_profile.listed','provider_profile.updated','provider_profile.enabled','provider_profile.disabled','provider_profile.archived','provider_profile.credential_replaced','provider_profile.tested',
  'voice_provider_configuration.updated',
  'import_job.previewed','import_job.started','import_job.completed','import_job.failed','import_job.cancellation_requested','import_job.cancelled','import_job.history_removed','import_job.credential_cleaned'
));

ALTER TABLE audit_records DROP CONSTRAINT chk_audit_records_target_type;
ALTER TABLE audit_records ADD CONSTRAINT chk_audit_records_target_type CHECK (
  target_type IN ('tenant','inventory','inventory_access_grant','inventory_invitation','custom_asset_type','custom_field_definition','asset','asset_tag','attachment','audit_record','undoable_operation','provider_profile','import_job')
);
