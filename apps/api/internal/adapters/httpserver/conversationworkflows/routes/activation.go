package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/conversationworkflows/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func registerActivation(api huma.API, application app.App) {
	huma.Post(api, "/tenants/{tenantId}/conversation-workflows/{workflowId}/activation", func(ctx context.Context, input *dto.ActivateInput) (*dto.RevisionOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		expected := ports.WorkflowSelectionReference{}
		if input.Body.Expected != nil {
			expected = ports.WorkflowSelectionReference{WorkflowID: model.WorkflowID(input.Body.Expected.WorkflowID), RevisionID: model.WorkflowRevisionID(input.Body.Expected.RevisionID)}
		}
		cases := make([]app.EvaluationRunCaseReference, 0, len(input.Body.Cases))
		for _, pin := range input.Body.Cases {
			cases = append(cases, app.EvaluationRunCaseReference{CaseID: model.EvaluationCaseID(pin.CaseID), RevisionID: model.EvaluationCaseRevisionID(pin.RevisionID)})
		}
		revision, err := application.ActivateConversationWorkflow(ctx, app.ActivateWorkflowInput{EvaluationRunAccess: app.EvaluationRunAccess{Principal: principal, TenantID: tenant.ID(input.TenantID), Source: audit.SourceAPI, RequestID: input.RequestID}, WorkflowID: model.WorkflowID(input.WorkflowID), RevisionID: model.WorkflowRevisionID(input.Body.RevisionID), RunID: model.EvaluationRunID(input.Body.RunID), Cases: cases, Expected: expected})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return revisionOutput(revision, input.TenantID), nil
	}, huma.OperationTags("conversation workflows"), shared.SecuredOperation)
}
