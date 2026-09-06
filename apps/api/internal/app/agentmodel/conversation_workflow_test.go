package agentmodel

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func workflowServiceInput() SaveConversationWorkflowInput {
	return SaveConversationWorkflowInput{Principal: testPrincipal(), TenantID: "tenant-home", Source: audit.SourceAPI, Definition: domain.WorkflowDefinitionInput{Name: "Household", Budget: domain.WorkflowBudget{ToolCalls: 2, ModelCalls: 8, ElapsedSeconds: 60, FollowUpTurns: 4}}}
}
func workflowServiceLimits() domain.WorkflowLimits {
	return domain.WorkflowLimits{Budget: domain.WorkflowBudget{ToolCalls: 6, ModelCalls: 20, ElapsedSeconds: 180, FollowUpTurns: 12}, MaxNameRunes: 120, MaxInstructionRunes: 4000}
}

func TestConversationWorkflowSaveAuthorizesBeforeMutation(t *testing.T) {
	repository := newWorkflowFakeRepository()
	service := NewConversationWorkflowService(ConversationWorkflowDependencies{Authorizer: denyTenantAuthorizer{}, Repository: repository, Profiles: newFakeProviderProfileRepository(), IDs: &workflowSequenceIDs{}, Clock: fixedClock{}, Limits: workflowServiceLimits()})
	if _, err := service.SaveRevision(context.Background(), workflowServiceInput()); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("unauthorized save: %v", err)
	}
	if len(repository.revisions) != 0 || len(repository.audit) != 0 {
		t.Fatal("denied save changed state")
	}
}

func TestConversationWorkflowSaveSnapshotsAndAudits(t *testing.T) {
	repository := newWorkflowFakeRepository()
	service := NewConversationWorkflowService(ConversationWorkflowDependencies{Authorizer: allowTenantConfigureAuthorizer{}, Repository: repository, Profiles: newFakeProviderProfileRepository(), IDs: &workflowSequenceIDs{}, Clock: fixedClock{}, Limits: workflowServiceLimits()})
	input := workflowServiceInput()
	input.Principal.ID = "john.smith"
	revision, err := service.SaveRevision(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	got := revision.Snapshot()
	if got.TenantID != "tenant-home" || got.Number != 1 || !got.CreatedAt.Equal((fixedClock{}).Now()) {
		t.Fatalf("bad revision snapshot: %+v", got)
	}
	if len(repository.audit) != 1 || repository.audit[0].Action != audit.ActionConversationWorkflowRevisionCreated {
		t.Fatal("missing revision audit")
	}
	input.WorkflowID = got.WorkflowID
	if _, err := service.SaveRevision(context.Background(), input); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("stale save must conflict: %v", err)
	}
	if len(repository.audit) != 1 {
		t.Fatal("stale save added audit")
	}
	input.ExpectedRevision = 1
	second, err := service.SaveRevision(context.Background(), input)
	if err != nil || second.Snapshot().Number != 2 || second.Snapshot().ID == got.ID {
		t.Fatalf("second immutable revision failed: %v", err)
	}
	original, found, err := repository.WorkflowRevision(context.Background(), input.TenantID, got.WorkflowID, got.ID)
	if err != nil || !found || original.Snapshot().Number != 1 {
		t.Fatal("second save overwrote original")
	}
	// A stale read snapshot must not defeat the authoritative append CAS.
	stale := service
	stale.deps.Repository = workflowStaleHeadView{workflowFakeRepository: repository, head: ports.WorkflowHeadRecord{TenantID: input.TenantID, ID: got.WorkflowID, LatestRevision: 1}}
	if _, err := stale.SaveRevision(context.Background(), input); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("append race did not map to conflict: %v", err)
	}
	if len(repository.audit) != 2 {
		t.Fatal("conflicting append changed audit")
	}

}

func TestConversationWorkflowRejectsUnknownOrWrongCapabilityProvider(t *testing.T) {
	repository := newWorkflowFakeRepository()
	profiles := newFakeProviderProfileRepository()
	service := NewConversationWorkflowService(ConversationWorkflowDependencies{Authorizer: allowTenantConfigureAuthorizer{}, Repository: repository, Profiles: profiles, IDs: &workflowSequenceIDs{}, Clock: fixedClock{}, Limits: workflowServiceLimits()})
	for _, profile := range []domain.ProviderProfile{
		mustProviderProfile(t, "provider-one", "tenant-other", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileEnabled),
		mustProviderProfile(t, "provider-one", "tenant-home", domain.ProviderCapabilitySpeechToText, domain.ProviderProfileEnabled),
	} {
		profiles.saved["provider-one"] = profile
		input := workflowServiceInput()
		input.Definition.ProviderProfileID = "provider-one"
		if _, err := service.SaveRevision(context.Background(), input); !errors.Is(err, apperrors.ErrValidation) {
			t.Fatalf("invalid provider selection: %v", err)
		}
	}
	if len(repository.revisions) != 0 {
		t.Fatal("invalid provider saved workflow")
	}
}

