package observability

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type observedAudit struct {
	delegate  ports.AuditRepository
	telemetry ports.Telemetry
}

func ObserveAudit(delegate ports.AuditRepository, telemetry ports.Telemetry) ports.AuditRepository {
	if delegate == nil {
		return nil
	}
	if telemetry == nil {
		telemetry = ports.NoopTelemetry{}
	}
	return observedAudit{delegate, telemetry}
}

func (a observedAudit) SaveAuditRecord(ctx context.Context, record audit.Record) (err error) {
	ctx, finish := a.telemetry.Start(ctx, ports.OperationAuditWrite)
	defer func() { finish(err) }()
	return a.delegate.SaveAuditRecord(ctx, record)
}

func (a observedAudit) ListTenantAuditRecords(ctx context.Context, tenantID tenant.ID, page ports.AuditRecordPageRequest) (result []audit.Record, err error) {
	ctx, finish := a.telemetry.Start(ctx, ports.OperationAuditRead)
	defer func() { finish(err) }()
	return a.delegate.ListTenantAuditRecords(ctx, tenantID, page)
}

func (a observedAudit) ListInventoryAuditRecords(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.AuditRecordPageRequest) (result []audit.Record, err error) {
	ctx, finish := a.telemetry.Start(ctx, ports.OperationAuditRead)
	defer func() { finish(err) }()
	return a.delegate.ListInventoryAuditRecords(ctx, tenantID, inventoryID, page)
}

func (a observedAudit) ListAssetAuditRecords(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, targetID string, request ports.AssetAuditRecordListRequest) (result []audit.Record, err error) {
	ctx, finish := a.telemetry.Start(ctx, ports.OperationAuditRead)
	defer func() { finish(err) }()
	return a.delegate.ListAssetAuditRecords(ctx, tenantID, inventoryID, targetID, request)
}
