package agentmodel

import "slices"

// RestoreEvaluationRun validates persistence data as rigorously as new input.
// Observed outcomes determine verdicts; stored pass flags are never authoritative.
func RestoreEvaluationRun(snapshot EvaluationRunSnapshot) (EvaluationRun, error) {
	queued, err := NewEvaluationRun(snapshot.Input)
	if err != nil {
		return EvaluationRun{}, ErrInvalidEvaluationRun
	}
	if snapshot.UpdatedAt.Before(snapshot.Input.CreatedAt) || snapshot.Attempts < 0 || snapshot.Attempts > snapshot.Input.MaxAttempts || len(snapshot.Results) > len(snapshot.Input.Cases) || snapshot.Version < 1+snapshot.Attempts+len(snapshot.Results) {
		return EvaluationRun{}, ErrInvalidEvaluationRun
	}
	if snapshot.Attempts == 0 {
		if !snapshot.StartedAt.IsZero() || len(snapshot.Results) > 0 {
			return EvaluationRun{}, ErrInvalidEvaluationRun
		}
	} else if snapshot.StartedAt.IsZero() || snapshot.StartedAt.Before(snapshot.Input.CreatedAt) || snapshot.StartedAt.After(snapshot.UpdatedAt) {
		return EvaluationRun{}, ErrInvalidEvaluationRun
	}
	allPassed := true
	previous := snapshot.StartedAt
	for index, result := range snapshot.Results {
		revision := snapshot.Input.Cases[index].Snapshot()
		if result.CaseRevisionID != revision.ID || !revision.Definition.validObservedOutcome(result.Observation) || result.ModelCalls < 0 || result.ModelCalls > snapshot.Input.Workflow.Snapshot().Definition.Settings().Budget.ModelCalls || result.Duration < 0 || result.CompletedAt.IsZero() || result.CompletedAt.Before(previous) || result.CompletedAt.After(snapshot.UpdatedAt) {
			return EvaluationRun{}, ErrInvalidEvaluationRun
		}
		verdict := revision.Definition.Evaluate(result.Observation)
		if verdict.Passed != result.Verdict.Passed || !slices.Equal(verdict.Failures, result.Verdict.Failures) {
			return EvaluationRun{}, ErrInvalidEvaluationRun
		}
		allPassed = allPassed && verdict.Passed
		previous = result.CompletedAt
	}
	if !validEvaluationRunState(snapshot, allPassed) {
		return EvaluationRun{}, ErrInvalidEvaluationRun
	}
	snapshot.Input = queued.snapshot.Input
	return EvaluationRun{snapshot: EvaluationRun{snapshot: snapshot}.Snapshot()}, nil
}

func validEvaluationRunState(value EvaluationRunSnapshot, allPassed bool) bool {
	complete := len(value.Results) == len(value.Input.Cases)
	hasLease := value.LeaseToken != "" || !value.LeaseUntil.IsZero()
	switch value.State {
	case EvaluationRunQueued:
		return value.Version == 1 && value.Attempts == 0 && len(value.Results) == 0 && !hasLease && value.FinishedAt.IsZero() && value.FailureCode == "" && value.UpdatedAt.Equal(value.Input.CreatedAt)
	case EvaluationRunRunning:
		return value.Attempts > 0 && !complete && workflowIdentifierValid(value.LeaseToken) && value.LeaseUntil.After(value.UpdatedAt) && value.FinishedAt.IsZero() && value.FailureCode == ""
	case EvaluationRunSucceeded, EvaluationRunFailed, EvaluationRunCancelled:
		if hasLease || value.FinishedAt.IsZero() || !value.FinishedAt.Equal(value.UpdatedAt) {
			return false
		}
	default:
		return false
	}
	if value.State == EvaluationRunSucceeded {
		return value.Attempts > 0 && complete && allPassed && value.FailureCode == ""
	}
	if value.State == EvaluationRunCancelled {
		return !complete && value.FailureCode == "" && value.Version >= 2+value.Attempts+len(value.Results)
	}
	if value.Attempts == 0 {
		return false
	}
	if complete {
		return !allPassed && value.FailureCode == ""
	}
	if value.Version < 2+value.Attempts+len(value.Results) {
		return false
	}
	switch value.FailureCode {
	case EvaluationRunFailureConfigurationChanged, EvaluationRunFailureAccessRevoked, EvaluationRunFailureExecution:
		return true
	case EvaluationRunFailureWorkerLost:
		return value.Attempts == value.Input.MaxAttempts
	default:
		return false
	}
}