type workflowFakeRepository struct {
	selected  map[tenant.ID]ports.WorkflowSelectionReference
	revisions map[string]domain.WorkflowRevision
	heads     map[string]ports.WorkflowHeadRecord
	audit     []audit.Record
}

func newWorkflowFakeRepository() *workflowFakeRepository {
	return &workflowFakeRepository{revisions: map[string]domain.WorkflowRevision{}, heads: map[string]ports.WorkflowHeadRecord{}}
}
func (r *workflowFakeRepository) WorkflowHead(_ context.Context, t tenant.ID, w domain.WorkflowID) (ports.WorkflowHeadRecord, bool, error) {
	v, ok := r.heads[t.String()+"/"+string(w)]
	return v, ok, nil
}
func (r *workflowFakeRepository) WorkflowRevision(_ context.Context, t tenant.ID, w domain.WorkflowID, id domain.WorkflowRevisionID) (domain.WorkflowRevision, bool, error) {
	v, ok := r.revisions[t.String()+"/"+string(w)+"/"+string(id)]
	return v, ok, nil
}
func (r *workflowFakeRepository) AppendWorkflowRevision(_ context.Context, v domain.WorkflowRevision, expected int, a audit.Record) error {
	s := v.Snapshot()
	key := string(s.TenantID) + "/" + string(s.WorkflowID)
	head := r.heads[key]
	if head.LatestRevision != expected || s.Number != expected+1 {
		return ports.ErrWorkflowConflict
	}

	if _, exists := r.revisions[key+"/"+string(s.ID)]; exists {
		return ports.ErrWorkflowConflict
	}
	for _, existing := range r.audit {
		if existing.ID == a.ID {
			return ports.ErrWorkflowConflict
		}
	}
	r.revisions[key+"/"+string(s.ID)] = v
	createdAt := head.CreatedAt
	if createdAt.IsZero() {
		createdAt = s.CreatedAt
	}
	r.heads[key] = ports.WorkflowHeadRecord{TenantID: tenant.ID(s.TenantID), ID: s.WorkflowID, LatestRevision: s.Number, ActiveRevisionID: head.ActiveRevisionID, CreatedAt: createdAt, UpdatedAt: s.CreatedAt}
	r.audit = append(r.audit, a)
	return nil
}
func (r *workflowFakeRepository) ActivateWorkflowRevision(_ context.Context, t tenant.ID, w domain.WorkflowID, id domain.WorkflowRevisionID, expected ports.WorkflowSelectionReference, at time.Time, a audit.Record) error {
	key := t.String() + "/" + string(w)
	head, ok := r.heads[key]
	if !ok {
		return ports.ErrWorkflowNotFound
	}
	if _, ok := r.revisions[key+"/"+string(id)]; !ok {
		return ports.ErrWorkflowNotFound
	}
	if r.selected[t] != expected {
		return ports.ErrWorkflowConflict
	}
	for _, record := range r.audit {
		if record.ID == a.ID {
			return ports.ErrWorkflowConflict
		}
	}
	if r.selected == nil {
		r.selected = map[tenant.ID]ports.WorkflowSelectionReference{}
	}
	r.selected[t] = ports.WorkflowSelectionReference{WorkflowID: w, RevisionID: id}
	head.ActiveRevisionID = id
	head.UpdatedAt = at
	r.heads[key] = head
	r.audit = append(r.audit, a)
	return nil
}

type workflowSequenceIDs struct{ next int }

func (ids *workflowSequenceIDs) NewID() string {
	ids.next++
	return fmt.Sprintf("workflow-generated-%d", ids.next)
}

type workflowStaleHeadView struct {
	*workflowFakeRepository
	head ports.WorkflowHeadRecord
}

func (v workflowStaleHeadView) WorkflowHead(_ context.Context, t tenant.ID, w domain.WorkflowID) (ports.WorkflowHeadRecord, bool, error) {
	if v.head.TenantID != t || v.head.ID != w {
		return ports.WorkflowHeadRecord{}, false, nil
	}
	return v.head, true, nil
}

func (r *workflowFakeRepository) SelectedWorkflowRevision(_ context.Context, t tenant.ID) (ports.WorkflowSelectionReference, bool, error) {
	value, found := r.selected[t]
	return value, found, nil
}
