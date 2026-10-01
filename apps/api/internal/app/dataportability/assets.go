package dataportability

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s Service) collectAssets(ctx context.Context, in ExportInput, doc *ports.InventoryExportDocument) error {
	attachments := 0
	return collectPages(ctx, s.deps.MaxRecords, func(after string, limit int) ([]asset.Asset, error) {
		return s.deps.Assets.ListAssetsByInventory(ctx, in.TenantID, in.InventoryID, ports.AssetListPageRequest{AfterAssetID: asset.ID(after), Limit: limit, LifecycleFilter: ports.AssetLifecycleFilterAll, Sort: ports.AssetListSortIDAsc})
	}, func(a asset.Asset) string { return a.ID.String() }, func(a asset.Asset) error {
		if a.TenantID.String() != in.TenantID.String() || a.InventoryID.String() != in.InventoryID.String() {
			return apperrors.ErrUnauthorized
		}
		tags, err := s.deps.Tags.AllAssetTagsByAsset(ctx, in.TenantID, in.InventoryID, a.ID)
		if err != nil {
			return err
		}
		row := ports.InventoryExportAsset{ID: a.ID.String(), Title: a.Title.String(), Description: a.Description.String(), Kind: a.Kind.String(), ParentAssetID: a.ParentAssetID.String(), CustomAssetTypeID: a.CustomAssetTypeID.String(), LifecycleState: a.LifecycleState.String(), CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt, ExpirationDate: a.Expiration.Value(), ExpirationPrecision: string(a.Expiration.Precision()), CustomFields: a.CustomFields.Values(), TagIDs: []string{}, Attachments: []media.Attachment{}}
		for _, tag := range tags {
			if tag.TenantID.String() != in.TenantID.String() || tag.InventoryID.String() != in.InventoryID.String() {
				return apperrors.ErrUnauthorized
			}
			row.TagIDs = append(row.TagIDs, tag.ID.String())
		}
		checkout, found, err := s.deps.Checkouts.CurrentAssetCheckout(ctx, in.TenantID, in.InventoryID, a.ID)
		if err != nil {
			return err
		}
		if found {
			if checkout.TenantID.String() != in.TenantID.String() || checkout.InventoryID.String() != in.InventoryID.String() || checkout.AssetID != a.ID {
				return apperrors.ErrUnauthorized
			}
			row.CurrentCheckout = &checkout
		}
		err = collectPages(ctx, s.deps.MaxRecords, func(after string, limit int) ([]media.Attachment, error) {
			return s.deps.Attachments.ListAttachmentsByAsset(ctx, in.TenantID, in.InventoryID, a.ID, ports.AttachmentListPageRequest{IncludeArchived: true, AfterAttachmentID: media.ID(after), Limit: limit})
		}, func(m media.Attachment) string { return m.ID.String() }, func(m media.Attachment) error {
			if m.TenantID.String() != in.TenantID.String() || m.InventoryID.String() != in.InventoryID.String() || m.AssetID.String() != a.ID.String() {
				return apperrors.ErrUnauthorized
			}
			attachments++
			if attachments > s.deps.MaxRecords {
				return ports.ErrInventoryExportLimit
			}
			row.Attachments = append(row.Attachments, m)
			return nil
		})
		if err != nil {
			return err
		}
		doc.Assets = append(doc.Assets, row)
		return nil
	})
}
