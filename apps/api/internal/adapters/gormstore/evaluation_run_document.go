package gormstore

import (
	"encoding/json"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"time"
)

type evaluationRunWorkflowDocument struct {
	ID         model.WorkflowRevisionID
	WorkflowID model.WorkflowID
	TenantID   model.TenantID
	AuthorID   model.WorkflowAuthorID
	Number     int
	CreatedAt  time.Time
	Definition model.WorkflowDefinitionInput
	Limits     model.WorkflowLimits
}
type evaluationRunCaseDocument struct {
	ID         model.EvaluationCaseRevisionID
	CaseID     model.EvaluationCaseID
	TenantID   model.TenantID
	AuthorID   model.EvaluationCaseAuthorID
	Number     int
	CreatedAt  time.Time
	Definition model.EvaluationCaseDefinitionInput
}
type evaluationRunInputDocument struct {
	ID          model.EvaluationRunID
	TenantID    model.TenantID
	AuthorID    model.WorkflowAuthorID
	CreatedAt   time.Time
	Workflow    evaluationRunWorkflowDocument
	Cases       []evaluationRunCaseDocument
	Providers   []model.EvaluationRunProvider
	Limits      model.WorkflowLimits
	MaxAttempts int
}
type evaluationRunProgressDocument struct {
	State       model.EvaluationRunState
	Version     int
	Attempts    int
	LeaseToken  string
	LeaseUntil  time.Time
	StartedAt   time.Time
	UpdatedAt   time.Time
	FinishedAt  time.Time
	FailureCode model.EvaluationRunFailureCode
	Results     []model.EvaluationRunCaseResult
}

func encodeEvaluationRun(run model.EvaluationRun) (string, string, error) {
	snapshot := run.Snapshot()
	if _, err := model.RestoreEvaluationRun(snapshot); err != nil {
		return "", "", err
	}
	input := snapshot.Input
	workflow := input.Workflow.Snapshot()
	document := evaluationRunInputDocument{ID: input.ID, TenantID: input.TenantID, AuthorID: input.AuthorID, CreatedAt: input.CreatedAt, Providers: input.Providers, Limits: input.Limits, MaxAttempts: input.MaxAttempts, Workflow: evaluationRunWorkflowDocument{ID: workflow.ID, WorkflowID: workflow.WorkflowID, TenantID: workflow.TenantID, AuthorID: workflow.AuthorID, Number: workflow.Number, CreatedAt: workflow.CreatedAt, Definition: workflow.Definition.Settings(), Limits: workflow.Limits}}
	for _, revision := range input.Cases {
		v := revision.Snapshot()
		document.Cases = append(document.Cases, evaluationRunCaseDocument{ID: v.ID, CaseID: v.CaseID, TenantID: v.TenantID, AuthorID: v.AuthorID, Number: v.Number, CreatedAt: v.CreatedAt, Definition: v.Definition.Settings()})
	}
	encodedInput, err := json.Marshal(document)
	if err != nil {
		return "", "", err
	}
	progress := evaluationRunProgressDocument{State: snapshot.State, Version: snapshot.Version, Attempts: snapshot.Attempts, LeaseToken: snapshot.LeaseToken, LeaseUntil: snapshot.LeaseUntil, StartedAt: snapshot.StartedAt, UpdatedAt: snapshot.UpdatedAt, FinishedAt: snapshot.FinishedAt, FailureCode: snapshot.FailureCode, Results: snapshot.Results}
	encodedProgress, err := json.Marshal(progress)
	return string(encodedInput), string(encodedProgress), err
}
func decodeEvaluationRun(inputJSON, progressJSON string) (model.EvaluationRun, error) {
	var document evaluationRunInputDocument
	var progress evaluationRunProgressDocument
	if err := json.Unmarshal([]byte(inputJSON), &document); err != nil {
		return model.EvaluationRun{}, err
	}
	if err := json.Unmarshal([]byte(progressJSON), &progress); err != nil {
		return model.EvaluationRun{}, err
	}
	value := document.Workflow
	definition, err := model.NewWorkflowDefinition(value.Definition, value.Limits)
	if err != nil {
		return model.EvaluationRun{}, err
	}
	workflow, err := model.NewWorkflowRevision(model.WorkflowRevisionInput{ID: value.ID, WorkflowID: value.WorkflowID, TenantID: value.TenantID, AuthorID: value.AuthorID, Number: value.Number, CreatedAt: value.CreatedAt, Definition: definition, Limits: value.Limits})
	if err != nil {
		return model.EvaluationRun{}, err
	}
	input := model.EvaluationRunInput{ID: document.ID, TenantID: document.TenantID, AuthorID: document.AuthorID, CreatedAt: document.CreatedAt, Workflow: workflow, Providers: document.Providers, Limits: document.Limits, MaxAttempts: document.MaxAttempts}
	for _, value := range document.Cases {
		definition, err := model.NewEvaluationCaseDefinition(value.Definition)
		if err != nil {
			return model.EvaluationRun{}, err
		}
		revision, err := model.NewEvaluationCaseRevision(model.EvaluationCaseRevisionInput{ID: value.ID, CaseID: value.CaseID, TenantID: value.TenantID, AuthorID: value.AuthorID, Number: value.Number, CreatedAt: value.CreatedAt, Definition: definition})
		if err != nil {
			return model.EvaluationRun{}, err
		}
		input.Cases = append(input.Cases, revision)
	}
	return model.RestoreEvaluationRun(model.EvaluationRunSnapshot{Input: input, State: progress.State, Version: progress.Version, Attempts: progress.Attempts, LeaseToken: progress.LeaseToken, LeaseUntil: progress.LeaseUntil, StartedAt: progress.StartedAt, UpdatedAt: progress.UpdatedAt, FinishedAt: progress.FinishedAt, FailureCode: progress.FailureCode, Results: progress.Results})
}
