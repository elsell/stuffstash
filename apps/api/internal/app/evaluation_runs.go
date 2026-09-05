package app

import (
	"context"

	modelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
)

type QueueEvaluationRunInput = modelapp.QueueEvaluationRunInput
type CancelEvaluationRunInput = modelapp.CancelEvaluationRunInput
type EvaluationRunAccess = modelapp.EvaluationRunAccess
type EvaluationRunCaseReference = modelapp.EvaluationRunCaseReference

func (a App) QueueEvaluationRun(ctx context.Context, input QueueEvaluationRunInput) (agentmodel.EvaluationRun, error) {
	return a.evaluationRunCommands.Queue(ctx, input)
}
func (a App) CancelEvaluationRun(ctx context.Context, input CancelEvaluationRunInput) (agentmodel.EvaluationRun, error) {
	return a.evaluationRunCommands.Cancel(ctx, input)
}
func (a App) DrainEvaluationRuns(ctx context.Context, limit, concurrency int) error {
	return a.evaluationWorker.Drain(ctx, limit, concurrency)
}

type GetEvaluationRunInput = modelapp.GetEvaluationRunInput
type ListEvaluationRunsInput = modelapp.ListEvaluationRunsInput
type ListEvaluationRunsResult = modelapp.ListEvaluationRunsResult

func (a App) GetEvaluationRun(ctx context.Context, input GetEvaluationRunInput) (agentmodel.EvaluationRun, error) {
	return a.evaluationRunQueries.Get(ctx, input)
}
func (a App) ListEvaluationRuns(ctx context.Context, input ListEvaluationRunsInput) (ListEvaluationRunsResult, error) {
	return a.evaluationRunQueries.List(ctx, input)
}
