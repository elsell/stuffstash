package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func (c *Client) AuditRecords(ctx context.Context, q ports.AuditQuery) (ports.Result[[]ports.AuditRecord], error) {
	var response *http.Response
	var err error
	switch q.Level {
	case ports.HouseholdAudit:
		response, err = c.sdk.GetTenantsByTenantIdAuditRecords(ctx, q.Scope.Tenant, &generated.GetTenantsByTenantIdAuditRecordsParams{Limit: &q.Page.Limit, Cursor: &q.Page.Cursor})
	case ports.InventoryAudit:
		response, err = c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdAuditRecords(ctx, q.Scope.Tenant, q.Scope.Inventory, &generated.GetTenantsByTenantIdInventoriesByInventoryIdAuditRecordsParams{Limit: &q.Page.Limit, Cursor: &q.Page.Cursor})
	case ports.AssetAudit:
		response, err = c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdAuditRecords(ctx, q.Scope.Tenant, q.Scope.Inventory, q.AssetID, &generated.GetTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdAuditRecordsParams{Limit: &q.Page.Limit})
	default:
		return ports.Result[[]ports.AuditRecord]{}, ports.Failure("usage", "Select household, inventory, or asset audit history.")
	}
	r, err := read[generated.SuccessEnvelopeListRecordResponse](response, err)
	if err != nil {
		return ports.Result[[]ports.AuditRecord]{}, err
	}
	var records []ports.AuditRecord
	if r.Data.GetOrEmpty() != nil {
		records = make([]ports.AuditRecord, 0, len(r.Data.GetOrEmpty()))
	}
	for _, v := range r.Data.GetOrEmpty() {
		record := ports.AuditRecord{ID: v.Id, TenantID: v.TenantId, InventoryID: v.InventoryId, PrincipalID: v.PrincipalId, RequestID: v.RequestId, Action: v.Action, Source: v.Source, TargetType: v.TargetType, TargetID: v.TargetId, OccurredAt: v.OccurredAt, Metadata: v.Metadata}
		if v.Principal != nil {
			record.Principal = &ports.AuditPrincipal{ID: v.Principal.Id, Email: v.Principal.Email}
		}
		records = append(records, record)
	}
	return ports.Result[[]ports.AuditRecord]{Data: records, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
