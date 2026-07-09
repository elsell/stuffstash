DELETE FROM audit_records
WHERE action IN ('asset.searched','asset.return_details_updated');

ALTER TABLE audit_records DROP CONSTRAINT chk_audit_records_action;
ALTER TABLE audit_records ADD CONSTRAINT chk_audit_records_action CHECK (action IN (
  'tenant.created','tenant.viewed','tenant.listed','tenant.updated','tenant.archived','tenant.restored','tenant.deleted',
  'inventory.created','inventory.viewed','inventory.listed','inventory.updated','inventory.archived','inventory.restored','inventory.deleted',
  'inventory_access.granted','inventory_access_grant.viewed','inventory_access_grant.listed','inventory_access.revoked',
  'inventory_invitation.created','inventory_invitation.viewed','inventory_invitation.listed','inventory_invitation.accepted','inventory_invitation.expiration_updated','inventory_invitation.revoked','inventory_invitation.cancelled','inventory_invitation.deleted',
  'custom_asset_type.created','custom_asset_type.viewed','custom_asset_type.listed','custom_asset_type.updated','custom_asset_type.archived','custom_asset_type.restored','custom_asset_type.deleted',
  'custom_field_definition.created','custom_field_definition.viewed','custom_field_definition.listed','custom_field_definition.updated','custom_field_definition.archived','custom_field_definition.restored','custom_field_definition.deleted',
  'asset.created','asset.viewed','asset.listed','asset.updated','asset.moved','asset.archived','asset.restored','asset.deleted','asset.checked_out','asset.returned',
  'asset_tag.created','asset_tag.listed','asset_tag.updated','asset_tag.archived',
  'attachment.created','attachment.viewed','attachment.listed','attachment.content_downloaded','attachment.archived','attachment.restored','attachment.deleted',
  'audit_record.listed','undoable_operation.undone','undoable_operation.redone',
  'provider_profile.created','provider_profile.viewed','provider_profile.listed','provider_profile.updated','provider_profile.enabled','provider_profile.disabled','provider_profile.archived','provider_profile.credential_replaced','provider_profile.tested',
  'voice_provider_configuration.updated',
  'import_job.previewed','import_job.started','import_job.completed','import_job.failed','import_job.cancellation_requested','import_job.cancelled','import_job.history_removed','import_job.credential_cleaned'
));
