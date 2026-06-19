DROP INDEX IF EXISTS idx_audit_records_inventory_id;
DROP INDEX IF EXISTS idx_audit_records_tenant_id;

ALTER TABLE audit_records
  DROP CONSTRAINT IF EXISTS audit_records_inventory_id_fkey;

ALTER TABLE audit_records
  ADD CONSTRAINT audit_records_inventory_id_fkey
  FOREIGN KEY (inventory_id)
  REFERENCES inventories(id)
  ON UPDATE CASCADE
  ON DELETE CASCADE;

CREATE INDEX idx_audit_records_tenant_id
  ON audit_records(tenant_id, id);

CREATE INDEX idx_audit_records_inventory_id
  ON audit_records(tenant_id, inventory_id, id)
  WHERE inventory_id IS NOT NULL;
