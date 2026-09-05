package dto

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"time"
)

type Budget struct {
	EvidenceRounds int `json:"evidenceRounds"`
	ModelCalls     int `json:"modelCalls"`
	ElapsedSeconds int `json:"elapsedSeconds"`
	FollowUpTurns  int `json:"followUpTurns"`
}
type Step struct {
	Kind              string `json:"kind"`
	ProviderProfileID string `json:"providerProfileId,omitempty"`
	Instructions      string `json:"instructions,omitempty"`
	Attempts          int    `json:"attempts"`
}
type Definition struct {
	Name      string `json:"name"`
	Retrieval string `json:"retrieval"`
	Response  string `json:"response"`
	Budget    Budget `json:"budget"`
	Steps     []Step `json:"steps"`
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
	ID         string     `json:"id"`
	WorkflowID string     `json:"workflowId"`
	Number     int        `json:"number"`
	AuthorID   string     `json:"authorId"`
	CreatedAt  time.Time  `json:"createdAt"`
	Definition Definition `json:"definition"`
}
type RevisionOutput struct {
	Body shared.SuccessEnvelope[Revision]
}
