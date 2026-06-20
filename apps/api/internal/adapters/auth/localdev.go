package auth

import (
	"context"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const localDevPrefix = "Bearer dev:"

type LocalDevAuthenticator struct{}

func NewLocalDevAuthenticator() LocalDevAuthenticator {
	return LocalDevAuthenticator{}
}

func (LocalDevAuthenticator) Authenticate(_ context.Context, authorizationHeader string) (identity.Principal, error) {
	if !strings.HasPrefix(authorizationHeader, localDevPrefix) {
		return identity.Principal{}, ports.ErrUnauthenticated
	}

	token := strings.TrimPrefix(authorizationHeader, localDevPrefix)
	parts := strings.SplitN(token, ":", 2)
	principalID, ok := identity.NewPrincipalID(parts[0])
	if !ok {
		return identity.Principal{}, ports.ErrUnauthenticated
	}

	principal := identity.Principal{ID: principalID}
	if len(parts) == 2 {
		email, ok := identity.NewEmail(parts[1])
		if !ok {
			return identity.Principal{}, ports.ErrUnauthenticated
		}
		principal.Email = email
	}

	return principal, nil
}
