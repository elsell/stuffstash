package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/notifications/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/notifications/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/app/notifications"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func Register(api huma.API, application app.App) {
	registerInbox(api, application)
	registerDevices(api, application)
	const path = "/tenants/{tenantId}/inventories/{inventoryId}/notification-preferences"
	huma.Get(api, path, func(ctx context.Context, input *dto.ScopeInput) (*dto.PreferencesOutput, error) {
		scope, err := authenticateScope(ctx, application, input)
		if err != nil {
			return nil, err
		}
		value, err := application.Notifications().GetPreferences(ctx, scope)
		return preferencesOutput(value, err)
	}, huma.OperationTags("notifications"), shared.SecuredOperation)
	huma.Post(api, path+"/initialize", func(ctx context.Context, input *dto.InitializeInput) (*dto.PreferencesOutput, error) {
		scope, err := authenticateScope(ctx, application, &input.ScopeInput)
		if err != nil {
			return nil, err
		}
		value, err := application.Notifications().InitializePreferences(ctx, scope, input.Body.Timezone)
		return preferencesOutput(value, err)
	}, huma.OperationTags("notifications"), shared.SecuredOperation, func(op *huma.Operation) { op.DefaultStatus = 200 })
	huma.Put(api, path, func(ctx context.Context, input *dto.UpdateInput) (*dto.PreferencesOutput, error) {
		scope, err := authenticateScope(ctx, application, &input.ScopeInput)
		if err != nil {
			return nil, err
		}
		value, err := application.Notifications().UpdatePreferences(ctx, scope, input.Body.Revision, mapper.PolicyToDomain(input.Body.Defaults), input.Body.Timezone, input.Body.PushEnabled)
		return preferencesOutput(value, err)
	}, huma.OperationTags("notifications"), shared.SecuredOperation)
	huma.Put(api, path+"/types/{customAssetTypeId}", func(ctx context.Context, input *dto.TypeOverrideInput) (*dto.PreferencesOutput, error) {
		scope, err := authenticateScope(ctx, application, &input.ScopeInput)
		if err != nil {
			return nil, err
		}
		policy := mapper.PolicyToDomain(input.Body.Settings)
		value, err := application.Notifications().SetTypeOverride(ctx, scope, input.Body.Revision, input.CustomAssetTypeID, &policy)
		return preferencesOutput(value, err)
	}, huma.OperationTags("notifications"), shared.SecuredOperation)
	huma.Delete(api, path+"/types/{customAssetTypeId}", func(ctx context.Context, input *dto.DeleteTypeOverrideInput) (*dto.PreferencesOutput, error) {
		scope, err := authenticateScope(ctx, application, &input.ScopeInput)
		if err != nil {
			return nil, err
		}
		value, err := application.Notifications().SetTypeOverride(ctx, scope, input.Revision, input.CustomAssetTypeID, nil)
		return preferencesOutput(value, err)
	}, huma.OperationTags("notifications"), shared.SecuredOperation, func(op *huma.Operation) { op.DefaultStatus = 200 })
}
func authenticateScope(ctx context.Context, application app.App, input *dto.ScopeInput) (notifications.ScopeInput, error) {
	principal, err := shared.Authenticate(ctx, application, input.Authorization)
	if err != nil {
		return notifications.ScopeInput{}, err
	}
	return notifications.ScopeInput{Principal: principal, TenantID: tenant.ID(input.TenantID), InventoryID: inventory.InventoryID(input.InventoryID), Source: audit.SourceAPI, RequestID: input.RequestID}, nil
}
func preferencesOutput(value ports.NotificationPreferencesRecord, err error) (*dto.PreferencesOutput, error) {
	if err != nil {
		return nil, shared.ToHumaError(err)
	}
	return &dto.PreferencesOutput{Body: shared.SuccessEnvelope[dto.PreferencesResponse]{Data: mapper.SettingsToResponse(value.Settings, value.Revision), Meta: shared.Meta{TenantID: value.Scope.TenantID.String()}}}, nil
}
