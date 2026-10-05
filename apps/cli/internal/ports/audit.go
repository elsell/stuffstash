package ports

import (
	"context"
	"time"
)

type AuditLevel string

const (
	HouseholdAudit AuditLevel = "household"
	InventoryAudit AuditLevel = "inventory"
	AssetAudit     AuditLevel = "asset"
)

type AuditQuery struct {
	Scope   Scope
	Level   AuditLevel
	AssetID string
	Page    Page
}
type AuditPrincipal struct {
	ID    string  `json:"id"`
	Email *string `json:"email,omitempty"`
}
type AuditRecord struct {
	ID          string            `json:"id"`
	TenantID    string            `json:"tenantId"`
	InventoryID *string           `json:"inventoryId,omitempty"`
	PrincipalID string            `json:"principalId"`
	Principal   *AuditPrincipal   `json:"principal,omitempty"`
	RequestID   *string           `json:"requestId,omitempty"`
	Action      string            `json:"action"`
	Source      string            `json:"source"`
	TargetType  string            `json:"targetType"`
	TargetID    string            `json:"targetId"`
	OccurredAt  time.Time         `json:"occurredAt"`
	Metadata    map[string]string `json:"metadata"`
}
type AuditAPI interface {
	AuditRecords(context.Context, AuditQuery) (Result[[]AuditRecord], error)
}
