package gormstore

import (
	"encoding/json"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
)

type conversationWorkflowModel struct {
	Name             string      `gorm:"not null;type:text"`
	LatestRevisionID string      `gorm:"not null;size:64"`
	TenantID         string      `gorm:"primaryKey;size:64"`
	ID               string      `gorm:"primaryKey;size:64"`
	Tenant           tenantModel `gorm:"foreignKey:TenantID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	LatestRevision   int         `gorm:"not null"`
	ActiveRevisionID string      `gorm:"not null;default:'';size:64"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (conversationWorkflowModel) TableName() string { return "conversation_workflows" }

type conversationWorkflowRevisionModel struct {
	TenantID     string `gorm:"primaryKey;size:64;uniqueIndex:workflow_revision_sequence,priority:1"`
	WorkflowID   string `gorm:"primaryKey;size:64;uniqueIndex:workflow_revision_sequence,priority:2"`
	ID           string `gorm:"primaryKey;size:64"`
	Number       int    `gorm:"not null;uniqueIndex:workflow_revision_sequence,priority:3"`
	AuthorID     string `gorm:"not null;size:64"`
	SnapshotJSON string `gorm:"not null;type:text"`
	CreatedAt    time.Time
	Workflow     conversationWorkflowModel `gorm:"foreignKey:TenantID,WorkflowID;references:TenantID,ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

func (conversationWorkflowRevisionModel) TableName() string { return "conversation_workflow_revisions" }

type workflowDefinitionSnapshot struct {
	SettingsMigration agentmodel.WorkflowSettingsMigration
	Definition        agentmodel.WorkflowDefinitionInput
	Limits            agentmodel.WorkflowLimits
}

func workflowRevisionModel(revision agentmodel.WorkflowRevision) (conversationWorkflowRevisionModel, error) {
	input := revision.Snapshot()
	if _, err := agentmodel.NewWorkflowRevision(input); err != nil {
		return conversationWorkflowRevisionModel{}, err
	}
	raw, err := json.Marshal(workflowDefinitionSnapshot{Definition: input.Definition.Settings(), Limits: input.Limits, SettingsMigration: input.SettingsMigration})
	if err != nil {
		return conversationWorkflowRevisionModel{}, err
	}
	return conversationWorkflowRevisionModel{TenantID: string(input.TenantID), WorkflowID: string(input.WorkflowID), ID: string(input.ID), Number: input.Number, AuthorID: string(input.AuthorID), SnapshotJSON: string(raw), CreatedAt: input.CreatedAt}, nil
}

func workflowRevisionFromModel(model conversationWorkflowRevisionModel) (agentmodel.WorkflowRevision, error) {
	var snapshot workflowDefinitionSnapshot
	if err := json.Unmarshal([]byte(model.SnapshotJSON), &snapshot); err != nil {
		return agentmodel.WorkflowRevision{}, err
	}
	definition, err := agentmodel.NewWorkflowDefinition(snapshot.Definition, snapshot.Limits)
	if err != nil {
		return agentmodel.WorkflowRevision{}, err
	}
	return agentmodel.NewWorkflowRevision(agentmodel.WorkflowRevisionInput{ID: agentmodel.WorkflowRevisionID(model.ID), WorkflowID: agentmodel.WorkflowID(model.WorkflowID), TenantID: agentmodel.TenantID(model.TenantID), AuthorID: agentmodel.WorkflowAuthorID(model.AuthorID), Number: model.Number, Definition: definition, Limits: snapshot.Limits, CreatedAt: model.CreatedAt, SettingsMigration: snapshot.SettingsMigration})
}

type conversationWorkflowSelectionModel struct {
	TenantID    string                            `gorm:"primaryKey;size:64"`
	WorkflowID  string                            `gorm:"not null;size:64"`
	RevisionID  string                            `gorm:"not null;size:64"`
	ActivatedAt time.Time                         `gorm:"not null"`
	Revision    conversationWorkflowRevisionModel `gorm:"foreignKey:TenantID,WorkflowID,RevisionID;references:TenantID,WorkflowID,ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

func (conversationWorkflowSelectionModel) TableName() string {
	return "conversation_workflow_selections"
}
