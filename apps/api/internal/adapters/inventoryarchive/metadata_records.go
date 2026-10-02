package inventoryarchive

import (
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"time"
)

type metadataDocument struct {
	SchemaVersion          int             `json:"schemaVersion"`
	ExportedAt             time.Time       `json:"exportedAt"`
	TenantID               string          `json:"tenantId"`
	InventoryID            string          `json:"inventoryId"`
	InventoryName          string          `json:"inventoryName"`
	Assets                 []metadataAsset `json:"assets"`
	Tags                   []metadataTag   `json:"tags"`
	CustomAssetTypes       []metadataType  `json:"customAssetTypes"`
	CustomFieldDefinitions []metadataField `json:"customFieldDefinitions"`
}

type metadataTag struct {
	ID             assettag.ID             `json:"id"`
	Key            assettag.Key            `json:"key"`
	DisplayName    assettag.DisplayName    `json:"displayName"`
	Color          assettag.Color          `json:"color"`
	LifecycleState assettag.LifecycleState `json:"lifecycleState"`
	CreatedAt      time.Time               `json:"createdAt"`
	UpdatedAt      time.Time               `json:"updatedAt"`
}

type metadataType struct {
	ID                customfield.AssetTypeID             `json:"id"`
	Scope             customfield.Scope                   `json:"scope"`
	Key               customfield.Key                     `json:"key"`
	DisplayName       customfield.DisplayName             `json:"displayName"`
	Description       customfield.Description             `json:"description"`
	LifecycleState    customfield.AssetTypeLifecycleState `json:"lifecycleState"`
	ExpirationEnabled bool                                `json:"expirationEnabled"`
}

type metadataField struct {
	ID                 customfield.ID                       `json:"id"`
	Scope              customfield.Scope                    `json:"scope"`
	Key                customfield.Key                      `json:"key"`
	DisplayName        customfield.DisplayName              `json:"displayName"`
	Type               customfield.FieldType                `json:"type"`
	EnumOptions        []customfield.Key                    `json:"enumOptions"`
	Applicability      customfield.Applicability            `json:"applicability"`
	CustomAssetTypeIDs []customfield.AssetTypeID            `json:"customAssetTypeIds"`
	LifecycleState     customfield.DefinitionLifecycleState `json:"lifecycleState"`
}

type metadataAttachment struct {
	ID             media.ID             `json:"id"`
	FileName       media.FileName       `json:"fileName"`
	ContentType    media.ContentType    `json:"contentType"`
	SHA256         media.SHA256         `json:"sha256"`
	SizeBytes      int64                `json:"sizeBytes"`
	CreatedAt      time.Time            `json:"createdAt"`
	LifecycleState media.LifecycleState `json:"lifecycleState"`
}

type metadataCheckout struct {
	ID                    asset.CheckoutID      `json:"id"`
	State                 asset.CheckoutState   `json:"state"`
	CheckedOutAt          time.Time             `json:"checkedOutAt"`
	CheckedOutByPrincipal string                `json:"sourcePrincipalId"`
	CheckoutDetails       asset.CheckoutDetails `json:"details"`
	CreatedAt             time.Time             `json:"createdAt"`
	UpdatedAt             time.Time             `json:"updatedAt"`
}

type metadataAsset struct {
	ID                  string               `json:"id"`
	Title               string               `json:"title"`
	Description         string               `json:"description"`
	Kind                string               `json:"kind"`
	ParentAssetID       string               `json:"parentAssetId"`
	CustomAssetTypeID   string               `json:"customAssetTypeId"`
	LifecycleState      string               `json:"lifecycleState"`
	CreatedAt           time.Time            `json:"createdAt"`
	UpdatedAt           time.Time            `json:"updatedAt"`
	ExpirationDate      string               `json:"expirationDate"`
	ExpirationPrecision string               `json:"expirationPrecision"`
	CustomFields        map[string]any       `json:"customFields"`
	TagIDs              []string             `json:"tagIds"`
	CurrentCheckout     *metadataCheckout    `json:"currentCheckout"`
	Attachments         []metadataAttachment `json:"attachments"`
}
