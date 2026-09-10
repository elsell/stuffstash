package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/notifications/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/notifications/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
)

func registerInbox(api huma.API, application app.App) {
	const path = "/tenants/{tenantId}/inventories/{inventoryId}/notifications"
	huma.Get(api, path, func(ctx context.Context, input *dto.InboxListInput) (*dto.InboxOutput, error) {
		scope, err := authenticateScope(ctx, application, &input.ScopeInput)
		if err != nil {
			return nil, err
		}
		page, err := application.Notifications().ListInbox(ctx, scope, input.Cursor, input.Limit, input.UnreadOnly)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		values := make([]dto.NotificationResponse, 0, len(page.Items))
		for _, view := range page.Items {
			values = append(values, mapper.NotificationToResponse(view.Notification, view.Asset))
		}
		var cursor *string
		if page.NextCursor != "" {
			cursor = &page.NextCursor
		}
		return &dto.InboxOutput{Body: shared.SuccessEnvelope[[]dto.NotificationResponse]{Data: values, Meta: shared.PaginatedMeta(input.TenantID, input.Limit, cursor, cursor != nil)}}, nil
	}, huma.OperationTags("notifications"), shared.SecuredOperation)
	huma.Get(api, path+"/{notificationId}", func(ctx context.Context, input *dto.NotificationInput) (*dto.NotificationOutput, error) {
		scope, err := authenticateScope(ctx, application, &input.ScopeInput)
		if err != nil {
			return nil, err
		}
		view, err := application.Notifications().GetNotification(ctx, scope, input.NotificationID)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.NotificationOutput{Body: shared.SuccessEnvelope[dto.NotificationResponse]{Data: mapper.NotificationToResponse(view.Notification, view.Asset), Meta: shared.Meta{TenantID: input.TenantID}}}, nil
	}, huma.OperationTags("notifications"), shared.SecuredOperation)
	huma.Put(api, path+"/{notificationId}/read", func(ctx context.Context, input *dto.NotificationInput) (*dto.NotificationReadOutput, error) {
		scope, err := authenticateScope(ctx, application, &input.ScopeInput)
		if err != nil {
			return nil, err
		}
		if err := application.Notifications().MarkRead(ctx, scope, input.NotificationID); err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.NotificationReadOutput{Body: shared.SuccessEnvelope[dto.NotificationReadResponse]{Data: dto.NotificationReadResponse{ID: input.NotificationID, Read: true}, Meta: shared.Meta{TenantID: input.TenantID}}}, nil
	}, huma.OperationTags("notifications"), shared.SecuredOperation)
}
