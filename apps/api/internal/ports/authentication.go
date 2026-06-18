package ports

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
)

var ErrUnauthenticated = errors.New("unauthenticated")

type Authenticator interface {
	Authenticate(ctx context.Context, authorizationHeader string) (identity.Principal, error)
}
