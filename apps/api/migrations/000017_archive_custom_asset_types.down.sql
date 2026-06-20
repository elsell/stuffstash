DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM custom_asset_types
    WHERE lifecycle_state = 'archived'
  ) THEN
    RAISE EXCEPTION 'cannot roll back custom asset type archive support while archived custom asset types exist';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM audit_records
    WHERE action = 'custom_asset_type.archived'
  ) THEN
    RAISE EXCEPTION 'cannot roll back custom asset type archive support while archive audit records exist';
  END IF;
END $$;

DROP INDEX idx_custom_asset_types_lifecycle_state;

ALTER TABLE custom_asset_types
  DROP CONSTRAINT chk_custom_asset_types_lifecycle_state,
  DROP COLUMN lifecycle_state;

CREATE OR REPLACE FUNCTION stuffstash_custom_field_asset_type_target_scope_valid()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
  definition custom_field_definitions%ROWTYPE;
  asset_type custom_asset_types%ROWTYPE;
BEGIN
  SELECT * INTO definition
  FROM custom_field_definitions
  WHERE id = NEW.custom_field_definition_id;

  SELECT * INTO asset_type
  FROM custom_asset_types
  WHERE id = NEW.custom_asset_type_id;

  IF definition.tenant_id <> NEW.tenant_id OR asset_type.tenant_id <> NEW.tenant_id THEN
    RAISE foreign_key_violation USING
      MESSAGE = 'custom field target must stay inside the tenant boundary',
      CONSTRAINT = 'custom_field_definition_asset_types_tenant_scope';
  END IF;

  IF definition.scope = 'tenant' AND asset_type.scope <> 'tenant' THEN
    RAISE foreign_key_violation USING
      MESSAGE = 'tenant custom fields may only target tenant custom asset types',
      CONSTRAINT = 'custom_field_definition_asset_types_scope';
  END IF;

  IF definition.scope = 'inventory'
     AND asset_type.scope = 'inventory'
     AND asset_type.inventory_id <> definition.inventory_id THEN
    RAISE foreign_key_violation USING
      MESSAGE = 'inventory custom fields may only target visible custom asset types',
      CONSTRAINT = 'custom_field_definition_asset_types_scope';
  END IF;

  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION stuffstash_asset_custom_asset_type_scope_valid()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
  asset_type custom_asset_types%ROWTYPE;
BEGIN
  IF NEW.custom_asset_type_id IS NULL THEN
    RETURN NEW;
  END IF;

  SELECT * INTO asset_type
  FROM custom_asset_types
  WHERE id = NEW.custom_asset_type_id;

  IF asset_type.tenant_id <> NEW.tenant_id THEN
    RAISE foreign_key_violation USING
      MESSAGE = 'asset custom type must stay inside the tenant boundary',
      CONSTRAINT = 'assets_custom_asset_type_tenant_scope';
  END IF;

  IF asset_type.scope = 'inventory' AND asset_type.inventory_id <> NEW.inventory_id THEN
    RAISE foreign_key_violation USING
      MESSAGE = 'asset custom type must be visible to the asset inventory',
      CONSTRAINT = 'assets_custom_asset_type_inventory_scope';
  END IF;

  RETURN NEW;
END;
$$;

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
