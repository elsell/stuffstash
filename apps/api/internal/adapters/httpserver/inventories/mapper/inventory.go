package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/inventories/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
)

func InventoryToResponse(item inventory.Inventory) dto.InventoryResponse {
	return dto.InventoryResponse{
		ID:             item.ID.String(),
		TenantID:       item.TenantID.String(),
		Name:           item.Name.String(),
		LifecycleState: item.LifecycleState.String(),
	}
}

func InventoriesToResponse(items []inventory.Inventory) []dto.InventoryResponse {
	data := make([]dto.InventoryResponse, 0, len(items))
	for _, item := range items {
		data = append(data, InventoryToResponse(item))
	}
	return data
}
