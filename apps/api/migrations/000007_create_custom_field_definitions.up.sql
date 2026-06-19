CREATE FUNCTION stuffstash_custom_field_enum_options_valid(options jsonb)
RETURNS boolean
LANGUAGE sql
IMMUTABLE
AS $$
  SELECT
    jsonb_typeof(options) = 'array'
    AND NOT EXISTS (
      SELECT 1
      FROM jsonb_array_elements(
        CASE WHEN jsonb_typeof(options) = 'array' THEN options ELSE '[]'::jsonb END
      ) AS option(value)
      WHERE jsonb_typeof(option.value) <> 'string'
        OR option.value #>> '{}' !~ '^[a-z][a-z0-9-]{0,79}$'
    )
    AND (
      SELECT count(*)
      FROM jsonb_array_elements_text(
        CASE WHEN jsonb_typeof(options) = 'array' THEN options ELSE '[]'::jsonb END
      ) AS option(value)
    ) = (
      SELECT count(DISTINCT option.value)
      FROM jsonb_array_elements_text(
        CASE WHEN jsonb_typeof(options) = 'array' THEN options ELSE '[]'::jsonb END
      ) AS option(value)
    );
$$;

CREATE TABLE custom_field_definitions (
  id varchar(26) PRIMARY KEY,
  tenant_id varchar(26) NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  inventory_id varchar(26) REFERENCES inventories(id) ON UPDATE CASCADE ON DELETE CASCADE,
  scope varchar(32) NOT NULL,
  cursor_key varchar(32) NOT NULL,
  field_key varchar(80) NOT NULL,
  display_name varchar(120) NOT NULL,
  field_type varchar(32) NOT NULL,
  enum_options jsonb NOT NULL DEFAULT '[]',
  created_at timestamptz,
  updated_at timestamptz,
  CONSTRAINT chk_custom_field_definitions_scope CHECK (scope IN ('tenant','inventory')),
  CONSTRAINT chk_custom_field_definitions_field_type CHECK (field_type IN ('text','number','boolean','date','url','enum')),
  CONSTRAINT chk_custom_field_definitions_field_key CHECK (field_key ~ '^[a-z][a-z0-9-]{0,79}$'),
  CONSTRAINT chk_custom_field_definitions_display_name CHECK (display_name = btrim(display_name) AND length(display_name) BETWEEN 1 AND 120),
  CONSTRAINT chk_custom_field_definitions_scope_inventory CHECK (
    (scope = 'tenant' AND inventory_id IS NULL)
    OR (scope = 'inventory' AND inventory_id IS NOT NULL)
  ),
  CONSTRAINT chk_custom_field_definitions_cursor_key CHECK (
    (scope = 'tenant' AND cursor_key = '0:' || id)
    OR (scope = 'inventory' AND cursor_key = '1:' || id)
  ),
  CONSTRAINT chk_custom_field_definitions_enum_options CHECK (
    (field_type = 'enum' AND CASE WHEN jsonb_typeof(enum_options) = 'array' THEN jsonb_array_length(enum_options) > 0 ELSE false END)
    OR (field_type <> 'enum' AND enum_options = '[]'::jsonb)
  ),
  CONSTRAINT chk_custom_field_definitions_enum_option_values CHECK (
    field_type <> 'enum'
    OR stuffstash_custom_field_enum_options_valid(enum_options)
  )
);

CREATE UNIQUE INDEX idx_custom_field_definitions_tenant_key
  ON custom_field_definitions(tenant_id, field_key)
  WHERE scope = 'tenant';

CREATE UNIQUE INDEX idx_custom_field_definitions_inventory_key
  ON custom_field_definitions(tenant_id, inventory_id, field_key)
  WHERE scope = 'inventory';

CREATE INDEX idx_custom_field_definitions_tenant_cursor
  ON custom_field_definitions(tenant_id, cursor_key);

CREATE INDEX idx_custom_field_definitions_inventory_cursor
  ON custom_field_definitions(tenant_id, inventory_id, cursor_key);

CREATE FUNCTION stuffstash_custom_field_effective_key_unique()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  PERFORM pg_advisory_xact_lock(hashtextextended(NEW.tenant_id || ':' || NEW.field_key, 0));

  IF NEW.scope = 'tenant' THEN
    IF EXISTS (
      SELECT 1
      FROM custom_field_definitions existing
      WHERE existing.tenant_id = NEW.tenant_id
        AND existing.field_key = NEW.field_key
        AND existing.id <> NEW.id
    ) THEN
      RAISE EXCEPTION 'custom field definition key already exists in effective tenant scope'
        USING ERRCODE = '23505',
              CONSTRAINT = 'custom_field_definitions_effective_key_unique';
    END IF;
  ELSE
    IF EXISTS (
      SELECT 1
      FROM custom_field_definitions existing
      WHERE existing.tenant_id = NEW.tenant_id
        AND existing.field_key = NEW.field_key
        AND existing.id <> NEW.id
        AND (
          existing.scope = 'tenant'
          OR existing.inventory_id = NEW.inventory_id
        )
    ) THEN
      RAISE EXCEPTION 'custom field definition key already exists in effective inventory scope'
        USING ERRCODE = '23505',
              CONSTRAINT = 'custom_field_definitions_effective_key_unique';
    END IF;
  END IF;

  RETURN NEW;
END;
$$;

CREATE TRIGGER trg_custom_field_effective_key_unique
BEFORE INSERT OR UPDATE ON custom_field_definitions
FOR EACH ROW
EXECUTE FUNCTION stuffstash_custom_field_effective_key_unique();
