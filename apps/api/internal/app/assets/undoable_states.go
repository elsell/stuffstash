package assets

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func assetUndoableOperationStates(operation ports.UndoableOperation, direction ports.UndoableOperationDirection) (asset.Asset, asset.Asset, error) {
	if operation.AfterCheckout != nil {
		return asset.Asset{}, asset.Asset{}, apperrors.ErrInvalidInput
	}
	if operation.TargetType != audit.TargetAsset {
		return asset.Asset{}, asset.Asset{}, apperrors.ErrInvalidInput
	}
	after := operation.AfterAsset
	switch direction {
	case ports.UndoableOperationDirectionUndo:
		switch operation.OriginalAction {
		case audit.ActionAssetCreated:
			resulting := after
			resulting.LifecycleState = asset.LifecycleStateArchived
			return after, resulting, nil
		case audit.ActionAssetUpdated, audit.ActionAssetMoved, audit.ActionAssetArchived, audit.ActionAssetRestored:
			if operation.BeforeAsset == nil {
				return asset.Asset{}, asset.Asset{}, apperrors.ErrInvalidInput
			}
			return after, *operation.BeforeAsset, nil
		default:
			return asset.Asset{}, asset.Asset{}, apperrors.ErrInvalidInput
		}
	case ports.UndoableOperationDirectionRedo:
		switch operation.OriginalAction {
		case audit.ActionAssetCreated:
			expected := after
			expected.LifecycleState = asset.LifecycleStateArchived
			return expected, after, nil
		case audit.ActionAssetUpdated, audit.ActionAssetMoved, audit.ActionAssetArchived, audit.ActionAssetRestored:
			if operation.BeforeAsset == nil {
				return asset.Asset{}, asset.Asset{}, apperrors.ErrInvalidInput
			}
			return *operation.BeforeAsset, after, nil
		default:
			return asset.Asset{}, asset.Asset{}, apperrors.ErrInvalidInput
		}
	default:
		return asset.Asset{}, asset.Asset{}, apperrors.ErrInvalidInput
	}
}

func checkoutUndoableOperationStates(ctx context.Context, checkouts ports.AssetCheckoutRepository, operation ports.UndoableOperation, direction ports.UndoableOperationDirection) (asset.Checkout, asset.Checkout, error) {
	if operation.TargetType != audit.TargetAsset || operation.AfterCheckout == nil {
		return asset.Checkout{}, asset.Checkout{}, apperrors.ErrInvalidInput
	}
	if later, err := checkouts.HasLaterCheckout(ctx, *operation.AfterCheckout); err != nil {
		return asset.Checkout{}, asset.Checkout{}, err
	} else if later {
		return asset.Checkout{}, asset.Checkout{}, apperrors.ErrInvalidInput
	}
	after := *operation.AfterCheckout
	switch direction {
	case ports.UndoableOperationDirectionUndo:
		switch operation.OriginalAction {
		case audit.ActionAssetCheckedOut:
			resulting := after
			resulting.State = asset.CheckoutStateUndone
			resulting.ReturnedAt = time.Time{}
			resulting.ReturnedByPrincipal = ""
			resulting.ReturnDetails, _ = asset.NewCheckoutDetails("")
			return after, resulting, nil
		case audit.ActionAssetReturned:
			if operation.BeforeCheckout == nil {
				return asset.Checkout{}, asset.Checkout{}, apperrors.ErrInvalidInput
			}
			return after, *operation.BeforeCheckout, nil
		default:
			return asset.Checkout{}, asset.Checkout{}, apperrors.ErrInvalidInput
		}
	case ports.UndoableOperationDirectionRedo:
		switch operation.OriginalAction {
		case audit.ActionAssetCheckedOut:
			expected := after
			expected.State = asset.CheckoutStateUndone
			expected.ReturnedAt = time.Time{}
			expected.ReturnedByPrincipal = ""
			expected.ReturnDetails, _ = asset.NewCheckoutDetails("")
			return expected, after, nil
		case audit.ActionAssetReturned:
			if operation.BeforeCheckout == nil {
				return asset.Checkout{}, asset.Checkout{}, apperrors.ErrInvalidInput
			}
			return *operation.BeforeCheckout, after, nil
		default:
			return asset.Checkout{}, asset.Checkout{}, apperrors.ErrInvalidInput
		}
	default:
		return asset.Checkout{}, asset.Checkout{}, apperrors.ErrInvalidInput
	}
}
