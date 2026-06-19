CREATE TABLE authorization_outbox_events (
  id varchar(26) PRIMARY KEY,
  kind varchar(80) NOT NULL,
  principal_id varchar(128) NOT NULL,
  tenant_id varchar(26) NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  inventory_id varchar(26) REFERENCES inventories(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  attempts integer NOT NULL DEFAULT 0,
  last_error text NOT NULL DEFAULT '',
  processed_at timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT chk_authorization_outbox_events_kind CHECK (kind IN ('grant_tenant_owner', 'grant_inventory_owner')),
  CONSTRAINT chk_authorization_outbox_events_inventory_required CHECK (
    (kind = 'grant_inventory_owner' AND inventory_id IS NOT NULL)
    OR (kind = 'grant_tenant_owner' AND inventory_id IS NULL)
  )
);

CREATE INDEX idx_authorization_outbox_events_kind ON authorization_outbox_events(kind);
CREATE INDEX idx_authorization_outbox_events_principal_id ON authorization_outbox_events(principal_id);
CREATE INDEX idx_authorization_outbox_events_tenant_id ON authorization_outbox_events(tenant_id);
CREATE INDEX idx_authorization_outbox_events_inventory_id ON authorization_outbox_events(inventory_id);
CREATE INDEX idx_authorization_outbox_events_processed_at ON authorization_outbox_events(processed_at);
