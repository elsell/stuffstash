package agentmodel

import "time"

func (run EvaluationRun) Claim(token string, now time.Time, duration time.Duration) (EvaluationRun, error) {
	value := run.Snapshot()
	if !run.validTime(now) || !workflowIdentifierValid(token) || token == value.LeaseToken || duration <= 0 || value.Attempts >= value.Input.MaxAttempts || (value.State != EvaluationRunQueued && (value.State != EvaluationRunRunning || now.Before(value.LeaseUntil))) {
		return EvaluationRun{}, ErrEvaluationRunTransition
	}
	value.State = EvaluationRunRunning
	value.Attempts++
	value.LeaseToken = token
	value.LeaseUntil = now.Add(duration)
	if value.StartedAt.IsZero() {
		value.StartedAt = now
	}
	return changedEvaluationRun(value, now), nil
}
func (run EvaluationRun) Renew(token string, now time.Time, duration time.Duration) (EvaluationRun, error) {
	if !run.ownsLease(token, now) || duration <= 0 || !now.Add(duration).After(run.snapshot.LeaseUntil) {
		return EvaluationRun{}, ErrEvaluationRunTransition
	}
	value := run.Snapshot()
	value.LeaseUntil = now.Add(duration)
	return changedEvaluationRun(value, now), nil
}
func (run EvaluationRun) RecordCase(token string, index int, observation EvaluationObservedOutcome, calls int, duration time.Duration, now time.Time) (EvaluationRun, error) {
	if !run.ownsLease(token, now) || index != len(run.snapshot.Results) || index >= len(run.snapshot.Input.Cases) || calls < 0 || calls > run.snapshot.Input.Workflow.Snapshot().Definition.Settings().Budget.ModelCalls || duration < 0 {
		return EvaluationRun{}, ErrEvaluationRunTransition
	}
	definition := run.snapshot.Input.Cases[index].Snapshot().Definition
	if !definition.validObservedOutcome(observation) {
		return EvaluationRun{}, ErrInvalidEvaluationRun
	}
	value := run.Snapshot()
	value.Results = append(value.Results, EvaluationRunCaseResult{CaseRevisionID: value.Input.Cases[index].Snapshot().ID, Observation: cloneEvaluationObservation(observation), Verdict: definition.Evaluate(observation), ModelCalls: calls, Duration: duration, CompletedAt: now})
	if len(value.Results) == len(value.Input.Cases) {
		value.State = EvaluationRunSucceeded
		for _, result := range value.Results {
			if !result.Verdict.Passed {
				value.State = EvaluationRunFailed
			}
		}
		finishEvaluationRun(&value, now)
	}
	return changedEvaluationRun(value, now), nil
}
func (run EvaluationRun) Cancel(now time.Time) (EvaluationRun, error) {
	if !run.validTime(now) {
		return EvaluationRun{}, ErrEvaluationRunTransition
	}
	if run.snapshot.State == EvaluationRunCancelled {
		return run, nil
	}
	if run.snapshot.State != EvaluationRunQueued && run.snapshot.State != EvaluationRunRunning {
		return EvaluationRun{}, ErrEvaluationRunTransition
	}
	value := run.Snapshot()
	value.State = EvaluationRunCancelled
	finishEvaluationRun(&value, now)
	return changedEvaluationRun(value, now), nil
}
func (run EvaluationRun) Fail(token string, code EvaluationRunFailureCode, now time.Time) (EvaluationRun, error) {
	if !run.ownsLease(token, now) || (code != EvaluationRunFailureConfigurationChanged && code != EvaluationRunFailureAccessRevoked && code != EvaluationRunFailureExecution) {
		return EvaluationRun{}, ErrEvaluationRunTransition
	}
	value := run.Snapshot()
	value.State = EvaluationRunFailed
	value.FailureCode = code
	finishEvaluationRun(&value, now)
	return changedEvaluationRun(value, now), nil
}
func (run EvaluationRun) FailExpired(now time.Time) (EvaluationRun, error) {
	if !run.validTime(now) || run.snapshot.State != EvaluationRunRunning || now.Before(run.snapshot.LeaseUntil) || run.snapshot.Attempts < run.snapshot.Input.MaxAttempts {
		return EvaluationRun{}, ErrEvaluationRunTransition
	}
	value := run.Snapshot()
	value.State = EvaluationRunFailed
	value.FailureCode = EvaluationRunFailureWorkerLost
	finishEvaluationRun(&value, now)
	return changedEvaluationRun(value, now), nil
}
func (run EvaluationRun) validTime(now time.Time) bool {
	return run.snapshot.Version > 0 && !now.IsZero() && !now.Before(run.snapshot.UpdatedAt)
}
func (run EvaluationRun) ownsLease(token string, now time.Time) bool {
	return run.validTime(now) && run.snapshot.State == EvaluationRunRunning && token != "" && token == run.snapshot.LeaseToken && now.Before(run.snapshot.LeaseUntil)
}
func changedEvaluationRun(value EvaluationRunSnapshot, now time.Time) EvaluationRun {
	value.Version++
	value.UpdatedAt = now
	return EvaluationRun{snapshot: value}
}
func finishEvaluationRun(value *EvaluationRunSnapshot, now time.Time) {
	value.LeaseToken = ""
	value.LeaseUntil = time.Time{}
	value.FinishedAt = now
}
