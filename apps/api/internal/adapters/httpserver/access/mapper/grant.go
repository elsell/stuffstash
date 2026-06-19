package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/access/dto"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func GrantToResponse(grant ports.InventoryAccessGrant) dto.GrantResponse {
	return dto.GrantResponse{
		TenantID:     grant.TenantID.String(),
		InventoryID:  grant.InventoryID.String(),
		PrincipalID:  grant.PrincipalID.String(),
		Relationship: string(grant.Relationship),
	}
}

func GrantsToResponse(grants []ports.InventoryAccessGrant) []dto.GrantResponse {
	data := make([]dto.GrantResponse, 0, len(grants))
	for _, grant := range grants {
		data = append(data, GrantToResponse(grant))
	}
	return data
}
