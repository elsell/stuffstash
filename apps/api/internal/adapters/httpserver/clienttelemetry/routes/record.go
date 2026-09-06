package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/clienttelemetry/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/clienttelemetry/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
)

func Register(api huma.API, application app.App) {
	huma.Post(api, "/client-telemetry", func(ctx context.Context, input *dto.RecordInput) (*dto.RecordOutput, error) {
		if _, err := shared.Authenticate(ctx, application, input.Authorization); err != nil {
			return nil, err
		}
		if err := application.RecordClientTelemetry(ctx, mapper.Measurements(input.Body.Measurements)); err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.RecordOutput{Body: shared.SuccessEnvelope[dto.Accepted]{Data: dto.Accepted{Count: len(input.Body.Measurements)}}}, nil
	}, huma.OperationTags("telemetry"), shared.SecuredOperation)
}
