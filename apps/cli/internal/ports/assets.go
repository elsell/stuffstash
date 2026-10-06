package ports

type AssetQuery struct {
	Page            Page
	Lifecycle, Sort string
}
type Asset struct {
	ID                  string             `json:"id"`
	TenantID            string             `json:"tenantId"`
	InventoryID         string             `json:"inventoryId"`
	Title               string             `json:"title"`
	Kind                string             `json:"kind"`
	Description         string             `json:"description"`
	Parent              *string            `json:"parentAssetId,omitempty"`
	Lifecycle           string             `json:"lifecycleState"`
	CreatedAt           string             `json:"createdAt"`
	UpdatedAt           string             `json:"updatedAt"`
	CustomAssetTypeID   *string            `json:"customAssetTypeId,omitempty"`
	CustomFields        map[string]any     `json:"customFields"`
	Expiration          *Expiration        `json:"expiration"`
	ExpirationContext   *ExpirationContext `json:"expirationContext,omitempty"`
	Tags                []CompactTag       `json:"tags"`
	CurrentCheckout     *CurrentCheckout   `json:"currentCheckout,omitempty"`
	PrimaryPhoto        *PrimaryPhoto      `json:"primaryPhoto,omitempty"`
	PrintJobID          *string            `json:"printJobId,omitempty"`
	UndoableOperationID *string            `json:"undoableOperationId,omitempty"`
}
type Expiration struct {
	Date      string `json:"date"`
	Precision string `json:"precision"`
}
type ExpirationContext struct {
	State           string `json:"state"`
	TrackingEnabled bool   `json:"trackingEnabled"`
	AdvanceDays     int64  `json:"advanceDays"`
	Timezone        string `json:"timezone"`
}
type CompactTag struct {
	ID          string  `json:"id"`
	Key         string  `json:"key"`
	DisplayName string  `json:"displayName"`
	Color       *string `json:"color,omitempty"`
}
type CurrentCheckout struct {
	ID                      string     `json:"id"`
	State                   string     `json:"state"`
	CheckedOutAt            string     `json:"checkedOutAt"`
	CheckedOutByPrincipalID string     `json:"checkedOutByPrincipalId"`
	CheckedOutByPrincipal   *Principal `json:"checkedOutByPrincipal,omitempty"`
}
type PrimaryPhoto struct {
	ID          string          `json:"id"`
	FileName    string          `json:"fileName"`
	ContentType string          `json:"contentType"`
	SizeBytes   int64           `json:"sizeBytes"`
	Thumbnails  PhotoThumbnails `json:"thumbnails"`
}
type PhotoThumbnails struct {
	Small  string `json:"small"`
	Medium string `json:"medium"`
	Large  string `json:"large"`
}
