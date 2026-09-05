package memory

import (
	"context"
	"maps"
	"slices"
	"strings"

	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type evaluationCaseKey struct {
	tenant tenant.ID
	id     domain.EvaluationCaseID
}
type evaluationCaseRevisionKey struct {
	evaluationCaseKey
	revision domain.EvaluationCaseRevisionID
}

func (s *Store) EvaluationCaseHead(ctx context.Context, tenantID tenant.ID, id domain.EvaluationCaseID) (ports.EvaluationCaseHeadRecord, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return ports.EvaluationCaseHeadRecord{}, false, err
	}
	value, found := s.evaluationCaseHeads[evaluationCaseKey{tenantID, id}]
	return value, found, nil
}
func (s *Store) EvaluationCaseRevision(ctx context.Context, tenantID tenant.ID, caseID domain.EvaluationCaseID, id domain.EvaluationCaseRevisionID) (domain.EvaluationCaseRevision, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return domain.EvaluationCaseRevision{}, false, err
	}
	value, found := s.evaluationCaseRevisions[evaluationCaseRevisionKey{evaluationCaseKey{tenantID, caseID}, id}]
	return value, found, nil
}
func (s *Store) ListEvaluationCases(ctx context.Context, tenantID tenant.ID, page ports.EvaluationCasePageRequest) ([]ports.EvaluationCaseHeadRecord, error) {
	if page.Limit < 1 || page.Limit > ports.MaxEvaluationCasePageLimit {
		return nil, ports.ErrInvalidEvaluationCasePage
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	result := []ports.EvaluationCaseHeadRecord{}
	for key, head := range s.evaluationCaseHeads {
		if key.tenant == tenantID && string(key.id) > string(page.AfterID) {
			result = append(result, head)
		}
	}
	slices.SortFunc(result, func(a, b ports.EvaluationCaseHeadRecord) int { return strings.Compare(string(a.ID), string(b.ID)) })
	if len(result) > page.Limit {
		result = result[:page.Limit]
	}
	return result, nil
}
func (s *Store) AppendEvaluationCaseRevision(ctx context.Context, revision domain.EvaluationCaseRevision, expected int, record audit.Record) error {
	snapshot := revision.Snapshot()
	if _, err := domain.NewEvaluationCaseRevision(snapshot); err != nil {
		return err
	}
	if expected < 0 || snapshot.Number != expected+1 {
		return ports.ErrEvaluationCaseConflict
	}
	if record.TenantID.String() != string(snapshot.TenantID) || record.InventoryID != "" || record.ID == "" || record.Action != audit.ActionConversationEvaluationCaseRevisionCreated {
		return domain.ErrInvalidEvaluationCaseRevision
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	key := evaluationCaseKey{tenant.ID(snapshot.TenantID), snapshot.CaseID}
	if _, exists := s.tenants[key.tenant]; !exists {
		return ports.ErrForbidden
	}
	head, found := s.evaluationCaseHeads[key]
	if (expected == 0 && found) || (expected > 0 && (!found || head.LatestRevision != expected)) {
		return ports.ErrEvaluationCaseConflict
	}
	revisionKey := evaluationCaseRevisionKey{key, snapshot.ID}
	if _, exists := s.evaluationCaseRevisions[revisionKey]; exists {
		return ports.ErrEvaluationCaseConflict
	}
	if _, exists := s.auditRecords[record.ID]; exists {
		return ports.ErrEvaluationCaseConflict
	}
	if !found {
		head = ports.EvaluationCaseHeadRecord{TenantID: key.tenant, ID: key.id, CreatedAt: snapshot.CreatedAt}
	}
	head.Title = snapshot.Definition.Settings().Title
	head.LatestRevision = snapshot.Number
	head.LatestRevisionID = snapshot.ID
	head.UpdatedAt = snapshot.CreatedAt
	if s.evaluationCaseHeads == nil {
		s.evaluationCaseHeads = map[evaluationCaseKey]ports.EvaluationCaseHeadRecord{}
	}
	if s.evaluationCaseRevisions == nil {
		s.evaluationCaseRevisions = map[evaluationCaseRevisionKey]domain.EvaluationCaseRevision{}
	}
	record.Metadata = maps.Clone(record.Metadata)
	s.evaluationCaseHeads[key] = head
	s.evaluationCaseRevisions[revisionKey] = revision
	s.auditRecords[record.ID] = record
	return nil
}
