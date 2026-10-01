package identityaccess

import (
	"context"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a Service) ResolveUsersByID(ctx context.Context, ids []identity.PrincipalID) map[identity.PrincipalID]identity.User {
	if a.Users == nil || len(ids) == 0 {
		return map[identity.PrincipalID]identity.User{}
	}
	deduped := make([]identity.PrincipalID, 0, len(ids))
	seen := map[identity.PrincipalID]struct{}{}
	for _, id := range ids {
		if id.String() == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		deduped = append(deduped, id)
	}
	if len(deduped) == 0 {
		return map[identity.PrincipalID]identity.User{}
	}
	users, err := a.Users.UsersByID(ctx, deduped)
	if err != nil {
		a.Observer.Record(ctx, ports.Event{
			Name:    ports.EventPrincipalResolutionFailed,
			Message: "principal resolution failed",
			Fields: map[string]string{
				"count": strconv.Itoa(len(deduped)),
				"error": err.Error(),
			},
		})
		return map[identity.PrincipalID]identity.User{}
	}
	return users
}

type Service struct {
	Authenticator ports.Authenticator
	Users         ports.UserRepository
	Observer      ports.Observer
}

func (a Service) Authenticate(ctx context.Context, authorizationHeader string) (identity.Principal, error) {
	principal, err := a.Authenticator.Authenticate(ctx, authorizationHeader)
	if err != nil {
		a.Observer.Record(ctx, ports.Event{
			Name:    ports.EventAuthenticationFailed,
			Message: "authentication failed",
		})
		return identity.Principal{}, err
	}
	if err := a.SaveAuthenticatedUser(ctx, principal); err != nil {
		return identity.Principal{}, err
	}

	return principal, nil
}

func (a Service) SaveAuthenticatedUser(ctx context.Context, principal identity.Principal) error {
	if a.Users == nil {
		return nil
	}
	user, ok := identity.NewUser(principal.ID, principal.Email)
	if !ok {
		return nil
	}
	return a.Users.SaveUser(ctx, user)
}
