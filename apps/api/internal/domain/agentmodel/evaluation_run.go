package agentmodel

import (
	"encoding/hex"
	"errors"
	"slices"
	"strings"
	"time"
)

var ErrInvalidEvaluationRun = errors.New("invalid evaluation run")
var ErrEvaluationRunTransition = errors.New("evaluation run transition rejected")

const MaxEvaluationRunCases = 100
const MaxEvaluationRunAttempts = 10

type EvaluationRuntimeContract string

const (
	CurrentEvaluationRuntimeContract EvaluationRuntimeContract = "conversation-tools-v1"
	LegacyEvaluationRuntimeContract  EvaluationRuntimeContract = "legacy-investigation-v1"
)

type EvaluationRunID string
type EvaluationRunState string

const (
	EvaluationRunQueued    EvaluationRunState = "queued"
	EvaluationRunRunning   EvaluationRunState = "running"
	EvaluationRunSucceeded EvaluationRunState = "succeeded"
	EvaluationRunFailed    EvaluationRunState = "failed"
	EvaluationRunCancelled EvaluationRunState = "cancelled"
)

type EvaluationRunFailureCode string

const (
	EvaluationRunFailureWorkerLost           EvaluationRunFailureCode = "worker_lost"
	EvaluationRunFailureConfigurationChanged EvaluationRunFailureCode = "configuration_changed"
	EvaluationRunFailureAccessRevoked        EvaluationRunFailureCode = "access_revoked"
	EvaluationRunFailureExecution            EvaluationRunFailureCode = "execution_failed"
)

type EvaluationRunProvider struct {
	Step            WorkflowStepKind
	ProfileID       ProviderProfileID
	ConfigurationID string
}
type EvaluationRunInput struct {
	RuntimeContract EvaluationRuntimeContract
	ID              EvaluationRunID
	TenantID        TenantID
	AuthorID        WorkflowAuthorID
	CreatedAt       time.Time
	Workflow        WorkflowRevision
	Cases           []EvaluationCaseRevision
	Limits          WorkflowLimits
	MaxAttempts     int
	Providers       []EvaluationRunProvider
}
type EvaluationRunCaseResult struct {
	CaseRevisionID EvaluationCaseRevisionID
	Observation    EvaluationObservedOutcome
	Verdict        EvaluationVerdict
	ModelCalls     int
	Duration       time.Duration
	CompletedAt    time.Time
}
type EvaluationRunSnapshot struct {
	Input       EvaluationRunInput
	State       EvaluationRunState
	Version     int
	Attempts    int
	LeaseToken  string
	LeaseUntil  time.Time
	StartedAt   time.Time
	UpdatedAt   time.Time
	FinishedAt  time.Time
	FailureCode EvaluationRunFailureCode
	Results     []EvaluationRunCaseResult
}

// All transitions return a new value. Persistence must additionally compare Version
// atomically; a lease token by itself does not protect against concurrent writes.
type EvaluationRun struct{ snapshot EvaluationRunSnapshot }

func NewEvaluationRun(input EvaluationRunInput) (EvaluationRun, error) {
	if input.RuntimeContract == "" {
		input.RuntimeContract = CurrentEvaluationRuntimeContract
	}
	if input.RuntimeContract != CurrentEvaluationRuntimeContract {
		return EvaluationRun{}, ErrInvalidEvaluationRun
	}
	return validatedEvaluationRun(input)
}

