package search

import (
	"context"
	"strconv"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
)

func (a Service) saveSearchAssetsReadAudit(ctx context.Context, input SearchAssetsInput, authorizedInventoryIDs []inventory.InventoryID, limit int, mode string, lifecycle string, checkout string, customAssetTypeID string, resultCount int) error {
	if input.Source.String() == "" {
		return nil
	}
	targetType := audit.TargetTenant
	targetID := input.TenantID.String()
	inventoryID := inventory.InventoryID("")
	scope := "tenant"
	if len(input.InventoryIDs) == 1 {
		targetType = audit.TargetInventory
		targetID = input.InventoryIDs[0].String()
		inventoryID = input.InventoryIDs[0]
		scope = "inventory"
	}
	return appsupport.SaveReadAuditRecord(ctx, a.deps.Audit, a.deps.IDs, a.deps.Clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: inventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionAssetSearched,
		TargetType:  targetType,
		TargetID:    targetID,
		Metadata: map[string]string{
			"scope":                    scope,
			"limit":                    strconv.Itoa(limit),
			"mode":                     mode,
			"lifecycle":                lifecycle,
			"checkout":                 checkout,
			"custom_asset_type_filter": strconv.FormatBool(strings.TrimSpace(customAssetTypeID) != ""),
			"authorized_inventories":   strconv.Itoa(len(authorizedInventoryIDs)),
			"result_count":             strconv.Itoa(resultCount),
		},
	})
}
