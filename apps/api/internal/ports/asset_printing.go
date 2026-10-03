package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

// PreparedAssetPrint is committed as one boundary, including optional promotion/tags.
type PreparedAssetPrint struct {
	Asset      PreparedCreateAsset
	Label      printing.Label
	LabelAudit audit.Record
	Job        PrintJobCreate
}
type AssetPrintResult struct {
	Asset   asset.Asset
	Job     printing.Job
	Created bool
}
type AssetPrintUnitOfWork interface {
	CreateAssetWithPrint(context.Context, PreparedAssetPrint) (AssetPrintResult, error)
}

// Asset preparation stays behind a port so printing coordinates, rather than
// duplicating, asset authorization and validation.
type AssetPrintPreparation interface {
	AuthorizeAssetCreation(context.Context, CreateAssetInput) error
	PrepareCreateAssetForPrint(context.Context, CreateAssetInput) (PreparedCreateAsset, error)
	RecordAssetCreated(context.Context, asset.Asset, identity.PrincipalID)
}
