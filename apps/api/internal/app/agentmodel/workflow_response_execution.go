package agentmodel

import (
	"context"

	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (e *WorkflowModelExecution) GenerateResponse(ctx context.Context, input ports.VoiceResponseGenerationInput) (ports.VoiceResponseGenerationResult, error) {
	if err := ctx.Err(); err != nil {
		return ports.VoiceResponseGenerationResult{}, err
	}
	if input.Brief.Validate() != nil {
		return ports.VoiceResponseGenerationResult{}, ports.ErrInvalidProviderInput
	}
	if e.definition.Settings().Response == domain.WorkflowResponseGrounded {
		spoken, err := RenderGroundedVoiceResponse(input.Brief, 500)
		if err != nil {
			return ports.VoiceResponseGenerationResult{}, err
		}
		displayed, err := RenderGroundedVoiceResponse(input.Brief, 1000)
		return ports.VoiceResponseGenerationResult{SpokenResponse: spoken, DisplayResponse: displayed}, err
	}
	step := e.steps[domain.WorkflowStepRespond]
	input.WorkflowInstructions = step.Instructions
	input.PromptTemplate = e.providers[domain.WorkflowStepRespond].PromptTemplate
	return executeWorkflowModelStep(ctx, e, domain.WorkflowStepRespond, func(callCtx context.Context) (ports.VoiceResponseGenerationResult, error) {
		return e.providers[domain.WorkflowStepRespond].Response.GenerateResponse(callCtx, input)
	})
}
