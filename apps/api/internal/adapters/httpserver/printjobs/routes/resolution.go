package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printjobs/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	p "github.com/stuffstash/stuff-stash/internal/domain/printing"
)

func registerResolution(api huma.API, a app.App) {
	huma.Post(api, path+"/print-jobs/{jobId}/resolution", func(ctx context.Context, in *dto.ResolveInput) (*dto.Output, error) {
		scope, err := authenticate(ctx, a, in.Scope)
		if err != nil {
			return nil, err
		}
		j, err := a.PrintJobs().Resolve(ctx, scope, p.JobID(in.JobID), in.Body.Revision, p.ReportedOutcome(in.Body.ReportedOutcome), in.Body.AcknowledgeUncertainty)
		return output(j, 200, err)
	}, huma.OperationTags("printing"), shared.SecuredOperation)
	huma.Post(api, "/print-consumer/attempts/{attemptId}/idle-confirmation", func(ctx context.Context, in *dto.PrintIdleInput) (*dto.PrintConsumerOutput, error) {
		token, err := consumerToken(in.Authorization)
		if err != nil {
			return nil, err
		}
		s := a.PrintConsumerJobs()
		j, err := s.ConfirmIdle(ctx, token, p.AttemptID(in.AttemptID), in.Body.Revision)
		return consumerOutput(s, j, p.AttemptID(in.AttemptID), false, err)
	}, huma.OperationTags("printing"), shared.SecuredOperation)
}
