package ports

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type AuditRepository interface {
	SaveAuditRecord(ctx context.Context, record audit.Record) error
	ListTenantAuditRecords(ctx context.Context, tenantID tenant.ID, page AuditRecordPageRequest) ([]audit.Record, error)
	ListInventoryAuditRecords(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page AuditRecordPageRequest) ([]audit.Record, error)
	ListAssetAuditRecords(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, targetID string, request AssetAuditRecordListRequest) ([]audit.Record, error)
}

type AuditRecordPageRequest struct {
	AfterOccurredAt time.Time
	AfterRecordID   audit.ID
	Limit           int
}

type AssetAuditRecordListRequest struct {
	Limit int
}
