package ports

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
)

type EvaluationCoverage string

const EvaluationCoverageText EvaluationCoverage = "text_only"

// Provider instances are execution-only dependencies; never persist this input.
// The calling application must authorize tenant configuration before dispatch.
type ConversationEvaluationInput struct {
	Case              agentmodel.EvaluationCaseDefinition
	Revision          agentmodel.WorkflowRevision
	Limits            agentmodel.WorkflowLimits
	Principal         identity.Principal
	Providers         RealtimeVoiceProviderSet
	WorkflowProviders WorkflowLanguageProviderResolver
}
type ConversationEvaluationResult struct {
	Outcome         agentmodel.EvaluationObservedOutcome
	Coverage        EvaluationCoverage
	ModelCalls      int
	Duration        time.Duration
	SafeFailureCode string
}
type ConversationEvaluationExecutor interface {
	Execute(context.Context, ConversationEvaluationInput) (ConversationEvaluationResult, error)
}
