package gormstore

import "time"

type actionPlanModel struct {
	ID                         string         `gorm:"primaryKey;size:64"`
	TenantID                   string         `gorm:"not null;size:26;index:idx_action_plans_tenant_created,priority:1;index:idx_action_plans_inventory_created,priority:1;index:idx_action_plans_principal_created,priority:1"`
	Tenant                     tenantModel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID                string         `gorm:"not null;size:26;index:idx_action_plans_inventory_created,priority:2"`
	Inventory                  inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:InventoryID;references:ID"`
	PrincipalID                string         `gorm:"not null;size:128;index:idx_action_plans_principal_created,priority:2"`
	Source                     string         `gorm:"not null;size:64"`
	RealtimeSessionID          string         `gorm:"not null;size:64;default:''"`
	State                      string         `gorm:"not null;size:32;index"`
	IntentSummary              string         `gorm:"not null;size:500;default:''"`
	ModelInterpretationSummary string         `gorm:"not null;size:500;default:''"`
	ConfirmationSummary        string         `gorm:"not null;size:500"`
	CommandsJSON               []byte         `gorm:"column:commands;not null"`
	RisksJSON                  []byte         `gorm:"column:risks;not null"`
	ApprovedAt                 *time.Time
	CancelledAt                *time.Time
	ExecutedAt                 *time.Time
	FailedAt                   *time.Time
	CreatedAt                  time.Time `gorm:"not null;index:idx_action_plans_tenant_created,priority:2;index:idx_action_plans_inventory_created,priority:3;index:idx_action_plans_principal_created,priority:3"`
	UpdatedAt                  time.Time `gorm:"not null"`
}

func (actionPlanModel) TableName() string {
	return "action_plans"
}
