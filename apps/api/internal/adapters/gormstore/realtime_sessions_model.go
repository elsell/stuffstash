package gormstore

import "time"

type realtimeSessionModel struct {
	ID                         string         `gorm:"primaryKey;size:26"`
	TenantID                   string         `gorm:"not null;size:26;index:idx_realtime_sessions_tenant_started,priority:1;index:idx_realtime_sessions_inventory_started,priority:1;index:idx_realtime_sessions_principal_started,priority:1"`
	Tenant                     tenantModel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID                string         `gorm:"not null;size:26;index:idx_realtime_sessions_inventory_started,priority:2"`
	Inventory                  inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:InventoryID;references:ID"`
	PrincipalID                string         `gorm:"not null;size:128;index:idx_realtime_sessions_principal_started,priority:2"`
	Source                     string         `gorm:"not null;size:64"`
	State                      string         `gorm:"not null;size:32;index"`
	SpeechToTextProfileID      string         `gorm:"not null;size:64"`
	LanguageInferenceProfileID string         `gorm:"not null;size:64"`
	TextToSpeechProfileID      string         `gorm:"not null;size:64"`
	SafeFailureCode            string         `gorm:"not null;size:64;default:''"`
	StartedAt                  time.Time      `gorm:"not null;index:idx_realtime_sessions_tenant_started,priority:2;index:idx_realtime_sessions_inventory_started,priority:3;index:idx_realtime_sessions_principal_started,priority:3"`
	LastActivityAt             time.Time      `gorm:"not null"`
	EndedAt                    *time.Time
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}

func (realtimeSessionModel) TableName() string {
	return "realtime_sessions"
}
