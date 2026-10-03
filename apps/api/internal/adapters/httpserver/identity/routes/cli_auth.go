package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/identity/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
)

func RegisterCLIAuth(api huma.API, metadata *dto.CLIAuthMetadata) {
	huma.Get(api, "/auth/cli/config", func(ctx context.Context, input *dto.CLIAuthConfigInput) (*dto.CLIAuthConfigOutput, error) {
		if metadata == nil {
			return nil, huma.Error503ServiceUnavailable("CLI sign-in is not configured.")
		}
		return &dto.CLIAuthConfigOutput{CacheControl: "no-store", Body: shared.SuccessEnvelope[dto.CLIAuthMetadata]{Data: *metadata, Meta: shared.Meta{}}}, nil
	}, func(op *huma.Operation) { op.OperationID = "get-cli-auth-config" }, huma.OperationTags("identity"))
}
