package routes

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func resolveSearchCheckoutPrincipals(ctx context.Context, application app.App, results []ports.AssetSearchResult) map[identity.PrincipalID]identity.User {
	ids := make([]identity.PrincipalID, 0, len(results))
	for _, result := range results {
		if result.CurrentCheckout == nil {
			continue
		}
		ids = append(ids, identity.PrincipalID(result.CurrentCheckout.CheckedOutByPrincipal))
	}
	return application.ResolveUsersByID(ctx, ids)
}
