package gormstore

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"time"
)

type auditRecordModel struct {
	ID          string          `gorm:"primaryKey;size:26"`
	TenantID    string          `gorm:"not null;size:26;index:idx_audit_records_tenant_id"`
	Tenant      tenantModel     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID *string         `gorm:"size:26;index:idx_audit_records_inventory_id"`
	Inventory   *inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:InventoryID;references:ID"`
	PrincipalID string          `gorm:"not null;size:128;index"`
	Action      string          `gorm:"not null;size:80;index"`
	Source      string          `gorm:"not null;size:40"`
	TargetType  string          `gorm:"not null;size:80;index"`
	TargetID    string          `gorm:"not null;size:180;index"`
	OccurredAt  time.Time       `gorm:"not null;index"`
	RequestID   string          `gorm:"not null;default:'';size:128;index"`
	Metadata    string          `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (auditRecordModel) TableName() string {
	return "audit_records"
}

func (m auditRecordModel) toDomain() (audit.Record, bool) {
	id, ok := audit.NewID(m.ID)
	if !ok {
		return audit.Record{}, false
	}
	action, ok := audit.NewAction(m.Action)
	if !ok {
		return audit.Record{}, false
	}
	source, ok := audit.NewSource(m.Source)
	if !ok {
		return audit.Record{}, false
	}
	targetType, ok := audit.NewTargetType(m.TargetType)
	if !ok {
		return audit.Record{}, false
	}
	inventoryID := audit.InventoryID("")
	if m.InventoryID != nil {
		inventoryID = audit.InventoryID(*m.InventoryID)
	}
	metadata := map[string]string{}
	if err := json.Unmarshal([]byte(m.Metadata), &metadata); err != nil {
		return audit.Record{}, false
	}
	return audit.NewRecord(
		id,
		audit.TenantID(m.TenantID),
		inventoryID,
		audit.PrincipalID(m.PrincipalID),
		action,
		source,
		targetType,
		m.TargetID,
		m.OccurredAt,
		m.RequestID,
		metadata,
	)
}
