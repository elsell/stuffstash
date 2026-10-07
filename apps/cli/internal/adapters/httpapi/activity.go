package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) AssetActivity(ctx context.Context, scope ports.Scope, asset string, view ports.ActivityView, p ports.Page) (ports.Result[[]ports.Activity], error) {
	params := &generated.GetTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdActivityParams{Limit: &p.Limit, Cursor: &p.Cursor}
	if view != "" {
		if view != ports.ActivityChanges && view != ports.ActivityAll {
			return ports.Result[[]ports.Activity]{}, ports.Failure("usage", "Select changes or all for --view.")
		}
		v := generated.GetTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdActivityParamsView(view)
		params.View = &v
	}
	response, err := c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdActivity(ctx, scope.Tenant, scope.Inventory, asset, params)
	r, err := read[generated.SuccessEnvelopeListAssetActivityResponse](response, err)
	if err != nil {
		return ports.Result[[]ports.Activity]{}, err
	}
	var events []ports.Activity
	if r.Data.GetOrEmpty() != nil {
		events = make([]ports.Activity, 0, len(r.Data.GetOrEmpty()))
	}
	for _, v := range r.Data.GetOrEmpty() {
		event := ports.Activity{ID: v.Id, Action: string(v.Action), Category: string(v.Category), OccurredAt: v.OccurredAt, PrincipalID: v.PrincipalId, RequestID: v.RequestId, Source: string(v.Source), TechnicalMetadata: v.TechnicalMetadata}
		if v.Principal != nil {
			event.Principal = &ports.AuditPrincipal{ID: v.Principal.Id, Email: v.Principal.Email}
		}
		if v.Undo != nil {
			event.Undo = &ports.ActivityUndo{OperationID: v.Undo.OperationId, Status: string(v.Undo.Status)}
		}
		if v.Changes.GetOrEmpty() != nil {
			event.Changes = make([]ports.ActivityChange, 0, len(v.Changes.GetOrEmpty()))
		}
		for _, change := range v.Changes.GetOrEmpty() {
			event.Changes = append(event.Changes, ports.ActivityChange{Field: string(change.Field), PreviousValue: change.PreviousValue, CurrentValue: change.CurrentValue})
		}
		events = append(events, event)
	}
	return ports.Result[[]ports.Activity]{Data: events, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
