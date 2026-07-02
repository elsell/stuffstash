package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type SearchAssetsInput struct {
	Authorization     string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID          string `path:"tenantId" doc:"Tenant ID"`
	Query             string `query:"q" required:"true" minLength:"1" maxLength:"120" doc:"Search query"`
	Mode              string `query:"mode" enum:"fuzzy,exact" doc:"Search mode; defaults to fuzzy"`
	CustomAssetTypeID string `query:"customAssetTypeId" doc:"Custom asset type filter"`
	LifecycleState    string `query:"lifecycleState" enum:"active,archived,all" doc:"Lifecycle filter; defaults to active"`
	Limit             int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor            string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type SearchAssetsOutput struct {
	Body shared.SuccessEnvelope[[]AssetSearchResultResponse]
}

type AssetSearchResultResponse struct {
	Type      string           `json:"type"`
	TenantID  string           `json:"tenantId"`
	Inventory InventorySummary `json:"inventory"`
	Asset     AssetSummary     `json:"asset"`
	Matches   []SearchMatch    `json:"matches"`
}

type InventorySummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AssetSummary struct {
	ID                string             `json:"id"`
	InventoryID       string             `json:"inventoryId"`
	ParentAssetID     string             `json:"parentAssetId,omitempty"`
	CustomAssetTypeID string             `json:"customAssetTypeId,omitempty"`
	Kind              string             `json:"kind"`
	Title             string             `json:"title"`
	Description       string             `json:"description"`
	CustomFields      map[string]any     `json:"customFields"`
	LifecycleState    string             `json:"lifecycleState"`
	CreatedAt         string             `json:"createdAt"`
	UpdatedAt         string             `json:"updatedAt"`
	PrimaryPhoto      *AssetPrimaryPhoto `json:"primaryPhoto,omitempty"`
}

type AssetPrimaryPhoto = shared.AssetPrimaryPhoto
type AssetPhotoThumbnails = shared.AssetPhotoThumbnails

type SearchMatch struct {
	Field string `json:"field"`
	Value string `json:"value"`
}
