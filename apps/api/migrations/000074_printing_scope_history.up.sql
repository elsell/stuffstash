-- Printing history and pending authorization removals outlive deleted scopes.
-- Application deletion atomically revokes/retire records before parent removal.
ALTER TABLE printers DROP CONSTRAINT printers_tenant_id_fkey, DROP CONSTRAINT printers_inventory_id_fkey;
ALTER TABLE print_connectors DROP CONSTRAINT print_connectors_tenant_id_fkey, DROP CONSTRAINT print_connectors_inventory_id_fkey;
ALTER TABLE print_jobs DROP CONSTRAINT print_jobs_tenant_id_fkey, DROP CONSTRAINT print_jobs_inventory_id_fkey;
ALTER TABLE inventory_print_settings DROP CONSTRAINT inventory_print_settings_tenant_id_fkey, DROP CONSTRAINT inventory_print_settings_inventory_id_fkey;
