ALTER TABLE audit_records
  DROP CONSTRAINT chk_audit_records_action,
  ADD CONSTRAINT chk_audit_records_action CHECK (
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
  DROP CONSTRAINT chk_audit_records_target_type,
  ADD CONSTRAINT chk_audit_records_target_type CHECK (
    target_type IN ('tenant','inventory','inventory_access_grant','custom_field_definition','asset')
  );

DROP INDEX IF EXISTS idx_assets_custom_asset_type_id;
DROP TRIGGER IF EXISTS trg_asset_custom_asset_type_scope_valid ON assets;
DROP FUNCTION IF EXISTS stuffstash_asset_custom_asset_type_scope_valid();

ALTER TABLE assets
  DROP COLUMN custom_asset_type_id;

DROP TRIGGER IF EXISTS trg_custom_field_asset_type_target_scope_valid ON custom_field_definition_asset_types;
DROP FUNCTION IF EXISTS stuffstash_custom_field_asset_type_target_scope_valid();
DROP TABLE custom_field_definition_asset_types;

ALTER TABLE custom_field_definitions
  DROP CONSTRAINT chk_custom_field_definitions_applicability,
  DROP COLUMN applicability;

DROP TRIGGER IF EXISTS trg_custom_asset_type_effective_key_unique ON custom_asset_types;
DROP FUNCTION IF EXISTS stuffstash_custom_asset_type_effective_key_unique();
DROP TABLE custom_asset_types;
