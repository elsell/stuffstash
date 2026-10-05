package ports

import "context"

type AssetSearch interface {
	SearchAssets(context.Context, Scope, SearchQuery) (Result[[]SearchResult], error)
}
type SearchQuery struct {
	Page                                          Page
	Query, Mode, TypeID, Lifecycle, CheckoutState string
	TagIDs                                        []string
}
type SearchInventory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type SearchAncestor struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
type SearchMatch struct {
	Field string `json:"field"`
	Value string `json:"value"`
}
type SearchResult struct {
	Type         string           `json:"type"`
	TenantID     string           `json:"tenantId"`
	Inventory    SearchInventory  `json:"inventory"`
	Asset        SearchAsset      `json:"asset"`
	Matches      []SearchMatch    `json:"matches"`
	AncestorPath []SearchAncestor `json:"ancestorPath"`
}
type SearchAsset struct {
	ID                string             `json:"id"`
	InventoryID       string             `json:"inventoryId"`
	Title             string             `json:"title"`
	Kind              string             `json:"kind"`
	Description       string             `json:"description"`
	Parent            *string            `json:"parentAssetId,omitempty"`
	Lifecycle         string             `json:"lifecycleState"`
	CreatedAt         string             `json:"createdAt"`
	UpdatedAt         string             `json:"updatedAt"`
	CustomAssetTypeID *string            `json:"customAssetTypeId,omitempty"`
	CustomFields      map[string]any     `json:"customFields"`
	Expiration        *Expiration        `json:"expiration,omitempty"`
	ExpirationContext *ExpirationContext `json:"expirationContext,omitempty"`
	Tags              []CompactTag       `json:"tags"`
	CurrentCheckout   *CurrentCheckout   `json:"currentCheckout,omitempty"`
	PrimaryPhoto      *PrimaryPhoto      `json:"primaryPhoto,omitempty"`
}
