package ports

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
)

type InventoryExportFormat string

const (
	InventoryExportJSON InventoryExportFormat = "json"
	InventoryExportCSV  InventoryExportFormat = "csv"
)

var ErrInventoryExportLimit = errors.New("inventory export exceeds configured limit")
var ErrInventoryExportFormat = errors.New("inventory export format is invalid")

// InventoryExportDocument contains inventory data, never provider/access secrets.
// Encoders must project attachment metadata and must not serialize storage keys.
type InventoryExportDocument struct {
	SchemaVersion          int
	ExportedAt             time.Time
	TenantID               string
	InventoryID            string
	InventoryName          string
	Assets                 []InventoryExportAsset
	Tags                   []assettag.Tag
	CustomAssetTypes       []customfield.AssetType
	CustomFieldDefinitions []customfield.Definition
}
type InventoryExportAsset struct {
	ID                  string
	Title               string
	Description         string
	Kind                string
	ParentAssetID       string
	CustomAssetTypeID   string
	LifecycleState      string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	ExpirationDate      string
	ExpirationPrecision string
	CustomFields        map[string]any
	TagIDs              []string
	CurrentCheckout     *asset.Checkout
	Attachments         []media.Attachment
}
type InventoryExportEncoder interface {
	Encode(context.Context, InventoryExportDocument, InventoryExportFormat, int) ([]byte, error)
}
