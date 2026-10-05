package ports

import (
	"context"
	"encoding/json"
)

// EvaluationOptional is retained for existing evaluation models.
type EvaluationOptional[T any] = Optional[T]

type EvaluationCaseDefinition struct {
	Assets       EvaluationOptional[[]EvaluationCaseFixtureAsset] `json:"assets,omitempty"`
	Expectations EvaluationCaseExpectations                       `json:"expectations"`
	Title        string                                           `json:"title"`
	Utterance    string                                           `json:"utterance"`
}

type EvaluationCaseExpectations struct {
	ForbiddenOperations EvaluationOptional[[]string]                 `json:"forbiddenOperations,omitempty"`
	Kind                string                                       `json:"kind"`
	Locations           EvaluationOptional[[]EvaluationCaseLocation] `json:"locations,omitempty"`
	Proposals           EvaluationOptional[[]EvaluationCaseProposal] `json:"proposals,omitempty"`
	ReferencedAssets    EvaluationOptional[[]string]                 `json:"referencedAssets,omitempty"`
}

type EvaluationCaseFixtureAsset struct {
	Description EvaluationOptional[string]   `json:"description,omitempty"`
	ID          string                       `json:"id"`
	Kind        string                       `json:"kind"`
	ParentID    EvaluationOptional[string]   `json:"parentId,omitempty"`
	TagNames    EvaluationOptional[[]string] `json:"tagNames,omitempty"`
	Title       string                       `json:"title"`
}

type EvaluationCaseHead struct {
	CreatedAt        string `json:"createdAt"`
	ID               string `json:"id"`
	LatestRevision   int64  `json:"latestRevision"`
	LatestRevisionID string `json:"latestRevisionId"`
	Title            string `json:"title"`
	UpdatedAt        string `json:"updatedAt"`
}

type EvaluationCaseLocation struct {
	AncestorID string `json:"ancestorId"`
	AssetID    string `json:"assetId"`
}

type EvaluationCaseProposal struct {
	DestinationID EvaluationOptional[string] `json:"destinationId,omitempty"`
	Details       EvaluationOptional[string] `json:"details,omitempty"`
	NewKind       EvaluationOptional[string] `json:"newKind,omitempty"`
	NewTitle      EvaluationOptional[string] `json:"newTitle,omitempty"`
	Operation     string                     `json:"operation"`
	TargetID      EvaluationOptional[string] `json:"targetId,omitempty"`
}

type EvaluationCaseRevision struct {
	AuthorID   string                   `json:"authorId"`
	CaseID     string                   `json:"caseId"`
	CreatedAt  string                   `json:"createdAt"`
	Definition EvaluationCaseDefinition `json:"definition"`
	ID         string                   `json:"id"`
	Number     int64                    `json:"number"`
}

type EvaluationRun struct {
	AuthorID       string                     `json:"authorId"`
	Cases          []EvaluationRunPinnedCase  `json:"cases"`
	CompletedCases int64                      `json:"completedCases"`
	Coverage       string                     `json:"coverage"`
	CreatedAt      string                     `json:"createdAt"`
	FailureCode    EvaluationOptional[string] `json:"failureCode,omitempty"`
	FinishedAt     *string                    `json:"finishedAt"`
	ID             string                     `json:"id"`
	PassedCases    int64                      `json:"passedCases"`
	Providers      []EvaluationRunProvider    `json:"providers"`
	Results        []EvaluationRunResult      `json:"results"`
	RevisionID     string                     `json:"revisionId"`
	StartedAt      *string                    `json:"startedAt"`
	State          string                     `json:"state"`
	TotalCases     int64                      `json:"totalCases"`
	UpdatedAt      string                     `json:"updatedAt"`
	Version        int64                      `json:"version"`
	WorkflowID     string                     `json:"workflowId"`
}

type EvaluationRunFailure struct {
	Code      string                     `json:"code"`
	FixtureID EvaluationOptional[string] `json:"fixtureId,omitempty"`
	Operation EvaluationOptional[string] `json:"operation,omitempty"`
}

type EvaluationRunHead struct {
	CompletedCases int64  `json:"completedCases"`
	CreatedAt      string `json:"createdAt"`
	ID             string `json:"id"`
	PassedCases    int64  `json:"passedCases"`
	RevisionID     string `json:"revisionId"`
	State          string `json:"state"`
	TotalCases     int64  `json:"totalCases"`
	UpdatedAt      string `json:"updatedAt"`
	Version        int64  `json:"version"`
	WorkflowID     string `json:"workflowId"`
}

type EvaluationRunObservation struct {
	ExecutedOperations []string                 `json:"executedOperations"`
	Kind               string                   `json:"kind"`
	Locations          []EvaluationCaseLocation `json:"locations"`
	Proposals          []EvaluationCaseProposal `json:"proposals"`
	ReferencedAssets   []string                 `json:"referencedAssets"`
}

type EvaluationRunPinnedCase struct {
	CaseID     string `json:"caseId"`
	RevisionID string `json:"revisionId"`
	Title      string `json:"title"`
}

type EvaluationRunProvider struct {
	ConfigurationID string `json:"configurationId"`
	ProfileID       string `json:"profileId"`
}

type EvaluationRunResult struct {
	CaseRevisionID       string                   `json:"caseRevisionId"`
	CompletedAt          string                   `json:"completedAt"`
	DurationMilliseconds json.Number              `json:"durationMilliseconds"`
	ModelCalls           int64                    `json:"modelCalls"`
	Observation          EvaluationRunObservation `json:"observation"`
	Verdict              EvaluationRunVerdict     `json:"verdict"`
}

type EvaluationRunVerdict struct {
	Failures []EvaluationRunFailure `json:"failures"`
	Passed   bool                   `json:"passed"`
}

type EvaluationCancellation struct {
	ExpectedVersion int64   `json:"expectedVersion"`
	Schema          *string `json:"$schema,omitempty"`
}

type EvaluationAPI interface {
	CreateEvaluationCase(context.Context, string, []byte) (Result[EvaluationCaseRevision], error)
	CreateEvaluationRevision(context.Context, string, string, []byte) (Result[EvaluationCaseRevision], error)
	CreateEvaluationRun(context.Context, string, []byte) (Result[EvaluationRun], error)

	CancelEvaluationRun(context.Context, string, string, EvaluationCancellation) (Result[EvaluationRun], error)
	EvaluationCases(context.Context, string, Page) (Result[[]EvaluationCaseHead], error)
	EvaluationCase(context.Context, string, string) (Result[EvaluationCaseRevision], error)
	EvaluationRevisions(context.Context, string, string, Page) (Result[[]EvaluationCaseRevision], error)
	EvaluationRevision(context.Context, string, string, string) (Result[EvaluationCaseRevision], error)
	EvaluationRuns(context.Context, string, Page) (Result[[]EvaluationRunHead], error)
	EvaluationRun(context.Context, string, string) (Result[EvaluationRun], error)
}
