package gormstore

import (
	"encoding/json"
	"time"

	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type evaluationCaseModel struct {
	TenantID         string      `gorm:"primaryKey;size:64"`
	ID               string      `gorm:"primaryKey;size:64"`
	Tenant           tenantModel `gorm:"foreignKey:TenantID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Title            string      `gorm:"not null;type:text"`
	LatestRevision   int         `gorm:"not null"`
	LatestRevisionID string      `gorm:"not null;size:64"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (evaluationCaseModel) TableName() string { return "conversation_evaluation_cases" }

type evaluationCaseRevisionModel struct {
	TenantID       string `gorm:"primaryKey;size:64;uniqueIndex:evaluation_case_revision_sequence,priority:1"`
	CaseID         string `gorm:"primaryKey;size:64;uniqueIndex:evaluation_case_revision_sequence,priority:2"`
	ID             string `gorm:"primaryKey;size:64"`
	Number         int    `gorm:"not null;uniqueIndex:evaluation_case_revision_sequence,priority:3"`
	AuthorID       string `gorm:"not null;type:text"`
	DefinitionJSON string `gorm:"not null;type:text"`
	CreatedAt      time.Time
	Case           evaluationCaseModel `gorm:"foreignKey:TenantID,CaseID;references:TenantID,ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

func (evaluationCaseRevisionModel) TableName() string {
	return "conversation_evaluation_case_revisions"
}
func evaluationCaseRevisionToModel(revision domain.EvaluationCaseRevision) (evaluationCaseRevisionModel, error) {
	snapshot := revision.Snapshot()
	if _, err := domain.NewEvaluationCaseRevision(snapshot); err != nil {
		return evaluationCaseRevisionModel{}, err
	}
	raw, err := json.Marshal(snapshot.Definition.Settings())
	if err != nil {
		return evaluationCaseRevisionModel{}, err
	}
	return evaluationCaseRevisionModel{TenantID: string(snapshot.TenantID), CaseID: string(snapshot.CaseID), ID: string(snapshot.ID), Number: snapshot.Number, AuthorID: string(snapshot.AuthorID), DefinitionJSON: string(raw), CreatedAt: snapshot.CreatedAt}, nil
}
func evaluationCaseRevisionFromModel(model evaluationCaseRevisionModel) (domain.EvaluationCaseRevision, error) {
	var input domain.EvaluationCaseDefinitionInput
	if err := json.Unmarshal([]byte(model.DefinitionJSON), &input); err != nil {
		return domain.EvaluationCaseRevision{}, err
	}
	definition, err := domain.NewEvaluationCaseDefinition(input)
	if err != nil {
		return domain.EvaluationCaseRevision{}, err
	}
	return domain.NewEvaluationCaseRevision(domain.EvaluationCaseRevisionInput{ID: domain.EvaluationCaseRevisionID(model.ID), CaseID: domain.EvaluationCaseID(model.CaseID), TenantID: domain.TenantID(model.TenantID), AuthorID: domain.EvaluationCaseAuthorID(model.AuthorID), Number: model.Number, Definition: definition, CreatedAt: model.CreatedAt})
}
func evaluationCaseHeadFromModel(model evaluationCaseModel) ports.EvaluationCaseHeadRecord {
	return ports.EvaluationCaseHeadRecord{TenantID: tenant.ID(model.TenantID), ID: domain.EvaluationCaseID(model.ID), Title: model.Title, LatestRevision: model.LatestRevision, LatestRevisionID: domain.EvaluationCaseRevisionID(model.LatestRevisionID), CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt}
}
