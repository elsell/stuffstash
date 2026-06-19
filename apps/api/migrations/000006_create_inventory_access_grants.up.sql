ALTER TABLE authorization_outbox_events DROP CONSTRAINT chk_authorization_outbox_events_kind;
ALTER TABLE authorization_outbox_events DROP CONSTRAINT chk_authorization_outbox_events_inventory_required;

ALTER TABLE authorization_outbox_events
  ADD CONSTRAINT chk_authorization_outbox_events_kind CHECK (
    kind IN ('grant_tenant_owner','grant_inventory_owner','grant_inventory_viewer','grant_inventory_editor')
  );

ALTER TABLE authorization_outbox_events
  ADD CONSTRAINT chk_authorization_outbox_events_inventory_required CHECK (
    (kind IN ('grant_inventory_owner','grant_inventory_viewer','grant_inventory_editor') AND inventory_id IS NOT NULL)
    OR (kind = 'grant_tenant_owner' AND inventory_id IS NULL)
  );

CREATE TABLE inventory_access_grants (
  tenant_id varchar(26) NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  inventory_id varchar(26) NOT NULL REFERENCES inventories(id) ON UPDATE CASCADE ON DELETE CASCADE,
  grant_key varchar(180) NOT NULL,
  principal_id varchar(128) NOT NULL,
  relationship varchar(32) NOT NULL,
  created_at timestamptz,
  updated_at timestamptz,
  PRIMARY KEY (tenant_id, inventory_id, grant_key),
  CONSTRAINT chk_inventory_access_grants_relationship CHECK (relationship IN ('viewer','editor')),
  CONSTRAINT chk_inventory_access_grants_key CHECK (grant_key = principal_id || ':' || relationship)
);

CREATE INDEX idx_inventory_access_grants_inventory ON inventory_access_grants(tenant_id, inventory_id);
CREATE INDEX idx_inventory_access_grants_principal_id ON inventory_access_grants(principal_id);
