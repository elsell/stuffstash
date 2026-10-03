package routes

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/labels/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/labels/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	labelapp "github.com/stuffstash/stuff-stash/internal/app/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

const scopePath = "/tenants/{tenantId}/inventories/{inventoryId}"

func labelError(err error) error {
	if errors.Is(err, labelapp.ErrLabelsUnavailable) {
		return huma.Error503ServiceUnavailable("Labels are not configured.")
	}
	return shared.ToHumaError(err)
}
func authenticate(ctx context.Context, application app.App, input dto.ScopeInput) (labelapp.LabelScope, error) {
	principal, err := shared.Authenticate(ctx, application, input.Authorization)
	if err != nil {
		return labelapp.LabelScope{}, err
	}
	if application.Labels() == nil {
		return labelapp.LabelScope{}, labelError(labelapp.ErrLabelsUnavailable)
	}
	return labelapp.LabelScope{Principal: principal, TenantID: tenant.ID(input.TenantID), InventoryID: inventory.InventoryID(input.InventoryID), RequestID: input.RequestID}, nil
}
func output(view printing.LabelView, err error) (*dto.LabelOutput, error) {
	if err != nil {
		return nil, labelError(err)
	}
	return &dto.LabelOutput{CacheControl: "private, no-store", Body: shared.SuccessEnvelope[dto.LabelResponse]{Data: mapper.Label(view), Meta: shared.Meta{TenantID: view.Label.TenantID}}}, nil
}
func Register(api huma.API, application app.App) {
	huma.Get(api, "/instance", func(ctx context.Context, _ *struct{}) (*dto.InstanceOutput, error) {
		id, err := application.Labels().Instance(ctx)
		if err != nil {
			return nil, labelError(err)
		}
		return &dto.InstanceOutput{CacheControl: "no-store", Body: shared.SuccessEnvelope[dto.InstanceResponse]{Data: dto.InstanceResponse{ProtocolVersion: 1, InstanceID: string(id)}}}, nil
	}, huma.OperationTags("labels"))
	huma.Get(api, "/labels/v1/{instanceId}/{labelId}", func(ctx context.Context, input *dto.ResolveInput) (*dto.LabelOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		if application.Labels() == nil {
			return nil, labelError(labelapp.ErrLabelsUnavailable)
		}
		return output(application.Labels().Resolve(ctx, principal, input.InstanceID, input.LabelID, input.RequestID))
	}, huma.OperationTags("labels"), shared.SecuredOperation)
	huma.Post(api, scopePath+"/assets/{assetId}/label", func(ctx context.Context, input *dto.AssetInput) (*dto.LabelOutput, error) {
		scope, err := authenticate(ctx, application, input.ScopeInput)
		if err != nil {
			return nil, err
		}
		return output(application.Labels().Provision(ctx, scope, asset.ID(input.AssetID)))
	}, huma.OperationTags("labels"), shared.SecuredOperation, func(op *huma.Operation) { op.DefaultStatus = 200 })
	huma.Get(api, scopePath+"/assets/{assetId}/label", func(ctx context.Context, input *dto.AssetInput) (*dto.LabelOutput, error) {
		scope, err := authenticate(ctx, application, input.ScopeInput)
		if err != nil {
			return nil, err
		}
		return output(application.Labels().Get(ctx, scope, asset.ID(input.AssetID)))
	}, huma.OperationTags("labels"), shared.SecuredOperation)
	registerRenders(api, application)
}
