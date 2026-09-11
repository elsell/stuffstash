package assets

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"strings"
)

type ExpirationInput struct {
	Date      string
	Precision string
}
type ExpirationUpdate struct {
	Present bool
	Value   *ExpirationInput
}

func (s Service) validateExpiration(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, typeID asset.CustomAssetTypeID, input *ExpirationInput) (expirationdate.Date, error) {
	if input == nil {
		return expirationdate.Date{}, nil
	}
	date, err := expirationdate.ParseDate(input.Date, expirationdate.Precision(input.Precision))
	if err != nil || typeID == "" || s.customAssetTypes == nil {
		return expirationdate.Date{}, apperrors.ErrInvalidInput
	}
	kind, found, err := s.customAssetTypes.CustomAssetTypeByID(ctx, tenantID, inventoryID, customfield.AssetTypeID(typeID.String()))
	if err != nil {
		return expirationdate.Date{}, err
	}
	if !found || !kind.IsActive() || !kind.ExpirationEnabled {
		return expirationdate.Date{}, apperrors.ErrInvalidInput
	}
	return date, nil
}

func (s Service) applyAssetTypeAndExpiration(ctx context.Context, input UpdateAssetInput, current asset.Asset) (asset.Asset, bool, error) {
	updated := current
	fieldsChanged := false

	if input.CustomAssetTypeID != nil {
		if strings.TrimSpace(*input.CustomAssetTypeID) == "" {
			return asset.Asset{}, false, apperrors.ErrInvalidInput
		}
		if current.CustomAssetTypeID != "" && current.CustomAssetTypeID.String() != strings.TrimSpace(*input.CustomAssetTypeID) {
			return asset.Asset{}, false, apperrors.ErrInvalidInput
		}
		if current.CustomAssetTypeID == "" {
			typeID, err := s.validatedAssetCustomAssetTypeID(ctx, input.TenantID, input.InventoryID, *input.CustomAssetTypeID)
			if err != nil {
				return asset.Asset{}, false, err
			}
			values := current.CustomFields.Values()
			if input.CustomFields != nil {
				values = input.CustomFields
			}
			if _, err := s.validatedCustomFields(ctx, input.TenantID, input.InventoryID, typeID, values); err != nil {
				return asset.Asset{}, false, err
			}
			updated.CustomAssetTypeID = typeID
			fieldsChanged = true
		}
	}

	if input.Expiration.Present {
		date, err := s.validateExpiration(ctx, input.TenantID, input.InventoryID, updated.CustomAssetTypeID, input.Expiration.Value)
		if err != nil {
			return asset.Asset{}, false, err
		}
		updated.Expiration = date
		fieldsChanged = fieldsChanged || date != current.Expiration
	}

	return updated, fieldsChanged, nil
}
