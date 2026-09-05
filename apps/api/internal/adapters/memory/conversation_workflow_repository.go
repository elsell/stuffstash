package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"maps"
	"time"
)

type workflowKey struct {
	tenant   tenant.ID
	workflow agentmodel.WorkflowID
}
type workflowRevisionKey struct {
	workflowKey
	revision agentmodel.WorkflowRevisionID
}

func (s *Store) WorkflowHead(_ context.Context, tenantID tenant.ID, id agentmodel.WorkflowID) (ports.WorkflowHeadRecord, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, found := s.workflowHeads[workflowKey{tenantID, id}]
	return value, found, nil
}
func (s *Store) WorkflowRevision(_ context.Context, tenantID tenant.ID, workflowID agentmodel.WorkflowID, id agentmodel.WorkflowRevisionID) (agentmodel.WorkflowRevision, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, found := s.workflowRevisions[workflowRevisionKey{workflowKey{tenantID, workflowID}, id}]
	return value, found, nil
}
func (s *Store) AppendWorkflowRevision(ctx context.Context, revision agentmodel.WorkflowRevision, expected int, record audit.Record) error {
	snapshot := revision.Snapshot()
	if _, err := agentmodel.NewWorkflowRevision(snapshot); err != nil {
		return err
	}
	if expected < 0 || snapshot.Number != expected+1 {
		return ports.ErrWorkflowConflict
	}
	if record.TenantID.String() != string(snapshot.TenantID) || record.Action != audit.ActionConversationWorkflowRevisionCreated {
		return agentmodel.ErrInvalidWorkflowRevision
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	key := workflowKey{tenant.ID(snapshot.TenantID), snapshot.WorkflowID}
	if _, exists := s.tenants[key.tenant]; !exists {
		return ports.ErrForbidden
	}
	head, found := s.workflowHeads[key]
	if (expected == 0 && found) || (expected > 0 && (!found || head.LatestRevision != expected)) {
		return ports.ErrWorkflowConflict
	}
	revisionKey := workflowRevisionKey{key, snapshot.ID}
	if _, exists := s.workflowRevisions[revisionKey]; exists {
		return ports.ErrWorkflowConflict
	}
	if _, exists := s.auditRecords[record.ID]; exists {
		return ports.ErrWorkflowConflict
	}
	if record.ID == "" || record.InventoryID != "" {
		return agentmodel.ErrInvalidWorkflowRevision
	}
	if !found {
		head = ports.WorkflowHeadRecord{TenantID: key.tenant, ID: key.workflow, CreatedAt: snapshot.CreatedAt}
	}
	head.LatestRevision = snapshot.Number
	head.UpdatedAt = snapshot.CreatedAt
	if s.workflowHeads == nil {
		s.workflowHeads = map[workflowKey]ports.WorkflowHeadRecord{}
	}
	if s.workflowRevisions == nil {
		s.workflowRevisions = map[workflowRevisionKey]agentmodel.WorkflowRevision{}
	}
	record.Metadata = maps.Clone(record.Metadata)
	s.workflowHeads[key] = head
	s.workflowRevisions[revisionKey] = revision
	s.auditRecords[record.ID] = record
	return nil
}
func (s *Store) ActivateWorkflowRevision(ctx context.Context, tenantID tenant.ID, workflowID agentmodel.WorkflowID, revisionID, expectedActive agentmodel.WorkflowRevisionID, at time.Time, record audit.Record) error {
	if at.IsZero() || record.TenantID.String() != tenantID.String() || record.Action != audit.ActionConversationWorkflowActivated || record.ID == "" || record.InventoryID != "" {
		return agentmodel.ErrInvalidWorkflowRevision
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	key := workflowKey{tenantID, workflowID}
	if _, found := s.workflowRevisions[workflowRevisionKey{key, revisionID}]; !found {
		return ports.ErrWorkflowNotFound
	}
	head, found := s.workflowHeads[key]
	if !found || head.ActiveRevisionID != expectedActive {
		return ports.ErrWorkflowConflict
	}
	if _, exists := s.auditRecords[record.ID]; exists {
		return ports.ErrWorkflowConflict
	}
	head.ActiveRevisionID = revisionID
	head.UpdatedAt = at
	record.Metadata = maps.Clone(record.Metadata)
	s.workflowHeads[key] = head
	s.auditRecords[record.ID] = record
	return nil
}
