package dto

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"time"
)

type EvaluationCaseFixtureAsset struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Kind        string   `json:"kind"`
	Description string   `json:"description,omitempty"`
	ParentID    string   `json:"parentId,omitempty"`
	TagNames    []string `json:"tagNames,omitempty"`
}
type EvaluationCaseLocation struct {
	AssetID    string `json:"assetId"`
	AncestorID string `json:"ancestorId"`
}
type EvaluationCaseProposal struct {
	Operation     string `json:"operation"`
	TargetID      string `json:"targetId,omitempty"`
	DestinationID string `json:"destinationId,omitempty"`
	NewTitle      string `json:"newTitle,omitempty"`
	NewKind       string `json:"newKind,omitempty"`
	Details       string `json:"details,omitempty"`
}
type EvaluationCaseExpectations struct {
	Kind                string                   `json:"kind"`
	ReferencedAssets    []string                 `json:"referencedAssets,omitempty"`
	Locations           []EvaluationCaseLocation `json:"locations,omitempty"`
	Proposals           []EvaluationCaseProposal `json:"proposals,omitempty"`
	ForbiddenOperations []string                 `json:"forbiddenOperations,omitempty"`
}
type EvaluationCaseDefinition struct {
	Title        string                       `json:"title"`
	Utterance    string                       `json:"utterance"`
	Assets       []EvaluationCaseFixtureAsset `json:"assets,omitempty"`
	Expectations EvaluationCaseExpectations   `json:"expectations"`
}
type AccessInput struct {
	Authorization string `header:"Authorization"`
	RequestID     string `header:"X-Request-ID"`
	TenantID      string `path:"tenantId"`
}
type CreateInput struct {
	AccessInput
	Body EvaluationCaseCreateBody
}
type EvaluationCaseCreateBody struct {
	Definition EvaluationCaseDefinition `json:"definition"`
}
type AppendInput struct {
	AccessInput
	CaseID string `path:"caseId"`
	Body   EvaluationCaseAppendBody
}
type EvaluationCaseAppendBody struct {
	ExpectedRevision int                      `json:"expectedRevision" minimum:"1"`
	Definition       EvaluationCaseDefinition `json:"definition"`
}
type GetInput struct {
	AccessInput
	CaseID string `path:"caseId"`
}
type GetRevisionInput struct {
	GetInput
	RevisionID string `path:"revisionId"`
}
type ListInput struct {
	AccessInput
	Limit  int    `query:"limit" minimum:"1" maximum:"100"`
	Cursor string `query:"cursor"`
}
type EvaluationCaseRevision struct {
	ID         string                   `json:"id"`
	CaseID     string                   `json:"caseId"`
	Number     int                      `json:"number"`
	AuthorID   string                   `json:"authorId"`
	CreatedAt  time.Time                `json:"createdAt"`
	Definition EvaluationCaseDefinition `json:"definition"`
}
type EvaluationCaseHead struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	LatestRevision   int       `json:"latestRevision"`
	LatestRevisionID string    `json:"latestRevisionId"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}
type RevisionOutput struct {
	Body shared.SuccessEnvelope[EvaluationCaseRevision]
}
type ListOutput struct {
	Body shared.SuccessEnvelope[[]EvaluationCaseHead]
}
