package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/notifications/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
)

func registerInboxBatch(api huma.API, application app.App) {
	const path = "/tenants/{tenantId}/inventories/{inventoryId}/notifications"
	huma.Get(api, path+"/unread-count", func(ctx context.Context, input *dto.InboxBatchInput) (*dto.UnreadCountOutput, error) {
		scope, err := authenticateScope(ctx, application, &input.ScopeInput)
		if err != nil {
			return nil, err
		}
		page, err := application.Notifications().CountUnreadPage(ctx, scope, input.Cursor)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		var cursor *string
		if page.NextCursor != "" {
			cursor = &page.NextCursor
		}
		return &dto.UnreadCountOutput{Body: shared.SuccessEnvelope[dto.UnreadCountResponse]{Data: dto.UnreadCountResponse{Count: page.Count}, Meta: shared.PaginatedMeta(input.TenantID, 100, cursor, cursor != nil)}}, nil
	}, huma.OperationTags("notifications"), shared.SecuredOperation)
	huma.Put(api, path+"/read-all", func(ctx context.Context, input *dto.InboxBatchInput) (*dto.InboxReadAllOutput, error) {
		scope, err := authenticateScope(ctx, application, &input.ScopeInput)
		if err != nil {
			return nil, err
		}
		page, err := application.Notifications().MarkInboxPageRead(ctx, scope, input.Cursor)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		var cursor *string
		if page.NextCursor != "" {
			cursor = &page.NextCursor
		}
		return &dto.InboxReadAllOutput{Body: shared.SuccessEnvelope[dto.InboxReadAllResponse]{Data: dto.InboxReadAllResponse{Complete: cursor == nil}, Meta: shared.PaginatedMeta(input.TenantID, 100, cursor, cursor != nil)}}, nil
	}, huma.OperationTags("notifications"), shared.SecuredOperation)
}
