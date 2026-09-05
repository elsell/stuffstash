package agentmodel

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func workflowServiceInput() SaveConversationWorkflowInput {
	return SaveConversationWorkflowInput{Principal: testPrincipal(), TenantID: "tenant-home", Source: audit.SourceAPI, Definition: domain.WorkflowDefinitionInput{Name: "Household", Retrieval: domain.WorkflowRetrievalPreciseFirst, Response: domain.WorkflowResponseGroundedFallback, Budget: domain.WorkflowBudget{EvidenceRounds: 2, ModelCalls: 8, ElapsedSeconds: 60, FollowUpTurns: 4}, Steps: []domain.WorkflowStep{{Kind: domain.WorkflowStepInterpret, Attempts: 1}, {Kind: domain.WorkflowStepAssess, Attempts: 1}, {Kind: domain.WorkflowStepRespond, Attempts: 1}}}}
}
func workflowServiceLimits() domain.WorkflowLimits {
	return domain.WorkflowLimits{Budget: domain.WorkflowBudget{EvidenceRounds: 6, ModelCalls: 20, ElapsedSeconds: 180, FollowUpTurns: 12}, MaxStepAttempts: 3, MaxNameRunes: 120, MaxInstructionRunes: 4000}
}

func TestConversationWorkflowSaveAuthorizesBeforeMutation(t *testing.T) {
	repository := newWorkflowFakeRepository()
	service := NewConversationWorkflowService(ConversationWorkflowDependencies{Authorizer: denyTenantAuthorizer{}, Repository: repository, Profiles: newFakeProviderProfileRepository(), IDs: fixedIDGenerator{}, Clock: fixedClock{}, Limits: workflowServiceLimits()})
	if _, err := service.SaveRevision(context.Background(), workflowServiceInput()); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("unauthorized save: %v", err)
	}
	if len(repository.revisions) != 0 || len(repository.audit) != 0 {
		t.Fatal("denied save changed state")
	}
}

func TestConversationWorkflowSaveSnapshotsAndAudits(t *testing.T) {
	repository := newWorkflowFakeRepository()
	service := NewConversationWorkflowService(ConversationWorkflowDependencies{Authorizer: allowTenantConfigureAuthorizer{}, Repository: repository, Profiles: newFakeProviderProfileRepository(), IDs: fixedIDGenerator{}, Clock: fixedClock{}, Limits: workflowServiceLimits()})
	input := workflowServiceInput()
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
}

func TestConversationWorkflowRejectsUnknownOrWrongCapabilityProvider(t *testing.T) {
	repository := newWorkflowFakeRepository()
	profiles := newFakeProviderProfileRepository()
	service := NewConversationWorkflowService(ConversationWorkflowDependencies{Authorizer: allowTenantConfigureAuthorizer{}, Repository: repository, Profiles: profiles, IDs: fixedIDGenerator{}, Clock: fixedClock{}, Limits: workflowServiceLimits()})
	for _, profile := range []domain.ProviderProfile{
		mustProviderProfile(t, "provider-one", "tenant-other", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileEnabled),
		mustProviderProfile(t, "provider-one", "tenant-home", domain.ProviderCapabilitySpeechToText, domain.ProviderProfileEnabled),
	} {
		profiles.saved["provider-one"] = profile
		input := workflowServiceInput()
		input.Definition.Steps[0].ProviderProfileID = "provider-one"
		if _, err := service.SaveRevision(context.Background(), input); !errors.Is(err, apperrors.ErrValidation) {
			t.Fatalf("invalid provider selection: %v", err)
		}
	}
	if len(repository.revisions) != 0 {
		t.Fatal("invalid provider saved workflow")
	}
}

type workflowFakeRepository struct {
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
	r.revisions[key+"/"+string(s.ID)] = v
	r.heads[key] = ports.WorkflowHeadRecord{TenantID: tenant.ID(s.TenantID), ID: s.WorkflowID, LatestRevision: s.Number, ActiveRevisionID: head.ActiveRevisionID, CreatedAt: s.CreatedAt, UpdatedAt: s.CreatedAt}
	r.audit = append(r.audit, a)
	return nil
}
func (r *workflowFakeRepository) ActivateWorkflowRevision(_ context.Context, t tenant.ID, w domain.WorkflowID, id, expected domain.WorkflowRevisionID, at time.Time, a audit.Record) error {
	key := t.String() + "/" + string(w)
	head, ok := r.heads[key]
	if !ok {
		return ports.ErrWorkflowNotFound
	}
	if _, ok := r.revisions[key+"/"+string(id)]; !ok {
		return ports.ErrWorkflowNotFound
	}
	if head.ActiveRevisionID != expected {
		return ports.ErrWorkflowConflict
	}
	head.ActiveRevisionID = id
	head.UpdatedAt = at
	r.heads[key] = head
	r.audit = append(r.audit, a)
	return nil
}
