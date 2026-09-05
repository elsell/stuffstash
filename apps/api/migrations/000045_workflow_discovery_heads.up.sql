ALTER TABLE conversation_workflows ADD COLUMN name TEXT;
ALTER TABLE conversation_workflows ADD COLUMN latest_revision_id VARCHAR(64);
UPDATE conversation_workflows AS head
SET name = (revision.snapshot_json::jsonb #>> '{Definition,Name}'),
    latest_revision_id = revision.id
FROM conversation_workflow_revisions AS revision
WHERE revision.tenant_id = head.tenant_id
  AND revision.workflow_id = head.id
  AND revision.number = head.latest_revision;
ALTER TABLE conversation_workflows ALTER COLUMN name SET NOT NULL;
ALTER TABLE conversation_workflows ALTER COLUMN latest_revision_id SET NOT NULL;
