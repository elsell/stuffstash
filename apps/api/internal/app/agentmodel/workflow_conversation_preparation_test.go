package agentmodel

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type workflowConversationProvider struct {
	calls        int
	instructions string
}

func (p *workflowConversationProvider) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	p.calls++
	p.instructions = input.Instructions
	return ports.ConversationModelTurn{Text: "Yes, you have chemicals, including Acetone."}, nil
}

type workflowConversationResolver struct {
	provider *workflowConversationProvider
	calls    int
}

func (r *workflowConversationResolver) ResolveWorkflowLanguageProvider(_ context.Context, input ports.WorkflowLanguageProviderResolutionInput) (ports.WorkflowLanguageProviderBinding, error) {
	r.calls++
	if input.TenantID != "tenant-home" || input.ProfileID != "primary-model" {
		return ports.WorkflowLanguageProviderBinding{}, ports.ErrInvalidProviderInput
	}
	return ports.WorkflowLanguageProviderBinding{ProfileID: input.ProfileID, PromptTemplate: "Speak concisely.", Provider: r.provider}, nil
}
func TestSelectedWorkflowUsesOneNativeModelAndPreservesGuidanceAndBudgets(t *testing.T) {
	clock := &workflowExecutionClock{now: time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)}
	service := NewConversationWorkflowService(ConversationWorkflowDependencies{Authorizer: workflowViewAuthorizer{}, Repository: newWorkflowFakeRepository(), Profiles: newFakeProviderProfileRepository(), IDs: &workflowSequenceIDs{}, Clock: clock, Limits: workflowServiceLimits()})
	revision, err := service.SaveRevision(context.Background(), workflowServiceInput())
	if err != nil {
		t.Fatal(err)
	}
	snapshot := revision.Snapshot()
	settings := snapshot.Definition.Settings()
	settings.Budget.ModelCalls = 2
	settings.Response = domain.WorkflowResponseGrounded
	settings.Steps[0].ProviderProfileID = "primary-model"
	settings.Steps[0].Instructions = "Use assigned tags."
	settings.Steps[1].ProviderProfileID = "retired-assessment-model"
	settings.Steps[2].ProviderProfileID = "retired-response-model"
	snapshot.Definition, err = domain.NewWorkflowDefinition(settings, workflowServiceLimits())
	if err != nil {
		t.Fatal(err)
	}
	revision, err = domain.NewWorkflowRevision(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	selected := &SelectedWorkflow{revision: revision, limits: workflowServiceLimits(), clock: clock}
	provider := &workflowConversationProvider{}
	resolver := &workflowConversationResolver{provider: provider}
	prepared, err := selected.Prepare(context.Background(), ports.RealtimeVoiceProviderSet{}, resolver)
	if err != nil {
		t.Fatalf("retired workflow stages blocked native preparation: %v", err)
	}
	accessor, ok := any(prepared).(interface {
		ConversationModel() ports.ConversationModel
	})
	if !ok || accessor.ConversationModel() == nil {
		t.Fatal("selected workflow cannot supply a native conversation model")
	}
	model := accessor.ConversationModel()
	clock.now = clock.now.Add(time.Hour)
	for i := 0; i < 2; i++ {
		turn, err := model.Converse(context.Background(), ports.ConversationModelInput{Instructions: "Use authorized inventory tools."})
		if err != nil || turn.Text != "Yes, you have chemicals, including Acetone." {
			t.Fatalf("model answer lost: %+v err=%v", turn, err)
		}
	}
	if resolver.calls != 1 || provider.calls != 2 || !strings.Contains(provider.instructions, "Speak concisely.") || !strings.Contains(provider.instructions, "Use assigned tags.") || !strings.Contains(provider.instructions, "Use authorized inventory tools.") {
		t.Fatalf("primary model or guidance lost: resolutions=%d provider=%+v", resolver.calls, provider)
	}
	if _, err := model.Converse(context.Background(), ports.ConversationModelInput{}); !errors.Is(err, ErrWorkflowBudgetExhausted) || provider.calls != 2 {
		t.Fatalf("configured call budget reset or bypassed: calls=%d err=%v", provider.calls, err)
	}
}

func TestNativeWorkflowBudgetDoesNotSpendOnCancellationAndExpiresAcrossUtterances(t *testing.T) {
	clock := &workflowExecutionClock{now: time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)}
	provider := &workflowConversationProvider{}
	model, err := newWorkflowConversationModel(provider, clock, workflowServiceInput().Definition, "Be concise.")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := model.Converse(ctx, ports.ConversationModelInput{}); !errors.Is(err, context.Canceled) || provider.calls != 0 {
		t.Fatalf("cancelled invocation reached provider: %v calls=%d", err, provider.calls)
	}
	clock.now = clock.now.Add(time.Hour)
	if _, err := model.Converse(context.Background(), ports.ConversationModelInput{}); err != nil {
		t.Fatalf("cancelled invocation started elapsed budget: %v", err)
	}
	clock.now = clock.now.Add(time.Minute)
	if model.CanContinue() {
		t.Fatal("expired session advertised more capacity")
	}
	if _, err := model.Converse(context.Background(), ports.ConversationModelInput{}); !errors.Is(err, ErrWorkflowBudgetExhausted) || provider.calls != 1 {
		t.Fatalf("elapsed budget was reset or bypassed: %v calls=%d", err, provider.calls)
	}
}

func TestSelectedWorkflowRejectsMissingModelBeforeCapture(t *testing.T) {
	clock := &workflowExecutionClock{now: time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)}
	service := NewConversationWorkflowService(ConversationWorkflowDependencies{Authorizer: workflowViewAuthorizer{}, Repository: newWorkflowFakeRepository(), Profiles: newFakeProviderProfileRepository(), IDs: &workflowSequenceIDs{}, Clock: clock, Limits: workflowServiceLimits()})
	revision, err := service.SaveRevision(context.Background(), workflowServiceInput())
	if err != nil {
		t.Fatal(err)
	}
	selected := &SelectedWorkflow{revision: revision, limits: workflowServiceLimits(), clock: clock}
	prepared, err := selected.Prepare(context.Background(), ports.RealtimeVoiceProviderSet{}, nil)
	if err == nil || prepared != nil {
		t.Fatal("retired-only provider activated the old workflow execution path")
	}
}

type workflowExecutionClock struct{ now time.Time }

func (c *workflowExecutionClock) Now() time.Time { return c.now }
