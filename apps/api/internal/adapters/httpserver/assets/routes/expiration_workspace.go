package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	expirationapp "github.com/stuffstash/stuff-stash/internal/app/expiration"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func RegisterExpirationWorkspace(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/expiration-assets", func(ctx context.Context, input *dto.ExpirationWorkspaceInput) (*dto.ExpirationWorkspaceOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		result, err := application.ListExpirationAssets(ctx, expirationapp.WorkspaceInput{Principal: principal, TenantID: tenant.ID(input.TenantID), InventoryID: inventory.InventoryID(input.InventoryID), Source: audit.SourceAPI, RequestID: input.RequestID, Mode: expirationapp.WorkspaceMode(input.Mode), Filter: expirationapp.WorkspaceFilter{Kind: asset.Kind(input.Kind), CheckoutState: ports.AssetCheckoutStateFilter(input.CheckoutState), Text: input.Query, TypeID: input.TypeID, TagIDs: input.TagIDs, LocationID: input.LocationID, FromDate: input.FromDate, ThroughDate: input.ThroughDate}, Limit: input.Limit, Cursor: input.Cursor})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		presentation := mapper.ExpirationWorkspacePresentation{Soon: result.Counts.Soon, Expired: result.Counts.Expired, All: result.Counts.All, Timezone: result.Timezone}
		for _, item := range result.Items {
			value := item.Expiration
			presentation.Contexts = append(presentation.Contexts, mapper.ExpirationWorkspaceContext{State: value.State, TrackingEnabled: value.TrackingEnabled, AdvanceDays: value.AdvanceDays, Timezone: value.Timezone})
		}
		data := mapper.ExpirationWorkspaceToResponse(result.Records, result.PrimaryPhotos, presentation)
		return &dto.ExpirationWorkspaceOutput{Body: shared.SuccessEnvelope[dto.ExpirationWorkspaceData]{Data: data, Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.HasMore)}}, nil

	}, huma.OperationTags("assets"), shared.SecuredOperation)
}
