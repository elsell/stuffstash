package gormstore

import (
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

type evaluationRunModel struct {
	TenantID       string      `gorm:"primaryKey;size:64"`
	ID             string      `gorm:"primaryKey;size:64"`
	Tenant         tenantModel `gorm:"foreignKey:TenantID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	State          string      `gorm:"not null;size:32;index:evaluation_run_queue,priority:1"`
	Version        int         `gorm:"not null"`
	LeaseUntil     *time.Time  `gorm:"index:evaluation_run_queue,priority:2"`
	WorkflowID     string      `gorm:"not null;size:64"`
	RevisionID     string      `gorm:"not null;size:64"`
	TotalCases     int         `gorm:"not null"`
	CompletedCases int         `gorm:"not null"`
	PassedCases    int         `gorm:"not null"`
	InputJSON      string      `gorm:"not null;type:text"`
	ProgressJSON   string      `gorm:"not null;type:text"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (evaluationRunModel) TableName() string { return "conversation_evaluation_runs" }
func evaluationRunToModel(run model.EvaluationRun) (evaluationRunModel, error) {
	input, progress, err := encodeEvaluationRun(run)
	if err != nil {
		return evaluationRunModel{}, err
	}
	snapshot := run.Snapshot()
	workflow := snapshot.Input.Workflow.Snapshot()
	value := evaluationRunModel{TenantID: string(snapshot.Input.TenantID), ID: string(snapshot.Input.ID), State: string(snapshot.State), Version: snapshot.Version, WorkflowID: string(workflow.WorkflowID), RevisionID: string(workflow.ID), TotalCases: len(snapshot.Input.Cases), CompletedCases: len(snapshot.Results), InputJSON: input, ProgressJSON: progress, CreatedAt: snapshot.Input.CreatedAt.UTC().Truncate(time.Microsecond), UpdatedAt: snapshot.UpdatedAt.UTC().Truncate(time.Microsecond)}
	if !snapshot.LeaseUntil.IsZero() {
		lease := snapshot.LeaseUntil.UTC().Truncate(time.Microsecond)
		value.LeaseUntil = &lease
	}
	for _, result := range snapshot.Results {
		if result.Verdict.Passed {
			value.PassedCases++
		}
	}
	return value, nil
}
func evaluationRunFromModel(value evaluationRunModel) (model.EvaluationRun, error) {
	run, err := decodeEvaluationRun(value.InputJSON, value.ProgressJSON)
	if err != nil {
		return model.EvaluationRun{}, err
	}
	expected, err := evaluationRunToModel(run)
	if err != nil {
		return model.EvaluationRun{}, err
	}
	leasesMatch := (expected.LeaseUntil == nil && value.LeaseUntil == nil) || (expected.LeaseUntil != nil && value.LeaseUntil != nil && expected.LeaseUntil.Equal(*value.LeaseUntil))
	if expected.TenantID != value.TenantID || expected.ID != value.ID || expected.State != value.State || expected.Version != value.Version || !leasesMatch || expected.WorkflowID != value.WorkflowID || expected.RevisionID != value.RevisionID || expected.TotalCases != value.TotalCases || expected.CompletedCases != value.CompletedCases || expected.PassedCases != value.PassedCases || !expected.CreatedAt.Equal(value.CreatedAt) || !expected.UpdatedAt.Equal(value.UpdatedAt) {
		return model.EvaluationRun{}, model.ErrInvalidEvaluationRun
	}
	return run, nil
}
func evaluationRunHeadFromModel(value evaluationRunModel) ports.EvaluationRunHead {
	return ports.EvaluationRunHead{EvaluationRunReference: ports.EvaluationRunReference{TenantID: tenant.ID(value.TenantID), ID: model.EvaluationRunID(value.ID)}, State: model.EvaluationRunState(value.State), Version: value.Version, WorkflowID: model.WorkflowID(value.WorkflowID), RevisionID: model.WorkflowRevisionID(value.RevisionID), TotalCases: value.TotalCases, CompletedCases: value.CompletedCases, PassedCases: value.PassedCases, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}
}
