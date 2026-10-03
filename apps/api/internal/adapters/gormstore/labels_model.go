package gormstore

import (
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"time"
)

type labelInstanceModel struct {
	Singleton  int    `gorm:"primaryKey;check:chk_label_instances_singleton,singleton = 1"`
	InstanceID string `gorm:"not null;size:26;uniqueIndex"`
}

func (labelInstanceModel) TableName() string { return "label_instances" }

type labelModel struct {
	ID          string    `gorm:"primaryKey;size:26"`
	InstanceID  string    `gorm:"not null;size:26"`
	TenantID    string    `gorm:"not null;size:26;uniqueIndex:idx_labels_asset"`
	InventoryID string    `gorm:"not null;size:26;uniqueIndex:idx_labels_asset"`
	AssetID     string    `gorm:"not null;size:26;uniqueIndex:idx_labels_asset"`
	Tombstoned  bool      `gorm:"not null"`
	CreatedAt   time.Time `gorm:"not null"`
}

func (labelModel) TableName() string { return "asset_labels" }
func (m labelModel) value() printing.Label {
	return printing.Label{ID: printing.LabelID(m.ID), InstanceID: printing.InstanceID(m.InstanceID), TenantID: m.TenantID, InventoryID: m.InventoryID, AssetID: m.AssetID, Tombstoned: m.Tombstoned, CreatedAt: m.CreatedAt}
}

type labelRenderModel struct {
	ID                   string    `gorm:"primaryKey;size:26"`
	TenantID             string    `gorm:"not null;size:26;index:idx_label_renders_asset"`
	InventoryID          string    `gorm:"not null;size:26;index:idx_label_renders_asset"`
	AssetID              string    `gorm:"not null;size:26;index:idx_label_renders_asset"`
	LabelID              string    `gorm:"not null;size:26"`
	SelectionFingerprint string    `gorm:"not null;size:64"`
	MediaFingerprint     string    `gorm:"not null;size:64"`
	ContentType          string    `gorm:"not null;size:80"`
	SHA256               string    `gorm:"not null;size:64"`
	Content              []byte    `gorm:"not null"`
	WidthPixels          int       `gorm:"not null"`
	HeightPixels         int       `gorm:"not null"`
	DisplayRotation      int       `gorm:"not null"`
	CreatedAt            time.Time `gorm:"not null"`
	ExpiresAt            time.Time `gorm:"not null;index:idx_label_renders_expiry"`
}

func (labelRenderModel) TableName() string { return "label_renders" }
func (m labelRenderModel) value() printing.LabelRender {
	return printing.LabelRender{ID: printing.RenderID(m.ID), TenantID: m.TenantID, InventoryID: m.InventoryID, AssetID: m.AssetID, LabelID: printing.LabelID(m.LabelID), SelectionFingerprint: m.SelectionFingerprint, MediaFingerprint: m.MediaFingerprint, ContentType: m.ContentType, SHA256: m.SHA256, Content: m.Content, WidthPixels: m.WidthPixels, HeightPixels: m.HeightPixels, DisplayRotation: m.DisplayRotation, CreatedAt: m.CreatedAt, ExpiresAt: m.ExpiresAt}
}
func labelRenderModelFrom(v printing.LabelRender) labelRenderModel {
	return labelRenderModel{ID: string(v.ID), TenantID: v.TenantID, InventoryID: v.InventoryID, AssetID: v.AssetID, LabelID: string(v.LabelID), SelectionFingerprint: v.SelectionFingerprint, MediaFingerprint: v.MediaFingerprint, ContentType: v.ContentType, SHA256: v.SHA256, Content: v.Content, WidthPixels: v.WidthPixels, HeightPixels: v.HeightPixels, DisplayRotation: v.DisplayRotation, CreatedAt: v.CreatedAt, ExpiresAt: v.ExpiresAt}
}
