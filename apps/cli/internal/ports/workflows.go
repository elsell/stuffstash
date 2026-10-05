package ports

import (
	"context"
	"time"
)

type WorkflowBudget struct {
	ElapsedSeconds int64 `json:"elapsedSeconds"`
	FollowUpTurns  int64 `json:"followUpTurns"`
	ModelCalls     int64 `json:"modelCalls"`
	ToolCalls      int64 `json:"toolCalls"`
}

type WorkflowDefinition struct {
	Budget            WorkflowBudget `json:"budget"`
	Instructions      *string        `json:"instructions,omitempty"`
	Name              string         `json:"name"`
	ProviderProfileId *string        `json:"providerProfileId,omitempty"`
}

type WorkflowRevision struct {
	AuthorId          string             `json:"authorId"`
	CreatedAt         time.Time          `json:"createdAt"`
	Definition        WorkflowDefinition `json:"definition"`
	Id                string             `json:"id"`
	Number            int64              `json:"number"`
	SettingsMigration *string            `json:"settingsMigration,omitempty"`
	WorkflowId        string             `json:"workflowId"`
}

type WorkflowHead struct {
	ActiveRevisionId *string   `json:"activeRevisionId,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	Id               string    `json:"id"`
	LatestRevision   int64     `json:"latestRevision"`
	LatestRevisionId string    `json:"latestRevisionId"`
	Name             string    `json:"name"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type WorkflowSelection struct {
	RevisionId string `json:"revisionId"`
	WorkflowId string `json:"workflowId"`
}

type WorkflowsAPI interface {
	Workflows(context.Context, string, Page) (Result[[]WorkflowHead], error)
	Workflow(context.Context, string, string) (Result[WorkflowRevision], error)
	WorkflowRevisions(context.Context, string, string, Page) (Result[[]WorkflowRevision], error)
	WorkflowRevision(context.Context, string, string, string) (Result[WorkflowRevision], error)
	WorkflowSelection(context.Context, string) (Result[*WorkflowSelection], error)
}
