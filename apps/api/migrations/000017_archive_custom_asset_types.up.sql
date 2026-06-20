ALTER TABLE custom_asset_types
  ADD COLUMN lifecycle_state varchar(32) NOT NULL DEFAULT 'active',
  ADD CONSTRAINT chk_custom_asset_types_lifecycle_state CHECK (lifecycle_state IN ('active','archived'));

CREATE INDEX idx_custom_asset_types_lifecycle_state
  ON custom_asset_types(tenant_id, lifecycle_state);

CREATE OR REPLACE FUNCTION stuffstash_custom_field_asset_type_target_scope_valid()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
  definition custom_field_definitions%ROWTYPE;
  asset_type custom_asset_types%ROWTYPE;
BEGIN
  IF TG_OP = 'UPDATE'
     AND NEW.custom_field_definition_id IS NOT DISTINCT FROM OLD.custom_field_definition_id
     AND NEW.custom_asset_type_id IS NOT DISTINCT FROM OLD.custom_asset_type_id
     AND NEW.tenant_id IS NOT DISTINCT FROM OLD.tenant_id
     AND NEW.inventory_id IS NOT DISTINCT FROM OLD.inventory_id THEN
    RETURN NEW;
  END IF;

  SELECT * INTO definition
  FROM custom_field_definitions
  WHERE id = NEW.custom_field_definition_id;

  SELECT * INTO asset_type
  FROM custom_asset_types
  WHERE id = NEW.custom_asset_type_id;

  IF asset_type.lifecycle_state <> 'active' THEN
    RAISE foreign_key_violation USING
      MESSAGE = 'custom field target custom asset type must be active',
      CONSTRAINT = 'custom_field_definition_asset_types_active_target';
  END IF;

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

  IF TG_OP = 'UPDATE'
     AND NEW.custom_asset_type_id IS NOT DISTINCT FROM OLD.custom_asset_type_id
     AND NEW.tenant_id IS NOT DISTINCT FROM OLD.tenant_id
     AND NEW.inventory_id IS NOT DISTINCT FROM OLD.inventory_id THEN
    RETURN NEW;
  END IF;

  SELECT * INTO asset_type
  FROM custom_asset_types
  WHERE id = NEW.custom_asset_type_id;

  IF asset_type.lifecycle_state <> 'active' THEN
    RAISE foreign_key_violation USING
      MESSAGE = 'asset custom type must be active',
      CONSTRAINT = 'assets_custom_asset_type_active';
  END IF;

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
