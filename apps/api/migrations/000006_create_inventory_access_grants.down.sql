DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM authorization_outbox_events
    WHERE kind IN ('grant_inventory_viewer','grant_inventory_editor')
  ) THEN
    RAISE EXCEPTION 'cannot roll back inventory access grants while viewer/editor authorization outbox events exist';
  END IF;
END $$;

DROP TABLE inventory_access_grants;

ALTER TABLE authorization_outbox_events DROP CONSTRAINT chk_authorization_outbox_events_kind;
ALTER TABLE authorization_outbox_events DROP CONSTRAINT chk_authorization_outbox_events_inventory_required;

ALTER TABLE authorization_outbox_events
  ADD CONSTRAINT chk_authorization_outbox_events_kind CHECK (
    kind IN ('grant_tenant_owner','grant_inventory_owner')
  );

ALTER TABLE authorization_outbox_events
  ADD CONSTRAINT chk_authorization_outbox_events_inventory_required CHECK (
    (kind = 'grant_inventory_owner' AND inventory_id IS NOT NULL)
    OR (kind = 'grant_tenant_owner' AND inventory_id IS NULL)
  );
