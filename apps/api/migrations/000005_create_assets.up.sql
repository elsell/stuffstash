CREATE TABLE assets (
  id varchar(26) PRIMARY KEY,
  tenant_id varchar(26) NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  inventory_id varchar(26) NOT NULL REFERENCES inventories(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  parent_asset_id varchar(26) REFERENCES assets(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  kind varchar(32) NOT NULL,
  title varchar(160) NOT NULL,
  description text NOT NULL DEFAULT '',
  custom_fields jsonb NOT NULL DEFAULT '{}'::jsonb,
  lifecycle_state varchar(32) NOT NULL,
  created_at timestamptz,
  updated_at timestamptz,
  CONSTRAINT chk_assets_kind CHECK (kind IN ('item', 'container', 'location')),
  CONSTRAINT chk_assets_lifecycle_state CHECK (lifecycle_state IN ('active', 'archived')),
  CONSTRAINT chk_assets_not_own_parent CHECK (parent_asset_id IS NULL OR parent_asset_id <> id),
  CONSTRAINT chk_assets_custom_fields_object CHECK (jsonb_typeof(custom_fields) = 'object')
);

CREATE INDEX idx_assets_tenant_inventory ON assets(tenant_id, inventory_id);
CREATE INDEX idx_assets_parent_asset_id ON assets(parent_asset_id);
CREATE INDEX idx_assets_inventory_parent ON assets(inventory_id, parent_asset_id);
CREATE INDEX idx_assets_inventory_kind ON assets(inventory_id, kind);
