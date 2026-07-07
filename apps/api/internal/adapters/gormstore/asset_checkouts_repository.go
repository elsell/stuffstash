package gormstore

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) CheckOutAsset(ctx context.Context, checkout asset.Checkout, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return checkOutAssetInTx(tx, checkout, auditRecord, undoableOperation)
	})
}

func checkOutAssetInTx(tx *gorm.DB, checkout asset.Checkout, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	var item assetModel
	err := tx.Where(&assetModel{
		ID:             checkout.AssetID.String(),
		TenantID:       checkout.TenantID.String(),
		InventoryID:    checkout.InventoryID.String(),
		LifecycleState: asset.LifecycleStateActive.String(),
	}).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.ErrForbidden
	}
	if err != nil {
		return err
	}
	if checkout.State != asset.CheckoutStateOpen {
		return ports.ErrForbidden
	}
	model := newAssetCheckoutModel(checkout)
	if err := tx.Create(&model).Error; err != nil {
		if assetCheckoutUniqueConflict(err) {
			return ports.ErrConflict
		}
		return err
	}
	if err := createAuditRecord(tx, auditRecord); err != nil {
		return err
	}
	return createUndoableOperation(tx, undoableOperation)
}

func assetCheckoutUniqueConflict(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	return strings.Contains(message, "idx_asset_checkouts_one_open") ||
		strings.Contains(message, "UNIQUE constraint failed") ||
		strings.Contains(message, "duplicate key value")
}

func (s Store) ReturnAsset(ctx context.Context, expectedCurrent asset.Checkout, returned asset.Checkout, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return returnAssetInTx(tx, expectedCurrent, returned, auditRecord, undoableOperation)
	})
}

func returnAssetInTx(tx *gorm.DB, expectedCurrent asset.Checkout, returned asset.Checkout, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	var model assetCheckoutModel
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&assetCheckoutModel{ID: expectedCurrent.ID.String(), TenantID: expectedCurrent.TenantID.String(), InventoryID: expectedCurrent.InventoryID.String()}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.ErrConflict
	}
	if err != nil {
		return err
	}
	current, ok := model.toDomain()
	if !ok {
		return fmt.Errorf("invalid asset checkout row %q", model.ID)
	}
	if !asset.CheckoutsEquivalentForStaleCheck(current, expectedCurrent) {
		return ports.ErrConflict
	}
	if current.State != asset.CheckoutStateOpen || returned.ID != current.ID || returned.TenantID != current.TenantID || returned.InventoryID != current.InventoryID || returned.AssetID != current.AssetID || returned.State != asset.CheckoutStateReturned || returned.ReturnedAt.IsZero() || returned.ReturnedByPrincipal == "" {
		return ports.ErrForbidden
	}
	if err := tx.Model(&model).Updates(assetCheckoutUpdateMap(returned)).Error; err != nil {
		return err
	}
	if err := createAuditRecord(tx, auditRecord); err != nil {
		return err
	}
	return createUndoableOperation(tx, undoableOperation)
}

func (s Store) CurrentAssetCheckout(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (asset.Checkout, bool, error) {
	var model assetCheckoutModel
	err := s.db.WithContext(ctx).Where(&assetCheckoutModel{
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
		AssetID:     assetID.String(),
		State:       asset.CheckoutStateOpen.String(),
	}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return asset.Checkout{}, false, nil
	}
	if err != nil {
		return asset.Checkout{}, false, err
	}
	checkout, ok := model.toDomain()
	if !ok {
		return asset.Checkout{}, false, fmt.Errorf("invalid asset checkout row %q", model.ID)
	}
	return checkout, true, nil
}

func (s Store) CurrentAssetCheckouts(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetIDs []asset.ID) (map[asset.ID]asset.Checkout, error) {
	if len(assetIDs) == 0 {
		return nil, nil
	}
	values := make([]any, 0, len(assetIDs))
	for _, assetID := range assetIDs {
		if assetID.String() != "" {
			values = append(values, assetID.String())
		}
	}
	if len(values) == 0 {
		return nil, nil
	}
	var models []assetCheckoutModel
	err := s.db.WithContext(ctx).
		Where(&assetCheckoutModel{
			TenantID:    tenantID.String(),
			InventoryID: inventoryID.String(),
			State:       asset.CheckoutStateOpen.String(),
		}).
		Where(clause.IN{Column: clause.Column{Name: "asset_id"}, Values: values}).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	checkouts := make(map[asset.ID]asset.Checkout, len(models))
	for _, model := range models {
		checkout, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid asset checkout row %q", model.ID)
		}
		checkouts[checkout.AssetID] = checkout
	}
	return checkouts, nil
}

func (s Store) AssetCheckoutByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, checkoutID asset.CheckoutID) (asset.Checkout, bool, error) {
	var model assetCheckoutModel
	err := s.db.WithContext(ctx).Where(&assetCheckoutModel{ID: checkoutID.String(), TenantID: tenantID.String(), InventoryID: inventoryID.String()}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return asset.Checkout{}, false, nil
	}
	if err != nil {
		return asset.Checkout{}, false, err
	}
	checkout, ok := model.toDomain()
	if !ok {
		return asset.Checkout{}, false, fmt.Errorf("invalid asset checkout row %q", model.ID)
	}
	return checkout, true, nil
}

