package dto

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"time"
)

type Budget struct {
	ToolCalls      int `json:"toolCalls"`
	ModelCalls     int `json:"modelCalls"`
	ElapsedSeconds int `json:"elapsedSeconds"`
	FollowUpTurns  int `json:"followUpTurns"`
}
type Definition struct {
	Name              string `json:"name"`
	ProviderProfileID string `json:"providerProfileId,omitempty"`
	Instructions      string `json:"instructions,omitempty"`
	Budget            Budget `json:"budget"`
}
type CreateInput struct {
	Authorization string `header:"Authorization"`
	RequestID     string `header:"X-Request-ID"`
	TenantID      string `path:"tenantId"`
	Body          struct {
		Definition Definition `json:"definition"`
	}
}
type AppendInput struct {
	Authorization string `header:"Authorization"`
	RequestID     string `header:"X-Request-ID"`
	TenantID      string `path:"tenantId"`
	WorkflowID    string `path:"workflowId"`
	Body          struct {
		ExpectedRevision int        `json:"expectedRevision" minimum:"1"`
		Definition       Definition `json:"definition"`
	}
}
type Revision struct {
	SettingsMigration string     `json:"settingsMigration,omitempty"`
	ID                string     `json:"id"`
	WorkflowID        string     `json:"workflowId"`
	Number            int        `json:"number"`
	AuthorID          string     `json:"authorId"`
	CreatedAt         time.Time  `json:"createdAt"`
	Definition        Definition `json:"definition"`
}
type RevisionOutput struct {
	Body shared.SuccessEnvelope[Revision]
}
