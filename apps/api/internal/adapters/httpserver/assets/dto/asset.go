package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type CreateAssetInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Body          CreateAssetBody
}

type CreateAssetBody struct {
	Kind              string         `json:"kind" enum:"item,container,location" doc:"Asset kind"`
	Title             string         `json:"title" maxLength:"160" doc:"Asset title"`
	Description       string         `json:"description,omitempty" doc:"Asset description"`
	ParentAssetID     string         `json:"parentAssetId,omitempty" doc:"Parent asset ID"`
	CustomAssetTypeID string         `json:"customAssetTypeId,omitempty" doc:"Custom asset type ID"`
	CustomFields      map[string]any `json:"customFields,omitempty" doc:"Custom field values"`
}

type CreateAssetOutput struct {
	Body shared.SuccessEnvelope[AssetResponse]
}

type UpdateAssetInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	AssetID       string `path:"assetId" doc:"Asset ID"`
	Body          UpdateAssetBody
}

type UpdateAssetBody struct {
	Title         *string               `json:"title,omitempty" maxLength:"160" doc:"Asset title"`
	Description   *string               `json:"description,omitempty" doc:"Asset description"`
	ParentAssetID shared.NullableString `json:"parentAssetId,omitempty" doc:"Parent asset ID, or null to move to inventory root"`
	CustomFields  map[string]any        `json:"customFields,omitempty" doc:"Custom field values"`
}

type UpdateAssetOutput struct {
	Body shared.SuccessEnvelope[AssetResponse]
}

type GetAssetInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	AssetID       string `path:"assetId" doc:"Asset ID"`
}

type GetAssetOutput struct {
	Body shared.SuccessEnvelope[AssetResponse]
}

type UpdateAssetLifecycleInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	AssetID       string `path:"assetId" doc:"Asset ID"`
}

type UpdateAssetLifecycleOutput struct {
	Body shared.SuccessEnvelope[AssetResponse]
}

type DeleteAssetOutput struct{}

type ListAssetsInput struct {
	Authorization  string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID      string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID       string `path:"tenantId" doc:"Tenant ID"`
	InventoryID    string `path:"inventoryId" doc:"Inventory ID"`
	Limit          int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor         string `query:"cursor" doc:"Opaque cursor from the previous page"`
	LifecycleState string `query:"lifecycleState" enum:"active,archived,all" doc:"Lifecycle filter; defaults to active"`
	Sort           string `query:"sort" enum:"id_asc,updated_desc" doc:"Sort order; defaults to id_asc"`
}

type ListAssetsOutput struct {
	Body shared.SuccessEnvelope[[]AssetResponse]
}

type AssetResponse struct {
	ID                string             `json:"id"`
	TenantID          string             `json:"tenantId"`
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
