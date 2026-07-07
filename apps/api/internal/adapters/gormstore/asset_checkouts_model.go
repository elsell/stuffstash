package gormstore

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
)

type assetCheckoutModel struct {
	ID                    string      `gorm:"primaryKey;size:26"`
	TenantID              string      `gorm:"not null;size:26;index:idx_asset_checkouts_scope;uniqueIndex:idx_asset_checkouts_one_open,where:state = 'open'"`
	Tenant                tenantModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID           string      `gorm:"not null;size:26;index:idx_asset_checkouts_scope;uniqueIndex:idx_asset_checkouts_one_open,where:state = 'open'"`
	Inventory             inventoryModel
	AssetID               string     `gorm:"not null;size:26;index:idx_asset_checkouts_asset;uniqueIndex:idx_asset_checkouts_one_open,where:state = 'open'"`
	Asset                 assetModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:AssetID;references:ID"`
	State                 string     `gorm:"not null;size:32;index;check:chk_asset_checkouts_state,state IN ('open','returned','undone')"`
	CheckedOutAt          time.Time  `gorm:"not null;index:idx_asset_checkouts_checked_out"`
	CheckedOutByPrincipal string     `gorm:"not null;size:255"`
	CheckoutDetails       string     `gorm:"not null;default:'';size:1000"`
	ReturnedAt            *time.Time
	ReturnedByPrincipal   string `gorm:"not null;default:'';size:255"`
	ReturnDetails         string `gorm:"not null;default:'';size:1000"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (assetCheckoutModel) TableName() string {
	return "asset_checkouts"
}

func newAssetCheckoutModel(checkout asset.Checkout) assetCheckoutModel {
	return assetCheckoutModel{
		ID:                    checkout.ID.String(),
		TenantID:              checkout.TenantID.String(),
		InventoryID:           checkout.InventoryID.String(),
		AssetID:               checkout.AssetID.String(),
		State:                 checkout.State.String(),
		CheckedOutAt:          checkout.CheckedOutAt,
		CheckedOutByPrincipal: checkout.CheckedOutByPrincipal,
		CheckoutDetails:       checkout.CheckoutDetails.String(),
		ReturnedAt:            timePtrFromTime(checkout.ReturnedAt),
		ReturnedByPrincipal:   checkout.ReturnedByPrincipal,
		ReturnDetails:         checkout.ReturnDetails.String(),
		CreatedAt:             checkout.CreatedAt,
		UpdatedAt:             checkout.UpdatedAt,
	}
}

func (m assetCheckoutModel) toDomain() (asset.Checkout, bool) {
	id, ok := asset.NewCheckoutID(m.ID)
	if !ok {
		return asset.Checkout{}, false
	}
	assetID, ok := asset.NewID(m.AssetID)
	if !ok {
		return asset.Checkout{}, false
	}
	state, ok := asset.NewCheckoutState(m.State)
	if !ok {
		return asset.Checkout{}, false
	}
	checkoutDetails, ok := asset.NewCheckoutDetails(m.CheckoutDetails)
	if !ok {
		return asset.Checkout{}, false
	}
	returnDetails, ok := asset.NewCheckoutDetails(m.ReturnDetails)
	if !ok {
		return asset.Checkout{}, false
	}
	returnedAt := time.Time{}
	if m.ReturnedAt != nil {
		returnedAt = *m.ReturnedAt
	}
	return asset.Checkout{
		ID:                    id,
		TenantID:              asset.TenantID(m.TenantID),
		InventoryID:           asset.InventoryID(m.InventoryID),
		AssetID:               assetID,
		State:                 state,
		CheckedOutAt:          m.CheckedOutAt,
		CheckedOutByPrincipal: m.CheckedOutByPrincipal,
		CheckoutDetails:       checkoutDetails,
		ReturnedAt:            returnedAt,
		ReturnedByPrincipal:   m.ReturnedByPrincipal,
		ReturnDetails:         returnDetails,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
	}, true
}

func timePtrFromTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}
