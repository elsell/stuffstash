package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/notifications/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/notifications/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/notifications"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func registerDevices(api huma.API, application app.App) {
	const path = "/tenants/{tenantId}/inventories/{inventoryId}/notification-devices"
	huma.Post(api, path, func(ctx context.Context, input *dto.RegisterDeviceInput) (*dto.DeviceOutput, error) {
		scope, err := authenticateScope(ctx, application, &input.ScopeInput)
		if err != nil {
			return nil, err
		}
		token, err := notification.ParseDeviceToken(input.Body.Token)
		if err != nil {
			return nil, shared.ToHumaError(apperrors.ErrInvalidInput)
		}
		value, err := application.Notifications().RegisterDevice(ctx, scope, notifications.RegisterDeviceInput{InstallationID: input.Body.InstallationID, Transport: notification.PushTransport(input.Body.Transport), Token: token, Revision: input.Body.Revision})
		return deviceOutput(value, err)
	}, huma.OperationTags("notifications"), shared.SecuredOperation, func(op *huma.Operation) { op.DefaultStatus = 200 })
	huma.Get(api, path+"/by-installation/{installationId}", func(ctx context.Context, input *dto.GetDeviceInput) (*dto.DeviceOutput, error) {
		scope, err := authenticateScope(ctx, application, &input.ScopeInput)
		if err != nil {
			return nil, err
		}
		value, err := application.Notifications().GetDeviceByInstallation(ctx, scope, input.InstallationID)
		return deviceOutput(value, err)
	}, huma.OperationTags("notifications"), shared.SecuredOperation)
	huma.Delete(api, path+"/{deviceId}", func(ctx context.Context, input *dto.RevokeDeviceInput) (*dto.DeviceOutput, error) {
		scope, err := authenticateScope(ctx, application, &input.ScopeInput)
		if err != nil {
			return nil, err
		}
		value, err := application.Notifications().RevokeDevice(ctx, scope, input.DeviceID, input.Revision)
		return deviceOutput(value, err)
	}, huma.OperationTags("notifications"), shared.SecuredOperation, func(op *huma.Operation) { op.DefaultStatus = 200 })
}
func deviceOutput(value ports.NotificationDevice, err error) (*dto.DeviceOutput, error) {
	if err != nil {
		return nil, shared.ToHumaError(err)
	}
	return &dto.DeviceOutput{Body: shared.SuccessEnvelope[dto.DeviceResponse]{Data: mapper.DeviceToResponse(value), Meta: shared.Meta{TenantID: value.Scope.TenantID.String()}}}, nil
}
