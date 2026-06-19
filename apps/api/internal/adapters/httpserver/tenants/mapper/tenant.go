package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/tenants/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func TenantToResponse(item tenant.Tenant) dto.TenantResponse {
	return dto.TenantResponse{
		ID:   item.ID.String(),
		Name: item.Name.String(),
	}
}
