ALTER TABLE authorization_outbox_events
  DROP CONSTRAINT chk_authorization_outbox_events_kind,
  DROP CONSTRAINT chk_authorization_outbox_events_inventory_required;

ALTER TABLE authorization_outbox_events
  ADD CONSTRAINT chk_authorization_outbox_events_kind CHECK (
    kind IN ('grant_tenant_owner','grant_inventory_owner','grant_inventory_viewer','grant_inventory_editor','revoke_inventory_viewer','revoke_inventory_editor')
  );

ALTER TABLE authorization_outbox_events
  ADD CONSTRAINT chk_authorization_outbox_events_inventory_required CHECK (
    (kind IN ('grant_inventory_owner','grant_inventory_viewer','grant_inventory_editor','revoke_inventory_viewer','revoke_inventory_editor') AND inventory_id IS NOT NULL)
    OR (kind = 'grant_tenant_owner' AND inventory_id IS NULL)
  );

ALTER TABLE audit_records DROP CONSTRAINT chk_audit_records_action;
ALTER TABLE audit_records ADD CONSTRAINT chk_audit_records_action CHECK (action IN (
    'tenant.created',
    'inventory.created',
    'inventory_access.granted',
    'inventory_access.revoked',
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