// Historical snapshots remain readable without endorsing their quality evidence.
func validatedEvaluationRun(input EvaluationRunInput) (EvaluationRun, error) {
	if input.RuntimeContract != CurrentEvaluationRuntimeContract && input.RuntimeContract != LegacyEvaluationRuntimeContract {
		return EvaluationRun{}, ErrInvalidEvaluationRun
	}
	if !workflowIdentifierValid(string(input.ID)) || !workflowAuthorValid(string(input.AuthorID)) || strings.TrimSpace(string(input.TenantID)) == "" || input.CreatedAt.IsZero() || len(input.Cases) == 0 || len(input.Cases) > MaxEvaluationRunCases || input.MaxAttempts < 1 || input.MaxAttempts > MaxEvaluationRunAttempts {
		return EvaluationRun{}, ErrInvalidEvaluationRun
	}
	workflow, err := NewWorkflowRevision(input.Workflow.Snapshot())
	if err != nil || workflow.Snapshot().TenantID != input.TenantID || workflow.Snapshot().CreatedAt.After(input.CreatedAt) {
		return EvaluationRun{}, ErrInvalidEvaluationRun
	}
	if _, err := NewWorkflowDefinition(workflow.Snapshot().Definition.Settings(), input.Limits); err != nil {
		return EvaluationRun{}, ErrInvalidEvaluationRun
	}
	seen := map[EvaluationCaseID]bool{}
	for _, revision := range input.Cases {
		value, err := NewEvaluationCaseRevision(revision.Snapshot())
		data := value.Snapshot()
		if err != nil || data.TenantID != input.TenantID || data.CreatedAt.After(input.CreatedAt) || seen[data.CaseID] {
			return EvaluationRun{}, ErrInvalidEvaluationRun
		}
		seen[data.CaseID] = true
	}
	definition := workflow.Snapshot().Definition.Settings()
	bindings := map[WorkflowStepKind]EvaluationRunProvider{}
	configurations := map[ProviderProfileID]string{}
	for _, binding := range input.Providers {
		_, duplicate := bindings[binding.Step]
		digest, err := hex.DecodeString(binding.ConfigurationID)
		if duplicate || binding.ProfileID == "" || !workflowProfileReferenceValid(string(binding.ProfileID)) || err != nil || len(digest) != 32 || strings.ToLower(binding.ConfigurationID) != binding.ConfigurationID {
			return EvaluationRun{}, ErrInvalidEvaluationRun
		}
		if previous, exists := configurations[binding.ProfileID]; exists && previous != binding.ConfigurationID {
			return EvaluationRun{}, ErrInvalidEvaluationRun
		}
		configurations[binding.ProfileID] = binding.ConfigurationID
		bindings[binding.Step] = binding
	}
	var defaultProfile ProviderProfileID
	for _, step := range definition.Steps {
		if step.Kind == WorkflowStepRespond && definition.Response == WorkflowResponseGrounded {
			continue
		}
		binding, ok := bindings[step.Kind]
		if !ok || (step.ProviderProfileID != "" && step.ProviderProfileID != string(binding.ProfileID)) {
			return EvaluationRun{}, ErrInvalidEvaluationRun
		}
		if step.ProviderProfileID == "" {
			if defaultProfile != "" && defaultProfile != binding.ProfileID {
				return EvaluationRun{}, ErrInvalidEvaluationRun
			}
			defaultProfile = binding.ProfileID
		}
		delete(bindings, step.Kind)
	}
	if len(bindings) != 0 {
		return EvaluationRun{}, ErrInvalidEvaluationRun
	}
	input.Cases = slices.Clone(input.Cases)
	input.Providers = slices.Clone(input.Providers)
	return EvaluationRun{snapshot: EvaluationRunSnapshot{Input: input, State: EvaluationRunQueued, Version: 1, UpdatedAt: input.CreatedAt}}, nil
}

func (run EvaluationRun) Snapshot() EvaluationRunSnapshot {
	value := run.snapshot
	value.Input.Cases = slices.Clone(value.Input.Cases)
	value.Input.Providers = slices.Clone(value.Input.Providers)
	value.Results = slices.Clone(value.Results)
	for i := range value.Results {
		value.Results[i].Observation = cloneEvaluationObservation(value.Results[i].Observation)
		value.Results[i].Verdict.Failures = slices.Clone(value.Results[i].Verdict.Failures)
	}
	return value
}
func cloneEvaluationObservation(value EvaluationObservedOutcome) EvaluationObservedOutcome {
	value.ReferencedAssets = slices.Clone(value.ReferencedAssets)
	value.Locations = slices.Clone(value.Locations)
	value.Proposals = slices.Clone(value.Proposals)
	value.ExecutedOperations = slices.Clone(value.ExecutedOperations)
	return value
}
