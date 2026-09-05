CREATE TABLE conversation_workflow_selections (
 tenant_id TEXT PRIMARY KEY,
 workflow_id TEXT NOT NULL,
 revision_id TEXT NOT NULL,
 activated_at TIMESTAMP WITH TIME ZONE NOT NULL,
 FOREIGN KEY (tenant_id, workflow_id, revision_id) REFERENCES conversation_workflow_revisions (tenant_id, workflow_id, id) ON UPDATE CASCADE ON DELETE RESTRICT
);
