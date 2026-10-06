package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func notificationDeviceResult(response *http.Response, err error) (ports.Result[ports.NotificationDevice], error) {
	r, err := read[generated.SuccessEnvelopeDeviceResponse](response, err)
	if err != nil {
		return ports.Result[ports.NotificationDevice]{}, err
	}
	v := r.Data
	return ports.Result[ports.NotificationDevice]{Schema: r.Schema, Meta: metadata(r.Meta), Data: ports.NotificationDevice{ID: v.Id, InstallationID: v.InstallationId, Transport: v.Transport, Revision: v.Revision, Active: v.Active}}, nil
}
func (c *Client) NotificationDevice(ctx context.Context, s ports.Scope, installation string) (ports.Result[ports.NotificationDevice], error) {
	return notificationDeviceResult(c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdNotificationDevicesByInstallationByInstallationId(ctx, s.Tenant, s.Inventory, installation, nil))
}
func (c *Client) RegisterNotificationDevice(ctx context.Context, s ports.Scope, body []byte) (ports.Result[ports.NotificationDevice], error) {
	return notificationDeviceResult(c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdNotificationDevicesWithBody(ctx, s.Tenant, s.Inventory, nil, "application/json", bytes.NewReader(body)))
}
func (c *Client) RemoveNotificationDevice(ctx context.Context, s ports.Scope, id string, revision int64) (ports.Result[ports.NotificationDevice], error) {
	return notificationDeviceResult(c.sdk.DeleteTenantsByTenantIdInventoriesByInventoryIdNotificationDevicesByDeviceId(ctx, s.Tenant, s.Inventory, id, &generated.DeleteTenantsByTenantIdInventoriesByInventoryIdNotificationDevicesByDeviceIdParams{Revision: revision}))
}
