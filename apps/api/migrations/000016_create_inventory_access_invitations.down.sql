DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM audit_records
    WHERE action IN ('inventory_invitation.created','inventory_invitation.accepted','inventory_invitation.revoked')
       OR target_type = 'inventory_invitation'
  ) THEN
    RAISE EXCEPTION 'cannot roll back inventory access invitations while invitation audit records exist';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM inventory_access_invitations
  ) THEN
    RAISE EXCEPTION 'cannot roll back inventory access invitations while invitation rows exist';
  END IF;
END $$;

DROP TABLE inventory_access_invitations;

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

ALTER TABLE audit_records DROP CONSTRAINT chk_audit_records_target_type;
ALTER TABLE audit_records ADD CONSTRAINT chk_audit_records_target_type CHECK (
  target_type IN ('tenant','inventory','inventory_access_grant','custom_asset_type','custom_field_definition','asset')
);
