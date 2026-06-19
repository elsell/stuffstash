package shared

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
)

func Authenticate(ctx context.Context, application app.App, authorization string) (identity.Principal, error) {
	principal, err := application.Authenticate(ctx, authorization)
	if err != nil {
		return identity.Principal{}, ToHumaError(err)
	}
	return principal, nil
}
