DROP TABLE conversation_workflow_revisions;
DROP TABLE conversation_workflows;
-- Retain the expanded audit-action constraint so historical workflow audit
-- records survive rollback without deletion or invalidating the constraint.
