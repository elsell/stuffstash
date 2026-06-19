CREATE TABLE custom_asset_types (
  id varchar(26) PRIMARY KEY,
  tenant_id varchar(26) NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  inventory_id varchar(26) REFERENCES inventories(id) ON UPDATE CASCADE ON DELETE CASCADE,
  scope varchar(32) NOT NULL,
  cursor_key varchar(32) NOT NULL,
  type_key varchar(80) NOT NULL,
  display_name varchar(120) NOT NULL,
  description varchar(1000) NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_custom_asset_types_scope CHECK (scope IN ('tenant','inventory')),
  CONSTRAINT chk_custom_asset_types_type_key CHECK (type_key ~ '^[a-z][a-z0-9-]{0,79}$'),
  CONSTRAINT chk_custom_asset_types_display_name CHECK (display_name = btrim(display_name) AND length(display_name) BETWEEN 1 AND 120),
  CONSTRAINT chk_custom_asset_types_description CHECK (description = btrim(description) AND length(description) <= 1000),
  CONSTRAINT chk_custom_asset_types_scope_inventory CHECK (
    (scope = 'tenant' AND inventory_id IS NULL)
    OR (scope = 'inventory' AND inventory_id IS NOT NULL)
  ),
  CONSTRAINT chk_custom_asset_types_cursor_key CHECK (
    (scope = 'tenant' AND cursor_key = '0:' || id)
    OR (scope = 'inventory' AND cursor_key = '1:' || id)
  )
);

CREATE UNIQUE INDEX idx_custom_asset_types_tenant_key
  ON custom_asset_types(tenant_id, type_key)
  WHERE inventory_id IS NULL;

CREATE UNIQUE INDEX idx_custom_asset_types_inventory_key
  ON custom_asset_types(tenant_id, inventory_id, type_key)
  WHERE inventory_id IS NOT NULL;

CREATE INDEX idx_custom_asset_types_tenant_cursor
  ON custom_asset_types(tenant_id, cursor_key);

CREATE INDEX idx_custom_asset_types_inventory_cursor
  ON custom_asset_types(tenant_id, inventory_id, cursor_key);

CREATE FUNCTION stuffstash_custom_asset_type_effective_key_unique()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  PERFORM pg_advisory_xact_lock(hashtext(NEW.tenant_id || ':' || NEW.type_key));

  IF NEW.scope = 'tenant' THEN
    IF EXISTS (
      SELECT 1 FROM custom_asset_types existing
      WHERE existing.tenant_id = NEW.tenant_id
        AND existing.type_key = NEW.type_key
        AND existing.id <> NEW.id
    ) THEN
      RAISE unique_violation USING
        MESSAGE = 'custom asset type key conflicts with an effective key in this tenant',
        CONSTRAINT = 'custom_asset_types_effective_key_unique';
    END IF;
  ELSE
    IF EXISTS (
      SELECT 1 FROM custom_asset_types existing
      WHERE existing.tenant_id = NEW.tenant_id
        AND existing.type_key = NEW.type_key
        AND existing.id <> NEW.id
        AND (
          existing.scope = 'tenant'
          OR existing.inventory_id = NEW.inventory_id
        )
    ) THEN
      RAISE unique_violation USING
        MESSAGE = 'custom asset type key conflicts with an effective key in this inventory',
        CONSTRAINT = 'custom_asset_types_effective_key_unique';
    END IF;
  END IF;

  RETURN NEW;
END;
$$;

CREATE TRIGGER trg_custom_asset_type_effective_key_unique
BEFORE INSERT OR UPDATE ON custom_asset_types
FOR EACH ROW
EXECUTE FUNCTION stuffstash_custom_asset_type_effective_key_unique();

ALTER TABLE custom_field_definitions
  ADD COLUMN applicability varchar(32) NOT NULL DEFAULT 'all_assets',
  ADD CONSTRAINT chk_custom_field_definitions_applicability CHECK (applicability IN ('all_assets','custom_asset_types'));

CREATE TABLE custom_field_definition_asset_types (
  custom_field_definition_id varchar(26) NOT NULL REFERENCES custom_field_definitions(id) ON UPDATE CASCADE ON DELETE CASCADE,
  custom_asset_type_id varchar(26) NOT NULL REFERENCES custom_asset_types(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  tenant_id varchar(26) NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  inventory_id varchar(26) REFERENCES inventories(id) ON UPDATE CASCADE ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (custom_field_definition_id, custom_asset_type_id)
);

CREATE INDEX idx_custom_field_definition_asset_types_tenant
  ON custom_field_definition_asset_types(tenant_id);

CREATE INDEX idx_custom_field_definition_asset_types_inventory
  ON custom_field_definition_asset_types(tenant_id, inventory_id)
  WHERE inventory_id IS NOT NULL;

CREATE FUNCTION stuffstash_custom_field_asset_type_target_scope_valid()
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

CREATE TRIGGER trg_custom_field_asset_type_target_scope_valid
BEFORE INSERT OR UPDATE ON custom_field_definition_asset_types
FOR EACH ROW
EXECUTE FUNCTION stuffstash_custom_field_asset_type_target_scope_valid();

ALTER TABLE assets
  ADD COLUMN custom_asset_type_id varchar(26) REFERENCES custom_asset_types(id) ON UPDATE CASCADE ON DELETE RESTRICT;

CREATE INDEX idx_assets_custom_asset_type_id ON assets(custom_asset_type_id);

CREATE FUNCTION stuffstash_asset_custom_asset_type_scope_valid()
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

CREATE TRIGGER trg_asset_custom_asset_type_scope_valid
BEFORE INSERT OR UPDATE ON assets
FOR EACH ROW
EXECUTE FUNCTION stuffstash_asset_custom_asset_type_scope_valid();

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
  ),
  DROP CONSTRAINT chk_audit_records_target_type,
  ADD CONSTRAINT chk_audit_records_target_type CHECK (
    target_type IN ('tenant','inventory','inventory_access_grant','custom_asset_type','custom_field_definition','asset')
  );
