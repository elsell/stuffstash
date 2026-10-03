package assets

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"strconv"
)

func (s Service) AuthorizeAssetCreation(ctx context.Context, input ports.CreateAssetInput) error {
	return s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionCreateAsset)
}
func (s Service) PrepareCreateAssetForPrint(ctx context.Context, input ports.CreateAssetInput) (ports.PreparedCreateAsset, error) {
	p, err := s.PrepareCreateAsset(ctx, input)
	if err != nil {
		return p, err
	}
	if len(input.TagIDs) == 0 {
		return p, nil
	}
	p.TagIDs, err = s.validateAssignableAssetTagIDs(ctx, input.TenantID, input.InventoryID, input.TagIDs)
	if err != nil {
		return p, err
	}
	record, err := s.newAuditRecord(auditRecordInput{Principal: input.Principal, TenantID: input.TenantID, InventoryID: input.InventoryID, Source: input.Source, RequestID: input.RequestID, Action: audit.ActionAssetUpdated, TargetType: audit.TargetAsset, TargetID: p.Asset.ID.String(), Metadata: map[string]string{"asset_id": p.Asset.ID.String(), "tag_count": strconv.Itoa(len(p.TagIDs))}})
	if err != nil {
		return p, err
	}
	p.TagAudit = &record
	return p, nil
}
