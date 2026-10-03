package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printjobs/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printjobs/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	p "github.com/stuffstash/stuff-stash/internal/domain/printing"
)

func registerRecovery(api huma.API, a app.App) {
	huma.Get(api, "/print-consumer/attempts", func(ctx context.Context, in *dto.PrintRecoveryInput) (*dto.PrintRecoveryOutput, error) {
		token, err := consumerToken(in.Authorization)
		if err != nil {
			return nil, err
		}
		s := a.PrintConsumerJobs()
		page, err := s.Unsettled(ctx, token, p.PrinterID(in.PrinterID), in.Limit, in.Cursor)
		if err != nil {
			return nil, jobError(err)
		}
		result := make([]dto.PrintConsumerAttempt, 0, len(page.Items))
		for _, j := range page.Items {
			result = append(result, *mapper.ConsumerAttempt(j, j.Attempts[len(j.Attempts)-1].ID, s.Now(), false))
		}
		return &dto.PrintRecoveryOutput{CacheControl: "private, no-store", Body: shared.SuccessEnvelope[[]dto.PrintConsumerAttempt]{Data: result, Meta: shared.PaginatedMeta("", page.Limit, page.NextCursor, page.HasMore)}}, nil
	}, huma.OperationTags("printing"), shared.SecuredOperation)
	huma.Post(api, "/print-consumer/attempts/{attemptId}/reconciliation", func(ctx context.Context, in *dto.PrintReconcileInput) (*dto.PrintConsumerOutput, error) {
		token, err := consumerToken(in.Authorization)
		if err != nil {
			return nil, err
		}
		s := a.PrintConsumerJobs()
		o := in.Body.Outcome
		j, err := s.Reconcile(ctx, token, p.AttemptID(in.AttemptID), in.Body.Revision, p.Outcome{Kind: p.OutcomeKind(o.Kind), CompletedCopies: o.CompletedCopies, Retryable: o.Retryable, Reason: p.OutcomeReason(o.Reason)})
		return consumerOutput(s, j, p.AttemptID(in.AttemptID), false, err)
	}, huma.OperationTags("printing"), shared.SecuredOperation, func(op *huma.Operation) { op.DefaultStatus = 200 })
}