func (s Store) ListAssetCheckoutHistory(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, page ports.AssetCheckoutHistoryPageRequest) ([]asset.Checkout, error) {
	query := s.db.WithContext(ctx).Where(&assetCheckoutModel{TenantID: tenantID.String(), InventoryID: inventoryID.String(), AssetID: assetID.String()})
	if page.AfterCheckoutID.String() != "" {
		query = query.Where(clause.Or(
			clause.Lt{Column: clause.Column{Name: "checked_out_at"}, Value: page.AfterCheckedOutAt},
			clause.And(
				clause.Eq{Column: clause.Column{Name: "checked_out_at"}, Value: page.AfterCheckedOutAt},
				clause.Lt{Column: clause.Column{Name: "id"}, Value: page.AfterCheckoutID.String()},
			),
		))
	}
	return listAssetCheckouts(query, page.Limit)
}

func (s Store) ListCheckedOutAssets(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.CheckedOutAssetsPageRequest) ([]ports.CheckedOutAsset, error) {
	query := s.db.WithContext(ctx).
		Preload("Asset", func(db *gorm.DB) *gorm.DB {
			return db.Where(&assetModel{
				TenantID:    tenantID.String(),
				InventoryID: inventoryID.String(),
			})
		}).
		Where(&assetCheckoutModel{
			TenantID:    tenantID.String(),
			InventoryID: inventoryID.String(),
			State:       asset.CheckoutStateOpen.String(),
		})
	if page.AfterAssetID.String() != "" {
		query = query.Where(clause.Or(
			clause.Lt{Column: clause.Column{Name: "checked_out_at"}, Value: page.AfterCheckedOutAt},
			clause.And(
				clause.Eq{Column: clause.Column{Name: "checked_out_at"}, Value: page.AfterCheckedOutAt},
				clause.Lt{Column: clause.Column{Name: "asset_id"}, Value: page.AfterAssetID.String()},
			),
		))
	}
	var checkoutModels []assetCheckoutModel
	query = query.Order(clause.OrderBy{Columns: []clause.OrderByColumn{
		{Column: clause.Column{Name: "checked_out_at"}, Desc: true},
		{Column: clause.Column{Name: "asset_id"}, Desc: true},
	}})
	if page.Limit > 0 {
		query = query.Limit(page.Limit)
	}
	if err := query.Find(&checkoutModels).Error; err != nil {
		return nil, err
	}
	items := make([]ports.CheckedOutAsset, 0, len(checkoutModels))
	for _, checkoutModel := range checkoutModels {
		checkout, ok := checkoutModel.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid asset checkout row %q", checkoutModel.ID)
		}
		item, ok := checkoutModel.Asset.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid asset row %q", checkoutModel.Asset.ID)
		}
		if item.TenantID.String() != tenantID.String() || item.InventoryID.String() != inventoryID.String() || item.ID != checkout.AssetID {
			return nil, fmt.Errorf("checkout asset row %q is outside requested inventory scope", checkoutModel.Asset.ID)
		}
		items = append(items, ports.CheckedOutAsset{Asset: item, Checkout: checkout})
	}
	return items, nil
}

func (s Store) HasLaterCheckout(ctx context.Context, checkout asset.Checkout) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&assetCheckoutModel{}).
		Where(&assetCheckoutModel{
			TenantID:    checkout.TenantID.String(),
			InventoryID: checkout.InventoryID.String(),
			AssetID:     checkout.AssetID.String(),
		}).
		Where(clause.Neq{Column: clause.Column{Name: "id"}, Value: checkout.ID.String()}).
		Where(clause.Or(
			clause.Gt{Column: clause.Column{Name: "checked_out_at"}, Value: checkout.CheckedOutAt},
			clause.And(
				clause.Eq{Column: clause.Column{Name: "checked_out_at"}, Value: checkout.CheckedOutAt},
				clause.Gt{Column: clause.Column{Name: "id"}, Value: checkout.ID.String()},
			),
		)).
		Count(&count).Error
	return count > 0, err
}

func listAssetCheckouts(query *gorm.DB, limit int) ([]asset.Checkout, error) {
	var models []assetCheckoutModel
	query = query.Order(clause.OrderBy{Columns: []clause.OrderByColumn{
		{Column: clause.Column{Name: "checked_out_at"}, Desc: true},
		{Column: clause.Column{Name: "id"}, Desc: true},
	}})
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]asset.Checkout, 0, len(models))
	for _, model := range models {
		checkout, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid asset checkout row %q", model.ID)
		}
		items = append(items, checkout)
	}
	return items, nil
}

func assetCheckoutUpdateMap(checkout asset.Checkout) map[string]any {
	return map[string]any{
		"state":                 checkout.State.String(),
		"returned_at":           timePtrFromTime(checkout.ReturnedAt),
		"returned_by_principal": checkout.ReturnedByPrincipal,
		"return_details":        checkout.ReturnDetails.String(),
		"updated_at":            checkout.UpdatedAt,
	}
}
