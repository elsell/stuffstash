package agentmodel

import (
	"errors"
	"reflect"
)

var ErrWorkflowActivationEvidence = errors.New("workflow activation requires matching successful evaluation")

type EvaluationCasePin struct {
	CaseID     EvaluationCaseID
	RevisionID EvaluationCaseRevisionID
}
type WorkflowActivationCandidate struct {
	Workflow  WorkflowRevision
	Limits    WorkflowLimits
	Cases     []EvaluationCasePin
	Providers []EvaluationRunProvider
}

// ValidateActivation compares immutable evidence, never a model's assertion that
// a configuration is safe. Text evidence retains its text-only coverage.
func (run EvaluationRun) ValidateActivation(candidate WorkflowActivationCandidate) error {
	value := run.snapshot
	evidenceWorkflow, candidateWorkflow := value.Input.Workflow.Snapshot(), candidate.Workflow.Snapshot()
	evidenceWorkflow.CreatedAt = evidenceWorkflow.CreatedAt.UTC()
	candidateWorkflow.CreatedAt = candidateWorkflow.CreatedAt.UTC()
	if value.Input.RuntimeContract != CurrentEvaluationRuntimeContract || value.State != EvaluationRunSucceeded || !reflect.DeepEqual(evidenceWorkflow, candidateWorkflow) {
		return ErrWorkflowActivationEvidence
	}
	if _, err := NewWorkflowDefinition(candidate.Workflow.Snapshot().Definition.Settings(), candidate.Limits); err != nil {
		return ErrWorkflowActivationEvidence
	}
	if len(candidate.Cases) != len(value.Input.Cases) || len(candidate.Providers) != len(value.Input.Providers) {
		return ErrWorkflowActivationEvidence
	}
	cases := make(map[EvaluationCasePin]bool, len(value.Input.Cases))
	for _, revision := range value.Input.Cases {
		pinned := revision.Snapshot()
		cases[EvaluationCasePin{CaseID: pinned.CaseID, RevisionID: pinned.ID}] = true
	}
	for _, pin := range candidate.Cases {
		if !cases[pin] {
			return ErrWorkflowActivationEvidence
		}
		delete(cases, pin)
	}
	providers := make(map[EvaluationRunProvider]bool, len(value.Input.Providers))
	for _, provider := range value.Input.Providers {
		providers[provider] = true
	}
	for _, provider := range candidate.Providers {
		if !providers[provider] {
			return ErrWorkflowActivationEvidence
		}
		delete(providers, provider)
	}
	return nil
}
