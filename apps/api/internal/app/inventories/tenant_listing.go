package inventories

import (
	"context"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a Service) ListMyTenants(ctx context.Context, input ListMyTenantsInput) (ListMyTenantsResult, error) {
	limit := appsupport.PageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	afterTenantID, err := decodeTenantCursor(input.Cursor)
	if err != nil {
		return ListMyTenantsResult{}, apperrors.ErrInvalidInput
	}

	visible := make([]MyTenantAccess, 0, limit+1)
	items, err := a.tenants.ListTenants(ctx, ports.TenantListPageRequest{
		AfterTenantID: afterTenantID,
		Limit:         tenantScanLimit(limit),
	})
	if err != nil {
		return ListMyTenantsResult{}, err
	}

	lastScannedID := tenant.ID("")
	for _, item := range items {
		lastScannedID = item.ID
		access, ok, err := a.effectiveTenantAccess(ctx, input.Principal, item.ID)
		if err != nil {
			return ListMyTenantsResult{}, err
		}
		if !ok {
			continue
		}
		visible = append(visible, MyTenantAccess{Tenant: item, Access: access})
	}

	hasMore := len(visible) > limit
	var nextCursor *string
	if hasMore {
		visible = visible[:limit]
		nextCursor = encodeTenantCursor(visible[len(visible)-1].Tenant.ID)
	} else if len(items) == tenantScanLimit(limit) {
		hasMore = true
		nextCursor = encodeTenantCursor(lastScannedID)
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventTenantsListed,
		Message: "tenants listed",
		Fields: map[string]string{
			"principal_id": input.Principal.ID.String(),
			"limit":        strconv.Itoa(limit),
		},
	})
	for _, item := range visible {
		if err := appsupport.SaveReadAuditRecord(ctx, a.audit, a.ids, a.clock, appsupport.AuditRecordInput{
			Principal:  input.Principal,
			TenantID:   item.Tenant.ID,
			Source:     input.Source,
			RequestID:  input.RequestID,
			Action:     audit.ActionTenantListed,
			TargetType: audit.TargetTenant,
			TargetID:   item.Tenant.ID.String(),
			Metadata: map[string]string{
				"limit": strconv.Itoa(limit),
			},
		}); err != nil {
			return ListMyTenantsResult{}, err
		}
	}

	return ListMyTenantsResult{
		Items:      visible,
		Limit:      limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func tenantScanLimit(limit int) int {
	return limit*2 + 1
}

func encodeTenantCursor(id tenant.ID) *string {
	return appsupport.EncodePageCursor("my_tenants", "me", id.String())
}

func decodeTenantCursor(cursor string) (tenant.ID, error) {
	decoded, err := appsupport.DecodePageCursor("my_tenants", "me", cursor)
	if err != nil {
		return tenant.ID(""), err
	}
	if decoded == "" {
		return tenant.ID(""), nil
	}
	id, ok := tenant.NewID(decoded)
	if !ok {
		return tenant.ID(""), apperrors.ErrInvalidInput
	}
	return id, nil
}
