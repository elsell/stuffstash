package ports

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
)

// ActionPlanAssetPreparation validates mutations without committing them.
type ActionPlanAssetPreparation interface {
	PrepareArchiveAsset(ctx context.Context, input UpdateAssetLifecycleInput) (PreparedUpdateAssetLifecycle, error)
	PrepareCheckoutAsset(ctx context.Context, input CheckoutAssetInput) (PreparedCheckoutOperation, error)
	PrepareCreateAsset(ctx context.Context, input CreateAssetInput) (PreparedCreateAsset, error)
	PrepareCreateAssetWithPendingParents(ctx context.Context, input CreateAssetInput, pendingParents map[asset.ID]asset.Kind) (PreparedCreateAsset, error)
	PrepareRestoreAsset(ctx context.Context, input UpdateAssetLifecycleInput) (PreparedUpdateAssetLifecycle, error)
	PrepareReturnAsset(ctx context.Context, input ReturnAssetInput) (PreparedCheckoutOperation, error)
	PrepareUpdateAsset(ctx context.Context, input UpdateAssetInput) (PreparedUpdateAsset, error)
	PrepareUpdateAssetWithPendingParents(ctx context.Context, input UpdateAssetInput, pendingParents map[asset.ID]asset.Kind) (PreparedUpdateAsset, error)
	RecordAssetCheckedOut(ctx context.Context, checkout asset.Checkout, principalID identity.PrincipalID)
	RecordAssetCreated(ctx context.Context, item asset.Asset, principalID identity.PrincipalID)
	RecordAssetLifecycleUpdated(ctx context.Context, prepared PreparedUpdateAssetLifecycle, principalID identity.PrincipalID)
	RecordAssetReturned(ctx context.Context, checkout asset.Checkout, principalID identity.PrincipalID)
	RecordAssetUpdated(ctx context.Context, item asset.Asset, principalID identity.PrincipalID)
}
