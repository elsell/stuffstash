package mapper

import (
	casedto "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/evaluationcases/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/evaluationruns/dto"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

func HeadToResponse(value ports.EvaluationRunHead) dto.EvaluationRunHead {
	return dto.EvaluationRunHead{ID: string(value.ID), State: string(value.State), Version: value.Version, WorkflowID: string(value.WorkflowID), RevisionID: string(value.RevisionID), TotalCases: value.TotalCases, CompletedCases: value.CompletedCases, PassedCases: value.PassedCases, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}
}
func RunToResponse(run model.EvaluationRun) dto.EvaluationRun {
	value := run.Snapshot()
	workflow := value.Input.Workflow.Snapshot()
	result := dto.EvaluationRun{EvaluationRunHead: dto.EvaluationRunHead{ID: string(value.Input.ID), State: string(value.State), Version: value.Version, WorkflowID: string(workflow.WorkflowID), RevisionID: string(workflow.ID), TotalCases: len(value.Input.Cases), CompletedCases: len(value.Results), CreatedAt: value.Input.CreatedAt, UpdatedAt: value.UpdatedAt}, AuthorID: string(value.Input.AuthorID), Coverage: "text_only", Cases: []dto.EvaluationRunPinnedCase{}, Providers: []dto.EvaluationRunProvider{}, Results: []dto.EvaluationRunResult{}, StartedAt: optionalTime(value.StartedAt), FinishedAt: optionalTime(value.FinishedAt), FailureCode: string(value.FailureCode)}
	for _, revision := range value.Input.Cases {
		pinned := revision.Snapshot()
		result.Cases = append(result.Cases, dto.EvaluationRunPinnedCase{EvaluationRunCaseReference: dto.EvaluationRunCaseReference{CaseID: string(pinned.CaseID), RevisionID: string(pinned.ID)}, Title: pinned.Definition.Settings().Title})
	}
	for _, provider := range value.Input.Providers {
		result.Providers = append(result.Providers, dto.EvaluationRunProvider{ProfileID: string(provider.ProfileID), ConfigurationID: provider.ConfigurationID})
	}
	for _, observed := range value.Results {
		if observed.Verdict.Passed {
			result.PassedCases++
		}
		result.Results = append(result.Results, caseResult(observed))
	}
	return result
}
func optionalTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}
func caseResult(value model.EvaluationRunCaseResult) dto.EvaluationRunResult {
	observation := dto.EvaluationRunObservation{Kind: string(value.Observation.Kind), ReferencedAssets: append([]string{}, value.Observation.ReferencedAssets...), Locations: []casedto.EvaluationCaseLocation{}, Proposals: []casedto.EvaluationCaseProposal{}, ExecutedOperations: []string{}}
	for _, location := range value.Observation.Locations {
		observation.Locations = append(observation.Locations, casedto.EvaluationCaseLocation{AssetID: location.AssetID, AncestorID: location.AncestorID})
	}
	for _, proposal := range value.Observation.Proposals {
		observation.Proposals = append(observation.Proposals, casedto.EvaluationCaseProposal{Operation: string(proposal.Operation), TargetID: proposal.TargetID, DestinationID: proposal.DestinationID, NewTitle: proposal.NewTitle, NewKind: string(proposal.NewKind), Details: proposal.Details})
	}
	for _, operation := range value.Observation.ExecutedOperations {
		observation.ExecutedOperations = append(observation.ExecutedOperations, string(operation))
	}
	verdict := dto.EvaluationRunVerdict{Passed: value.Verdict.Passed, Failures: []dto.EvaluationRunFailure{}}
	for _, failure := range value.Verdict.Failures {
		verdict.Failures = append(verdict.Failures, dto.EvaluationRunFailure{Code: string(failure.Code), FixtureID: failure.FixtureID, Operation: string(failure.Operation)})
	}
	return dto.EvaluationRunResult{CaseRevisionID: string(value.CaseRevisionID), Observation: observation, Verdict: verdict, ModelCalls: value.ModelCalls, DurationMilliseconds: float64(value.Duration) / float64(time.Millisecond), CompletedAt: value.CompletedAt}
}
