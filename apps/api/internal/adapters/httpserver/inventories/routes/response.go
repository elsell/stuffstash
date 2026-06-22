package routes

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/inventories/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/inventories/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func inventoryResponse(ctx context.Context, application app.App, principal identity.Principal, item inventory.Inventory) (dto.InventoryResponse, error) {
	access, err := application.InventoryAccess(ctx, principal, tenant.ID(item.TenantID.String()), item.ID)
	if err != nil {
		return dto.InventoryResponse{}, err
	}
	response := mapper.InventoryToResponse(item)
	response.Access = shared.AccessToResponse(access)
	return response, nil
}

func inventoryResponses(ctx context.Context, application app.App, principal identity.Principal, items []inventory.Inventory) ([]dto.InventoryResponse, error) {
	responses := mapper.InventoriesToResponse(items)
	for index, item := range items {
		access, err := application.InventoryAccess(ctx, principal, tenant.ID(item.TenantID.String()), item.ID)
		if err != nil {
			return nil, err
		}
		responses[index].Access = shared.AccessToResponse(access)
	}
	return responses, nil
}
