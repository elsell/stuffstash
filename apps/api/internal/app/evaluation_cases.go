package app

import (
	"context"
	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
)

type EvaluationCaseAccess = agentmodelapp.EvaluationCaseAccess
type SaveEvaluationCaseInput = agentmodelapp.SaveEvaluationCaseInput
type GetEvaluationCaseInput = agentmodelapp.GetEvaluationCaseInput
type ListEvaluationCasesInput = agentmodelapp.ListEvaluationCasesInput
type ListEvaluationCasesResult = agentmodelapp.ListEvaluationCasesResult

func (a App) SaveEvaluationCaseRevision(ctx context.Context, input SaveEvaluationCaseInput) (agentmodel.EvaluationCaseRevision, error) {
	return a.evaluationCaseService.SaveRevision(ctx, input)
}
func (a App) GetEvaluationCase(ctx context.Context, input GetEvaluationCaseInput) (agentmodel.EvaluationCaseRevision, error) {
	return a.evaluationCaseService.Get(ctx, input)
}
func (a App) ListEvaluationCases(ctx context.Context, input ListEvaluationCasesInput) (ListEvaluationCasesResult, error) {
	return a.evaluationCaseService.List(ctx, input)
}
