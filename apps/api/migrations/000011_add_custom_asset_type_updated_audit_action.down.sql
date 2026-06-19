ALTER TABLE audit_records
  DROP CONSTRAINT chk_audit_records_action,
  ADD CONSTRAINT chk_audit_records_action CHECK (
    action IN (
      'tenant.created',
      'inventory.created',
      'inventory_access.granted',
      'custom_asset_type.created',
      'custom_field_definition.created',
      'asset.created',
      'asset.updated',
      'asset.moved'
    )
  );
