// Package conversationeval composes isolated fixture storage with the production
// conversation application. It cannot access live inventory repositories.
package conversationeval

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"

	"github.com/stuffstash/stuff-stash/internal/app"
	modelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

var ErrInvalidExecution = errors.New("invalid isolated conversation evaluation")
var ErrFixtureMutation = errors.New("conversation evaluation changed fixtures")

type Dependencies struct {
	Clock    ports.Clock
	IDs      ports.IDGenerator
	Observer ports.Observer
}
type Executor struct{ deps Dependencies }

func New(deps Dependencies) *Executor { return &Executor{deps: deps} }

var _ ports.ConversationEvaluationExecutor = (*Executor)(nil)

func (e *Executor) Execute(ctx context.Context, input ports.ConversationEvaluationInput) (ports.ConversationEvaluationResult, error) {
	if err := ctx.Err(); err != nil {
		return ports.ConversationEvaluationResult{}, err
	}
	if e.deps.Clock == nil || e.deps.IDs == nil || strings.TrimSpace(input.Principal.ID.String()) == "" {
		return ports.ConversationEvaluationResult{}, ErrInvalidExecution
	}
	if _, err := domain.NewEvaluationCaseDefinition(input.Case.Settings()); err != nil {
		return ports.ConversationEvaluationResult{}, ErrInvalidExecution
	}
	if _, err := domain.NewWorkflowRevision(input.Revision.Snapshot()); err != nil {
		return ports.ConversationEvaluationResult{}, ErrInvalidExecution
	}
	if _, err := domain.NewWorkflowDefinition(input.Revision.Snapshot().Definition.Settings(), input.Limits); err != nil {
		return ports.ConversationEvaluationResult{}, ErrInvalidExecution
	}
	started := e.deps.Clock.Now()
	calls := &atomic.Int64{}
	runtime, err := e.prepare(ctx, input, calls)
	if err != nil {
		return ports.ConversationEvaluationResult{}, err
	}
	projector, err := modelapp.NewEvaluationProjector(input.Case, runtime.runtimeIDs)
	if err != nil {
		return ports.ConversationEvaluationResult{}, err
	}
	before, err := runtime.snapshot(ctx)
	if err != nil {
		return ports.ConversationEvaluationResult{}, err
	}
	session, err := runtime.application.StartRealtimeVoiceSession(ctx, app.RealtimeVoiceSessionInput{Principal: input.Principal, TenantID: tenant.ID(input.Revision.Snapshot().TenantID), InventoryID: runtime.inventoryID, Source: app.RealtimeVoiceSourceMobile, InputAudio: ports.RealtimeAudioFormat{MimeType: "audio/mp4", Channels: 1}, OutputAudio: app.RealtimeVoiceOutputAudio{MimeTypes: []string{"audio/mpeg"}}})
	if err != nil {
		return ports.ConversationEvaluationResult{}, err
	}
	result := ports.ConversationEvaluationResult{Coverage: ports.EvaluationCoverageText}
	completed := false
	err = runtime.application.RunRealtimeVoiceQuery(ctx, app.RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{{0}}}, func(event app.RealtimeVoiceEvent) error {
		switch event.Type {
		case app.RealtimeVoiceEventAssistantResponseCompleted:
			if event.Response == nil || completed {
				return modelapp.ErrInvalidEvaluationObservation
			}
			var projectionErr error
			result.Outcome, projectionErr = projector.Response(*event.Response)
			if projectionErr != nil {
				return projectionErr
			}
			completed = true
		case app.RealtimeVoiceEventActionPlanProposed:
			if event.ActionPlan == nil || completed {
				return modelapp.ErrInvalidEvaluationObservation
			}
			plan, found, readErr := runtime.store.ActionPlanByID(ctx, session.TenantID, session.InventoryID, event.ActionPlan.PlanID)
			if readErr != nil {
				return readErr
			}
			if !found || plan.State != actionplan.StateProposed || plan.PrincipalID != input.Principal.ID || plan.RealtimeSessionID != session.ID {
				return modelapp.ErrInvalidEvaluationObservation
			}
			result.Outcome, readErr = projector.Proposal(plan.Commands)
			if readErr != nil {
				return readErr
			}
			completed = true
		case app.RealtimeVoiceEventActionPlanApproved, app.RealtimeVoiceEventActionPlanExecuted:
			return ErrFixtureMutation
		}
		return nil
	})
	if ctx.Err() != nil {
		return ports.ConversationEvaluationResult{}, ctx.Err()
	}
	if snapshotErr := runtime.assertUnchanged(ctx, before); snapshotErr != nil {
		return ports.ConversationEvaluationResult{}, snapshotErr
	}

	if err != nil {
		if errors.Is(err, ports.ErrInvalidProviderInput) {
			return ports.ConversationEvaluationResult{}, err
		}
		code := app.RealtimeVoiceSafeErrorCode(err)
		if code != "language_inference_failed" {
			return ports.ConversationEvaluationResult{}, err
		}
		result.Outcome = domain.EvaluationObservedOutcome{Kind: domain.EvaluationOutcomeFailure}
		result.SafeFailureCode = code
	} else if !completed {
		return ports.ConversationEvaluationResult{}, modelapp.ErrInvalidEvaluationObservation
	}
	result.ModelCalls = int(calls.Load())
	result.Duration = e.deps.Clock.Now().Sub(started)
	if result.Duration < 0 {
		return ports.ConversationEvaluationResult{}, ErrInvalidExecution
	}
	return result, nil
}
