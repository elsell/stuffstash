package ports

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
)

// ActionPlanCustomizationPreparation validates mutations without committing them.
type ActionPlanCustomizationPreparation interface {
	PrepareInventoryCustomAssetType(ctx context.Context, input CreateCustomAssetTypeInput) (PreparedCustomAssetType, error)
	PrepareInventoryCustomFieldDefinition(ctx context.Context, input CreateCustomFieldDefinitionInput) (PreparedCustomFieldDefinition, error)
	RecordCustomAssetTypeCreated(ctx context.Context, input CreateCustomAssetTypeInput, assetType customfield.AssetType)
	RecordCustomFieldDefinitionCreated(ctx context.Context, input CreateCustomFieldDefinitionInput, definition customfield.Definition)
}
