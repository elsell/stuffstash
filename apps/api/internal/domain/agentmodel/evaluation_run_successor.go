package agentmodel

import "reflect"

// IsSuccessorOf replays one transition rather than trusting independently valid
// snapshots. Full comparison deliberately includes immutable inputs and history.
func (run EvaluationRun) IsSuccessorOf(previous EvaluationRun) bool {
	before, after := previous.Snapshot(), run.Snapshot()
	if after.Version != before.Version+1 {
		return false
	}
	if _, err := RestoreEvaluationRun(before); err != nil {
		return false
	}
	if _, err := RestoreEvaluationRun(after); err != nil {
		return false
	}
	var expected EvaluationRun
	var err error
	switch {
	case after.State == EvaluationRunCancelled:
		expected, err = previous.Cancel(after.UpdatedAt)
	case after.State == EvaluationRunFailed && after.FailureCode == EvaluationRunFailureWorkerLost:
		expected, err = previous.FailExpired(after.UpdatedAt)
	case after.State == EvaluationRunFailed && after.FailureCode != "":
		expected, err = previous.Fail(before.LeaseToken, after.FailureCode, after.UpdatedAt)
	case after.State == EvaluationRunRunning && after.Attempts == before.Attempts+1:
		expected, err = previous.Claim(after.LeaseToken, after.UpdatedAt, after.LeaseUntil.Sub(after.UpdatedAt))
	case len(after.Results) == len(before.Results)+1:
		result := after.Results[len(before.Results)]
		expected, err = previous.RecordCase(before.LeaseToken, len(before.Results), result.Observation, result.ModelCalls, result.Duration, after.UpdatedAt)
	case after.State == EvaluationRunRunning:
		expected, err = previous.Renew(before.LeaseToken, after.UpdatedAt, after.LeaseUntil.Sub(after.UpdatedAt))
	default:
		return false
	}
	return err == nil && reflect.DeepEqual(canonicalEvaluationRunSnapshot(expected.Snapshot()), canonicalEvaluationRunSnapshot(after))
}

// UTC removes monotonic metadata and normalizes equivalent zone representations.
func canonicalEvaluationRunSnapshot(value EvaluationRunSnapshot) EvaluationRunSnapshot {
	value.Input.CreatedAt = value.Input.CreatedAt.UTC()
	value.Input.Workflow.snapshot.CreatedAt = value.Input.Workflow.snapshot.CreatedAt.UTC()
	for index := range value.Input.Cases {
		value.Input.Cases[index].snapshot.CreatedAt = value.Input.Cases[index].snapshot.CreatedAt.UTC()
	}
	value.StartedAt = value.StartedAt.UTC()
	value.UpdatedAt = value.UpdatedAt.UTC()
	value.FinishedAt = value.FinishedAt.UTC()
	value.LeaseUntil = value.LeaseUntil.UTC()
	for index := range value.Results {
		value.Results[index].CompletedAt = value.Results[index].CompletedAt.UTC()
	}
	return value
}
