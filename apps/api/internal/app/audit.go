package app

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/app/audithistory"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
)

type ListAuditRecordsInput = audithistory.ListAuditRecordsInput
type ListAssetAuditHistoryInput = audithistory.ListAssetAuditHistoryInput
type ListAuditRecordsResult = audithistory.ListAuditRecordsResult
type ListAssetAuditHistoryResult = audithistory.ListAssetAuditHistoryResult
type auditRecordInput = appsupport.AuditRecordInput

func (a App) newAuditRecord(input auditRecordInput) (audit.Record, error) {
	return appsupport.NewAuditRecord(a.ids, a.clock, input)
}
func (a App) saveReadAuditRecord(ctx context.Context, input auditRecordInput) error {
	return appsupport.SaveReadAuditRecord(ctx, a.audit, a.ids, a.clock, input)
}
func (a App) auditHistoryService() audithistory.Service {
	return audithistory.New(audithistory.Dependencies{
		Access:           a.inventoryService(),
		Assets:           a.assets,
		Audit:            a.audit,
		Authorizer:       a.authorizer,
		Clock:            a.clock,
		DefaultPageLimit: a.defaultPageLimit,
		IDs:              a.ids,
		MaxPageLimit:     a.maxPageLimit,
		Observer:         a.observer,
		Undoables:        a.undoables,
		Users:            a.users,
	})
}
func (a App) ListTenantAuditRecords(ctx context.Context, input ListAuditRecordsInput) (ListAuditRecordsResult, error) {
	return a.auditHistoryService().ListTenantAuditRecords(ctx, input)
}
func (a App) ListInventoryAuditRecords(ctx context.Context, input ListAuditRecordsInput) (ListAuditRecordsResult, error) {
	return a.auditHistoryService().ListInventoryAuditRecords(ctx, input)
}
func (a App) ListAssetAuditHistory(ctx context.Context, input ListAssetAuditHistoryInput) (ListAssetAuditHistoryResult, error) {
	return a.auditHistoryService().ListAssetAuditHistory(ctx, input)
}
