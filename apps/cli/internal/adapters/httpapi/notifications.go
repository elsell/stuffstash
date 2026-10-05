package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func notification(v generated.NotificationResponse) ports.Notification {
	var trail []ports.NotificationAncestor
	if v.ParentTrail.GetOrEmpty() != nil {
		trail = make([]ports.NotificationAncestor, 0, len(v.ParentTrail.GetOrEmpty()))
	}
	for _, a := range v.ParentTrail.GetOrEmpty() {
		trail = append(trail, ports.NotificationAncestor{AssetID: a.AssetId, Title: a.Title, Kind: string(a.Kind)})
	}
	return ports.Notification{ID: v.Id, AssetID: v.AssetId, Title: v.Title, ParentAssetID: v.ParentAssetId, CustomAssetTypeID: v.CustomAssetTypeId, ExpirationDate: v.ExpirationDate, ExpirationPrecision: string(v.ExpirationPrecision), Milestone: string(v.Milestone), CreatedAt: v.CreatedAt, ReadAt: v.ReadAt, ParentTrail: trail, ParentTrailIncomplete: v.ParentTrailIncomplete}
}
func (c *Client) Notifications(ctx context.Context, s ports.Scope, p ports.Page, unread bool) (ports.Result[[]ports.Notification], error) {
	r, err := read[generated.SuccessEnvelopeListNotificationResponse](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdNotifications(ctx, s.Tenant, s.Inventory, &generated.GetTenantsByTenantIdInventoriesByInventoryIdNotificationsParams{Limit: &p.Limit, Cursor: &p.Cursor, UnreadOnly: &unread}))
	if err != nil {
		return ports.Result[[]ports.Notification]{}, err
	}
	var items []ports.Notification
	if r.Data.GetOrEmpty() != nil {
		items = make([]ports.Notification, 0, len(r.Data.GetOrEmpty()))
	}
	for _, v := range r.Data.GetOrEmpty() {
		items = append(items, notification(v))
	}
	return ports.Result[[]ports.Notification]{Data: items, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
func (c *Client) Notification(ctx context.Context, s ports.Scope, id string) (ports.Result[ports.Notification], error) {
	r, err := read[generated.SuccessEnvelopeNotificationResponse](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdNotificationsByNotificationId(ctx, s.Tenant, s.Inventory, id, nil))
	if err != nil {
		return ports.Result[ports.Notification]{}, err
	}
	return ports.Result[ports.Notification]{Data: notification(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) NotificationUnreadCount(ctx context.Context, s ports.Scope, cursor string) (ports.Result[ports.NotificationCount], error) {
	r, err := read[generated.SuccessEnvelopeUnreadCountResponse](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdNotificationsUnreadCount(ctx, s.Tenant, s.Inventory, &generated.GetTenantsByTenantIdInventoriesByInventoryIdNotificationsUnreadCountParams{Cursor: &cursor}))
	if err != nil {
		return ports.Result[ports.NotificationCount]{}, err
	}
	return ports.Result[ports.NotificationCount]{Data: ports.NotificationCount{Count: r.Data.Count}, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
func (c *Client) SetNotificationRead(ctx context.Context, s ports.Scope, id string, value bool) (ports.Result[ports.NotificationRead], error) {
	var response *http.Response
	var err error
	if value {
		response, err = c.sdk.PutTenantsByTenantIdInventoriesByInventoryIdNotificationsByNotificationIdRead(ctx, s.Tenant, s.Inventory, id, nil)
	} else {
		response, err = c.sdk.DeleteTenantsByTenantIdInventoriesByInventoryIdNotificationsByNotificationIdRead(ctx, s.Tenant, s.Inventory, id, nil)
	}
	r, err := read[generated.SuccessEnvelopeNotificationReadResponse](response, err)
	if err != nil {
		return ports.Result[ports.NotificationRead]{}, err
	}
	return ports.Result[ports.NotificationRead]{Data: ports.NotificationRead{ID: r.Data.Id, Read: r.Data.Read}, Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) ReadAllNotifications(ctx context.Context, s ports.Scope, cursor string) (ports.Result[ports.NotificationReadAll], error) {
	r, err := read[generated.SuccessEnvelopeInboxReadAllResponse](c.sdk.PutTenantsByTenantIdInventoriesByInventoryIdNotificationsReadAll(ctx, s.Tenant, s.Inventory, &generated.PutTenantsByTenantIdInventoriesByInventoryIdNotificationsReadAllParams{Cursor: &cursor}))
	if err != nil {
		return ports.Result[ports.NotificationReadAll]{}, err
	}
	return ports.Result[ports.NotificationReadAll]{Data: ports.NotificationReadAll{Complete: r.Data.Complete}, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
