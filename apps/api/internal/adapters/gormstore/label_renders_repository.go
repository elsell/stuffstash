package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func (s Store) SaveLabelRender(ctx context.Context, value printing.LabelRender, record audit.Record) error {
	if record.TenantID.String() != value.TenantID || record.InventoryID.String() != value.InventoryID || record.TargetID != value.AssetID {
		return ports.ErrForbidden
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockLabelAsset(tx, value.TenantID, value.InventoryID, value.AssetID); err != nil {
			return err
		}
		var label labelModel
		if err := tx.Where(&labelModel{ID: string(value.LabelID), TenantID: value.TenantID, InventoryID: value.InventoryID, AssetID: value.AssetID}).First(&label).Error; err != nil {
			return err
		}
		if label.Tombstoned {
			return ports.ErrForbidden
		}
		model := labelRenderModelFrom(value)
		if err := tx.Create(&model).Error; err != nil {
			return err
		}
		return createAuditRecord(tx, record)
	})
}
func (s Store) LabelRenderByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, id printing.RenderID) (printing.LabelRender, bool, error) {
	if tenantID == "" || inventoryID == "" || id == "" {
		return printing.LabelRender{}, false, ports.ErrForbidden
	}
	var model labelRenderModel
	err := s.db.WithContext(ctx).Where(&labelRenderModel{ID: string(id), TenantID: tenantID.String(), InventoryID: inventoryID.String()}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return printing.LabelRender{}, false, nil
	}
	return model.value(), err == nil, err
}
func (s Store) PurgeExpiredLabelRenders(ctx context.Context, now time.Time) error {
	return s.db.WithContext(ctx).Where(clause.Lte{Column: "expires_at", Value: now}).Delete(&labelRenderModel{}).Error
}
