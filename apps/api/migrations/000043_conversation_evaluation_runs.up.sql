ALTER TABLE audit_records DROP CONSTRAINT chk_audit_records_action;
ALTER TABLE audit_records ADD CONSTRAINT chk_audit_records_action CHECK (action IN (
  'tenant.created','tenant.viewed','tenant.listed','tenant.updated','tenant.archived','tenant.restored','tenant.deleted',
  'inventory.created','inventory.viewed','inventory.listed','inventory.updated','inventory.archived','inventory.restored','inventory.deleted',
  'inventory_access.granted','inventory_access_grant.viewed','inventory_access_grant.listed','inventory_access.revoked',
  'inventory_invitation.created','inventory_invitation.viewed','inventory_invitation.listed','inventory_invitation.accepted','inventory_invitation.expiration_updated','inventory_invitation.revoked','inventory_invitation.cancelled','inventory_invitation.deleted',
  'custom_asset_type.created','custom_asset_type.viewed','custom_asset_type.listed','custom_asset_type.updated','custom_asset_type.archived','custom_asset_type.restored','custom_asset_type.deleted',
  'custom_field_definition.created','custom_field_definition.viewed','custom_field_definition.listed','custom_field_definition.updated','custom_field_definition.archived','custom_field_definition.restored','custom_field_definition.deleted',
  'asset.created','asset.viewed','asset.listed','asset.searched','asset.updated','asset.moved','asset.archived','asset.restored','asset.deleted','asset.checked_out','asset.returned','asset.return_details_updated',
  'asset_tag.created','asset_tag.listed','asset_tag.updated','asset_tag.archived',
  'attachment.created','attachment.viewed','attachment.listed','attachment.content_downloaded','attachment.archived','attachment.restored','attachment.deleted',
  'audit_record.listed','undoable_operation.undone','undoable_operation.redone',
  'provider_profile.created','provider_profile.viewed','provider_profile.listed','provider_profile.updated','provider_profile.enabled','provider_profile.disabled','provider_profile.archived','provider_profile.credential_replaced','provider_profile.tested',
  'voice_provider_configuration.updated',
  'conversation_workflow.revision_created','conversation_workflow.activated','conversation_evaluation_case.revision_created','conversation_evaluation_case.viewed','conversation_evaluation_case.listed','conversation_evaluation_run.created','conversation_evaluation_run.progressed','conversation_evaluation_run.cancelled',
  'import_job.previewed','import_job.started','import_job.completed','import_job.failed','import_job.cancellation_requested','import_job.cancelled','import_job.history_removed','import_job.credential_cleaned'
));

ALTER TABLE audit_records DROP CONSTRAINT chk_audit_records_target_type;
ALTER TABLE audit_records ADD CONSTRAINT chk_audit_records_target_type CHECK (
 target_type IN ('tenant','inventory','inventory_access_grant','inventory_invitation',
 'custom_asset_type','custom_field_definition','asset','asset_tag','attachment',
 'audit_record','undoable_operation','provider_profile','import_job',
 'conversation_workflow','conversation_evaluation_case','conversation_evaluation_run')
);

CREATE TABLE conversation_evaluation_runs (
 tenant_id VARCHAR(64) NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
 id VARCHAR(64) NOT NULL,
 state VARCHAR(32) NOT NULL CHECK (state IN ('queued','running','succeeded','failed','cancelled')),
 version BIGINT NOT NULL CHECK (version > 0),
 lease_until TIMESTAMPTZ,
 workflow_id VARCHAR(64) NOT NULL,
 revision_id VARCHAR(64) NOT NULL,
 total_cases INTEGER NOT NULL CHECK (total_cases BETWEEN 1 AND 100),
 completed_cases INTEGER NOT NULL CHECK (completed_cases >= 0 AND completed_cases <= total_cases),
 passed_cases INTEGER NOT NULL CHECK (passed_cases >= 0 AND passed_cases <= completed_cases),
 input_json TEXT NOT NULL,
 progress_json TEXT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL,
 updated_at TIMESTAMPTZ NOT NULL,
 PRIMARY KEY (tenant_id,id)
);
CREATE INDEX evaluation_run_queue ON conversation_evaluation_runs(state,lease_until);
