package auth

import (
	"context"
	"strings"
	"unicode"

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

	principalID, ok := identity.NewPrincipalID(strings.TrimPrefix(authorizationHeader, localDevPrefix))
	if !ok || !validPrincipalID(principalID) {
		return identity.Principal{}, ports.ErrUnauthenticated
	}

	return identity.Principal{ID: principalID}, nil
}

func validPrincipalID(id identity.PrincipalID) bool {
	for _, r := range id.String() {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			continue
		}
		return false
	}

	return true
}
