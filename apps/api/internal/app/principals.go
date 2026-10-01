package app

import (
	"context"
	identityapp "github.com/stuffstash/stuff-stash/internal/app/identityaccess"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
)

func (a App) identityService() identityapp.Service {
	return identityapp.Service{Authenticator: a.auth, Users: a.users, Observer: a.observer}
}
func (a App) Authenticate(ctx context.Context, authorizationHeader string) (identity.Principal, error) {
	return a.identityService().Authenticate(ctx, authorizationHeader)
}
func (a App) saveAuthenticatedUser(ctx context.Context, principal identity.Principal) error {
	return a.identityService().SaveAuthenticatedUser(ctx, principal)
}
func (a App) ResolveUsersByID(ctx context.Context, ids []identity.PrincipalID) map[identity.PrincipalID]identity.User {
	return a.identityService().ResolveUsersByID(ctx, ids)
}
