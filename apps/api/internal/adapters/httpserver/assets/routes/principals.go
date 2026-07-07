package routes

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func resolveCheckoutPrincipals(ctx context.Context, application app.App, checkouts []asset.Checkout) map[identity.PrincipalID]identity.User {
	ids := make([]identity.PrincipalID, 0, len(checkouts))
	for _, checkout := range checkouts {
		ids = append(ids, identity.PrincipalID(checkout.CheckedOutByPrincipal))
	}
	return application.ResolveUsersByID(ctx, ids)
}

func resolveCheckoutPrincipalsFromMap(ctx context.Context, application app.App, checkouts map[asset.ID]asset.Checkout) map[identity.PrincipalID]identity.User {
	ids := make([]identity.PrincipalID, 0, len(checkouts))
	for _, checkout := range checkouts {
		ids = append(ids, identity.PrincipalID(checkout.CheckedOutByPrincipal))
	}
	return application.ResolveUsersByID(ctx, ids)
}

func resolveCheckedOutAssetPrincipals(ctx context.Context, application app.App, items []ports.CheckedOutAsset) map[identity.PrincipalID]identity.User {
	ids := make([]identity.PrincipalID, 0, len(items))
	for _, item := range items {
		ids = append(ids, identity.PrincipalID(item.Checkout.CheckedOutByPrincipal))
	}
	return application.ResolveUsersByID(ctx, ids)
}
