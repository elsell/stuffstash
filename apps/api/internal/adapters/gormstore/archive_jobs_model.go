package gormstore

import "time"

type archiveJobModel struct {
	ID                string      `gorm:"column:id;primaryKey"`
	TenantID          string      `gorm:"column:tenant_id;not null;uniqueIndex:idx_archive_job_request,priority:1;index:idx_archive_job_scope,priority:1"`
	SourceInventoryID string      `gorm:"column:source_inventory_id;not null;index:idx_archive_job_scope,priority:2"`
	PrincipalID       string      `gorm:"column:principal_id;not null;uniqueIndex:idx_archive_job_request,priority:2"`
	RequestKey        string      `gorm:"column:request_key;not null;uniqueIndex:idx_archive_job_request,priority:3"`
	RequestJSON       string      `gorm:"column:request_json;type:text;not null"`
	StateJSON         string      `gorm:"column:state_json;type:text;not null"`
	State             string      `gorm:"column:state;not null;index:idx_archive_job_queue,priority:1"`
	Revision          int64       `gorm:"column:revision;not null"`
	CreatedAt         time.Time   `gorm:"column:created_at;not null"`
	UpdatedAt         time.Time   `gorm:"column:updated_at;not null"`
	ExpiresAt         time.Time   `gorm:"column:expires_at;not null;index"`
	LeaseUntil        time.Time   `gorm:"column:lease_until;not null;index:idx_archive_job_queue,priority:2"`
	Tenant            tenantModel `gorm:"foreignKey:TenantID;references:ID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
}

func (archiveJobModel) TableName() string { return "archive_jobs" }
