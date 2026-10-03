package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printjobs/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printjobs/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	printingapp "github.com/stuffstash/stuff-stash/internal/app/printing"
	p "github.com/stuffstash/stuff-stash/internal/domain/printing"
	"strings"
)

func consumerToken(authorization string) (string, error) {
	scheme, token, ok := strings.Cut(authorization, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
		return "", huma.Error401Unauthorized("Invalid connector credential")
	}
	return token, nil
}
func consumerProof(id string, proof dto.PrintClaimProof) printingapp.ClaimProof {
	return printingapp.ClaimProof{AttemptID: p.AttemptID(id), SessionID: p.SessionID(proof.SessionID), Secret: proof.ClaimToken, Revision: proof.Revision}
}
func consumerOutput(s printingapp.ConsumerService, j p.Job, id p.AttemptID, artifact bool, err error) (*dto.PrintConsumerOutput, error) {
	if err != nil {
		return nil, jobError(err)
	}
	return &dto.PrintConsumerOutput{CacheControl: "private, no-store", Body: shared.SuccessEnvelope[*dto.PrintConsumerAttempt]{Data: mapper.ConsumerAttempt(j, id, s.Now(), artifact)}}, nil
}
func registerConsumers(api huma.API, a app.App) {
	registerRecovery(api, a)
	huma.Post(api, "/print-consumer/claims", func(ctx context.Context, in *dto.PrintClaimInput) (*dto.PrintConsumerOutput, error) {
		token, err := consumerToken(in.Authorization)
		if err != nil {
			return nil, err
		}
		s := a.PrintConsumerJobs()
		proof := printingapp.ClaimProof{AttemptID: p.AttemptID(in.Body.AttemptID), SessionID: p.SessionID(in.Body.SessionID), Secret: in.Body.ClaimToken}
		j, _, err := s.Claim(ctx, token, p.PrinterID(in.Body.PrinterID), proof)
		return consumerOutput(s, j, proof.AttemptID, true, err)
	}, huma.OperationTags("printing"), shared.SecuredOperation, func(op *huma.Operation) { op.DefaultStatus = 200 })
	huma.Get(api, "/print-consumer/attempts/{attemptId}", func(ctx context.Context, in *dto.PrintAttemptInput) (*dto.PrintConsumerOutput, error) {
		token, err := consumerToken(in.Authorization)
		if err != nil {
			return nil, err
		}
		s := a.PrintConsumerJobs()
		j, err := s.Read(ctx, token, p.AttemptID(in.AttemptID))
		return consumerOutput(s, j, p.AttemptID(in.AttemptID), false, err)
	}, huma.OperationTags("printing"), shared.SecuredOperation)
	huma.Post(api, "/print-consumer/claims/{attemptId}/start", func(ctx context.Context, in *dto.PrintAttemptMutation) (*dto.PrintConsumerOutput, error) {
		token, err := consumerToken(in.Authorization)
		if err != nil {
			return nil, err
		}
		s := a.PrintConsumerJobs()
		j, err := s.Start(ctx, token, consumerProof(in.AttemptID, in.Body))
		return consumerOutput(s, j, p.AttemptID(in.AttemptID), false, err)
	}, huma.OperationTags("printing"), shared.SecuredOperation, func(op *huma.Operation) { op.DefaultStatus = 200 })
	huma.Post(api, "/print-consumer/claims/{attemptId}/renewal", func(ctx context.Context, in *dto.PrintAttemptMutation) (*dto.PrintConsumerOutput, error) {
		token, err := consumerToken(in.Authorization)
		if err != nil {
			return nil, err
		}
		s := a.PrintConsumerJobs()
		j, err := s.Renew(ctx, token, consumerProof(in.AttemptID, in.Body))
		return consumerOutput(s, j, p.AttemptID(in.AttemptID), false, err)
	}, huma.OperationTags("printing"), shared.SecuredOperation, func(op *huma.Operation) { op.DefaultStatus = 200 })
	huma.Post(api, "/print-consumer/claims/{attemptId}/outcome", func(ctx context.Context, in *dto.PrintOutcomeInput) (*dto.PrintConsumerOutput, error) {
		token, err := consumerToken(in.Authorization)
		if err != nil {
			return nil, err
		}
		s := a.PrintConsumerJobs()
		o := in.Body.Outcome
		j, err := s.Report(ctx, token, consumerProof(in.AttemptID, in.Body.PrintClaimProof), p.Outcome{Kind: p.OutcomeKind(o.Kind), Reason: p.OutcomeReason(o.Reason), CompletedCopies: o.CompletedCopies, Retryable: o.Retryable})
		return consumerOutput(s, j, p.AttemptID(in.AttemptID), false, err)
	}, huma.OperationTags("printing"), shared.SecuredOperation, func(op *huma.Operation) { op.DefaultStatus = 200 })
	huma.Get(api, "/print-consumer/claims/{attemptId}/content", func(ctx context.Context, in *dto.PrintContentInput) (*dto.PrintContentOutput, error) {
		token, err := consumerToken(in.Authorization)
		if err != nil {
			return nil, err
		}
		content, err := a.PrintConsumerJobs().Content(ctx, token, printingapp.ClaimProof{AttemptID: p.AttemptID(in.AttemptID), SessionID: p.SessionID(in.SessionID), Secret: in.ClaimToken, Revision: in.Revision})
		if err != nil {
			return nil, jobError(err)
		}
		return &dto.PrintContentOutput{CacheControl: "private, no-store", ContentType: "image/png", Body: content}, nil
	}, huma.OperationTags("printing"), shared.SecuredOperation)
}
