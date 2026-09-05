package agentmodel

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type workflowStepResolver struct {
	calls     int
	profileID string
	provider  *workflowExecutionProvider
}

func (r *workflowStepResolver) ResolveWorkflowLanguageProvider(_ context.Context, input ports.WorkflowLanguageProviderResolutionInput) (ports.WorkflowLanguageProviderBinding, error) {
	r.calls++
	if input.TenantID != "tenant-home" || input.ProfileID != r.profileID {
		return ports.WorkflowLanguageProviderBinding{}, ports.ErrInvalidProviderInput
	}
	return ports.WorkflowLanguageProviderBinding{ProfileID: r.profileID, PromptTemplate: "Configured provider guidance.", Provider: r.provider}, nil
}

func TestWorkflowPreparationPinsRevisionProvidersAndStartsBudgetLazily(t *testing.T) {
	ctx := context.Background()
	repository := newWorkflowFakeRepository()
	clock := &workflowExecutionClock{now: time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)}
	service := NewConversationWorkflowService(ConversationWorkflowDependencies{Authorizer: allowTenantConfigureAuthorizer{}, Repository: repository, Profiles: newFakeProviderProfileRepository(), IDs: &workflowSequenceIDs{}, Clock: clock, Limits: workflowServiceLimits()})
	saved, err := service.SaveRevision(ctx, workflowServiceInput())
	if err != nil {
		t.Fatal(err)
	}
	snapshot := saved.Snapshot()
	settings := snapshot.Definition.Settings()
	for i := range settings.Steps {
		settings.Steps[i].ProviderProfileID = "step-model"
		settings.Steps[i].Instructions = "Use assigned tags."
	}
	snapshot.Definition, err = domain.NewWorkflowDefinition(settings, workflowServiceLimits())
	if err != nil {
		t.Fatal(err)
	}
	saved, err = domain.NewWorkflowRevision(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	repository.revisions["tenant-home/"+string(snapshot.WorkflowID)+"/"+string(snapshot.ID)] = saved
	repository.selected = map[tenant.ID]ports.WorkflowSelectionReference{"tenant-home": {WorkflowID: snapshot.WorkflowID, RevisionID: snapshot.ID}}
	provider := &workflowExecutionProvider{}
	resolver := &workflowStepResolver{profileID: "step-model", provider: provider}
	prepared, err := service.PrepareSelected(ctx, PrepareWorkflowInput{Principal: testPrincipal(), TenantID: "tenant-home", Resolver: resolver})
	if err != nil {
		t.Fatal(err)
	}
	if prepared == nil || prepared.Revision().Snapshot().ID != snapshot.ID || resolver.calls != 1 {
		t.Fatalf("revision or provider not pinned: %v calls=%d", prepared, resolver.calls)
	}
	delete(repository.selected, "tenant-home")
	clock.now = clock.now.Add(time.Hour)
	if _, err = prepared.NextTurn(ctx, ports.LanguageInferenceInput{Investigation: &domain.InvestigationInput{Phase: domain.InvestigationPhaseInitial}, PromptTemplate: "Wrong default guidance"}); err != nil {
		t.Fatal(err)
	}
	if provider.prompts[0] != "Configured provider guidance." || provider.instructions[0] != "Use assigned tags." {
		t.Fatal("selected workflow did not reach provider")
	}
	clock.now = clock.now.Add(time.Hour)
	if _, err = prepared.NextTurn(ctx, ports.LanguageInferenceInput{Investigation: &domain.InvestigationInput{Phase: domain.InvestigationPhaseEvidenceAssessment}}); !errors.Is(err, ErrWorkflowBudgetExhausted) {
		t.Fatalf("session did not retain execution budget: %v", err)
	}
	if resolver.calls != 1 {
		t.Fatal("pinned provider was resolved again")
	}
}

func TestWorkflowPreparationRequiresAccessAndFailsClosedOnBrokenSelection(t *testing.T) {
	repository := newWorkflowFakeRepository()
	repository.selected = map[tenant.ID]ports.WorkflowSelectionReference{"tenant-home": {WorkflowID: "missing", RevisionID: "missing"}}
	service := NewConversationWorkflowService(ConversationWorkflowDependencies{Authorizer: denyTenantAuthorizer{}, Repository: repository, Limits: workflowServiceLimits()})
	input := PrepareWorkflowInput{Principal: testPrincipal(), TenantID: "tenant-home"}
	if _, err := service.PrepareSelected(context.Background(), input); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("denial lost: %v", err)
	}
	service.deps.Authorizer = allowTenantConfigureAuthorizer{}
	if _, err := service.PrepareSelected(context.Background(), input); !errors.Is(err, apperrors.ErrPrecondition) {
		t.Fatalf("broken selection silently defaulted: %v", err)
	}
	delete(repository.selected, "tenant-home")
	if prepared, err := service.PrepareSelected(context.Background(), input); err != nil || prepared != nil {
		t.Fatalf("absent selection should use default: %v", err)
	}
}
