package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/identity/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/identity/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
)

func Register(api huma.API, application app.App) {
	huma.Get(api, "/me", func(ctx context.Context, input *dto.MeInput) (*dto.MeOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		return &dto.MeOutput{
			Body: shared.SuccessEnvelope[dto.PrincipalResponse]{
				Data: mapper.PrincipalToResponse(principal),
				Meta: shared.Meta{},
			},
		}, nil
	}, huma.OperationTags("identity"), shared.SecuredOperation)
}
