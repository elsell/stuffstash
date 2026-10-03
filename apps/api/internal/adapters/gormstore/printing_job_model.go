package gormstore

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"time"
)

// Indexed routing columns remain separate from the immutable rendering snapshot
// and attempt evidence. The aggregate is written only under the printer lock.
type printingJobModel struct {
	ArtifactContent      []byte
	ArtifactExpiresAt    time.Time
	ID                   string `gorm:"primaryKey;size:26"`
	TenantID             string `gorm:"not null;uniqueIndex:idx_print_job_request,priority:1;index:idx_print_job_queue,priority:1"`
	InventoryID          string `gorm:"not null;uniqueIndex:idx_print_job_request,priority:2;index:idx_print_job_queue,priority:2"`
	RequestedBy          string `gorm:"not null;uniqueIndex:idx_print_job_request,priority:3"`
	IdempotencyKey       string `gorm:"not null;uniqueIndex:idx_print_job_request,priority:4"`
	PrinterID            string `gorm:"not null;index:idx_print_job_queue,priority:3"`
	Status               string `gorm:"not null;index:idx_print_job_queue,priority:4"`
	MediaFingerprint     string `gorm:"not null"`
	RequestFingerprint   string `gorm:"not null"`
	Revision             uint64
	Snapshot             []byte `gorm:"not null"`
	CreatedAt, UpdatedAt time.Time
}

func (printingJobModel) TableName() string { return "print_jobs" }
func (m printingJobModel) domain() (printing.Job, error) {
	var j printing.Job
	err := json.Unmarshal(m.Snapshot, &j)
	return j, err
}
func printJobModel(j printing.Job, fingerprint string, content []byte) (printingJobModel, error) {
	data, err := json.Marshal(j)
	return printingJobModel{ArtifactContent: content, ArtifactExpiresAt: j.Artifact.ExpiresAt, ID: string(j.ID), TenantID: j.Scope.TenantID, InventoryID: j.Scope.InventoryID, PrinterID: string(j.PrinterID), RequestedBy: j.RequestedBy, IdempotencyKey: j.IdempotencyKey, Status: string(j.Status), MediaFingerprint: j.MediaFingerprint, RequestFingerprint: fingerprint, Revision: j.Revision, Snapshot: data, CreatedAt: j.CreatedAt, UpdatedAt: j.UpdatedAt}, err
}
