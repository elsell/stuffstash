package dto

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"time"
)

type WorkflowReadAccess struct {
	Authorization string `header:"Authorization"`
	RequestID     string `header:"X-Request-ID"`
	TenantID      string `path:"tenantId"`
}
type GetInput struct {
	WorkflowReadAccess
	WorkflowID string `path:"workflowId"`
}
type GetRevisionInput struct {
	GetInput
	RevisionID string `path:"revisionId"`
}
type ListInput struct {
	WorkflowReadAccess
	Limit  int    `query:"limit" minimum:"1" maximum:"100"`
	Cursor string `query:"cursor"`
}
type HistoryInput struct {
	ListInput
	WorkflowID string `path:"workflowId"`
}
type WorkflowHead struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	LatestRevisionID string    `json:"latestRevisionId"`
	LatestRevision   int       `json:"latestRevision"`
	ActiveRevisionID string    `json:"activeRevisionId,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}
type ListOutput struct {
	Body shared.SuccessEnvelope[[]WorkflowHead]
}
type HistoryOutput struct {
	Body shared.SuccessEnvelope[[]Revision]
}
type SelectionOutput struct {
	Body shared.SuccessEnvelope[*WorkflowSelection]
}
