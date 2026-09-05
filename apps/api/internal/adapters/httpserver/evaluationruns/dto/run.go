package dto

import (
	casedto "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/evaluationcases/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"time"
)

type AccessInput struct {
	Authorization string `header:"Authorization"`
	RequestID     string `header:"X-Request-ID"`
	TenantID      string `path:"tenantId"`
}
type EvaluationRunCaseReference struct {
	CaseID     string `json:"caseId" minLength:"1"`
	RevisionID string `json:"revisionId" minLength:"1"`
}
type EvaluationRunQueueBody struct {
	WorkflowID string                       `json:"workflowId" minLength:"1"`
	RevisionID string                       `json:"revisionId" minLength:"1"`
	Cases      []EvaluationRunCaseReference `json:"cases" minItems:"1" maxItems:"100"`
}
type QueueInput struct {
	AccessInput
	Body EvaluationRunQueueBody
}
type GetInput struct {
	AccessInput
	RunID string `path:"runId"`
}
type EvaluationRunCancellationBody struct {
	ExpectedVersion int `json:"expectedVersion" minimum:"1"`
}
type CancelInput struct {
	GetInput
	Body EvaluationRunCancellationBody
}
type ListInput struct {
	AccessInput
	Limit  int    `query:"limit" minimum:"1" maximum:"100"`
	Cursor string `query:"cursor"`
}
type EvaluationRunHead struct {
	ID             string    `json:"id"`
	State          string    `json:"state"`
	Version        int       `json:"version"`
	WorkflowID     string    `json:"workflowId"`
	RevisionID     string    `json:"revisionId"`
	TotalCases     int       `json:"totalCases"`
	CompletedCases int       `json:"completedCases"`
	PassedCases    int       `json:"passedCases"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
type EvaluationRunPinnedCase struct {
	EvaluationRunCaseReference
	Title string `json:"title"`
}
type EvaluationRunProvider struct {
	Step            string `json:"step"`
	ProfileID       string `json:"profileId"`
	ConfigurationID string `json:"configurationId"`
}
type EvaluationRunObservation struct {
	Kind               string                           `json:"kind"`
	ReferencedAssets   []string                         `json:"referencedAssets"`
	Locations          []casedto.EvaluationCaseLocation `json:"locations"`
	Proposals          []casedto.EvaluationCaseProposal `json:"proposals"`
	ExecutedOperations []string                         `json:"executedOperations"`
}
type EvaluationRunFailure struct {
	Code      string `json:"code"`
	FixtureID string `json:"fixtureId,omitempty"`
	Operation string `json:"operation,omitempty"`
}
type EvaluationRunVerdict struct {
	Passed   bool                   `json:"passed"`
	Failures []EvaluationRunFailure `json:"failures"`
}
type EvaluationRunResult struct {
	CaseRevisionID       string                   `json:"caseRevisionId"`
	Observation          EvaluationRunObservation `json:"observation"`
	Verdict              EvaluationRunVerdict     `json:"verdict"`
	ModelCalls           int                      `json:"modelCalls"`
	DurationMilliseconds float64                  `json:"durationMilliseconds"`
	CompletedAt          time.Time                `json:"completedAt"`
}
type EvaluationRun struct {
	EvaluationRunHead
	AuthorID    string                    `json:"authorId"`
	Coverage    string                    `json:"coverage" enum:"text_only"`
	Cases       []EvaluationRunPinnedCase `json:"cases"`
	Providers   []EvaluationRunProvider   `json:"providers"`
	Results     []EvaluationRunResult     `json:"results"`
	StartedAt   *time.Time                `json:"startedAt"`
	FinishedAt  *time.Time                `json:"finishedAt"`
	FailureCode string                    `json:"failureCode,omitempty"`
}
type RunOutput struct {
	Body shared.SuccessEnvelope[EvaluationRun]
}
type ListOutput struct {
	Body shared.SuccessEnvelope[[]EvaluationRunHead]
}
