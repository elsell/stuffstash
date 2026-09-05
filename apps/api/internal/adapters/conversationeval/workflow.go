package conversationeval

import (
	"context"
	"time"

	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// This scoped read adapter selects an immutable candidate only inside an
// isolated runtime. It cannot activate or append any production workflow.
type pinnedWorkflow struct{ revision domain.WorkflowRevision }

func (p pinnedWorkflow) SelectedWorkflowRevision(_ context.Context, tenantID tenant.ID) (ports.WorkflowSelectionReference, bool, error) {
	snapshot := p.revision.Snapshot()
	if snapshot.TenantID != domain.TenantID(tenantID) {
		return ports.WorkflowSelectionReference{}, false, nil
	}
	return ports.WorkflowSelectionReference{WorkflowID: snapshot.WorkflowID, RevisionID: snapshot.ID}, true, nil
}
func (p pinnedWorkflow) WorkflowRevision(_ context.Context, tenantID tenant.ID, workflowID domain.WorkflowID, revisionID domain.WorkflowRevisionID) (domain.WorkflowRevision, bool, error) {
	snapshot := p.revision.Snapshot()
	if snapshot.TenantID != domain.TenantID(tenantID) || snapshot.WorkflowID != workflowID || snapshot.ID != revisionID {
		return domain.WorkflowRevision{}, false, nil
	}
	return p.revision, true, nil
}
func (p pinnedWorkflow) WorkflowHead(_ context.Context, tenantID tenant.ID, workflowID domain.WorkflowID) (ports.WorkflowHeadRecord, bool, error) {
	snapshot := p.revision.Snapshot()
	if snapshot.TenantID != domain.TenantID(tenantID) || snapshot.WorkflowID != workflowID {
		return ports.WorkflowHeadRecord{}, false, nil
	}
	return ports.WorkflowHeadRecord{TenantID: tenantID, ID: workflowID, LatestRevision: snapshot.Number, ActiveRevisionID: snapshot.ID, CreatedAt: snapshot.CreatedAt, UpdatedAt: snapshot.CreatedAt}, true, nil
}
func (p pinnedWorkflow) AppendWorkflowRevision(context.Context, domain.WorkflowRevision, int, audit.Record) error {
	return ErrInvalidExecution
}
func (p pinnedWorkflow) ActivateWorkflowRevision(context.Context, tenant.ID, domain.WorkflowID, domain.WorkflowRevisionID, ports.WorkflowSelectionReference, time.Time, audit.Record) error {
	return ErrInvalidExecution
}
