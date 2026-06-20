package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/identity/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
)

func PrincipalToResponse(principal identity.Principal) dto.PrincipalResponse {
	return dto.PrincipalResponse{ID: principal.ID.String(), Email: principal.Email.String()}
}
