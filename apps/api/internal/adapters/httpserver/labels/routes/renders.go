package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/labels/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/labels/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	labelapp "github.com/stuffstash/stuff-stash/internal/app/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

func registerRenders(api huma.API, application app.App) {
	huma.Post(api, scopePath+"/assets/{assetId}/label-renders", func(ctx context.Context, input *dto.RenderInput) (*dto.RenderOutput, error) {
		scope, err := authenticate(ctx, application, input.ScopeInput)
		if err != nil {
			return nil, err
		}
		rendered, err := application.Labels().Render(ctx, scope, asset.ID(input.AssetID), labelapp.LabelRenderInput{Template: mapper.Selection(input.Body.Template), Media: mapper.Media(input.Body.Media), Format: printing.Format(input.Body.Format)})
		if err != nil {
			return nil, labelError(err)
		}
		return &dto.RenderOutput{CacheControl: "private, no-store", Body: shared.SuccessEnvelope[dto.RenderResponse]{Data: mapper.Render(rendered), Meta: shared.Meta{TenantID: input.TenantID}}}, nil
	}, huma.OperationTags("labels"), shared.SecuredOperation, func(op *huma.Operation) { op.DefaultStatus = 201 })
	huma.Get(api, scopePath+"/label-renders/{renderId}/content", func(ctx context.Context, input *dto.ContentInput) (*dto.ContentOutput, error) {
		scope, err := authenticate(ctx, application, input.ScopeInput)
		if err != nil {
			return nil, err
		}
		rendered, err := application.Labels().Content(ctx, scope, printing.RenderID(input.RenderID))
		if err != nil {
			return nil, labelError(err)
		}
		extension := "png"
		if rendered.ContentType == "application/pdf" {
			extension = "pdf"
		}
		return &dto.ContentOutput{CacheControl: "private, no-store", ContentType: rendered.ContentType, ContentDisposition: "attachment; filename=\"label." + extension + "\"", Body: rendered.Content}, nil
	}, huma.OperationTags("labels"), shared.SecuredOperation)
	huma.Get(api, scopePath+"/label-templates", func(ctx context.Context, input *dto.ScopeInput) (*dto.TemplatesOutput, error) {
		scope, err := authenticate(ctx, application, *input)
		if err != nil {
			return nil, err
		}
		values, err := application.Labels().Templates(ctx, scope)
		if err != nil {
			return nil, labelError(err)
		}
		result := make([]dto.TemplateResponse, len(values))
		for i, value := range values {
			result[i] = mapper.Template(value)
		}
		return &dto.TemplatesOutput{Body: shared.SuccessEnvelope[[]dto.TemplateResponse]{Data: result, Meta: shared.Meta{TenantID: input.TenantID}}}, nil
	}, huma.OperationTags("labels"), shared.SecuredOperation)
}
