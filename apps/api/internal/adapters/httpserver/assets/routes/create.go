package routes

import (
	"context"
	"errors"
	printingapp "github.com/stuffstash/stuff-stash/internal/app/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"net/http"
	"reflect"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterCreate(api huma.API, application app.App) {
	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets", func(ctx context.Context, input *dto.CreateAssetInput) (*dto.CreateAssetOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		command := app.CreateAssetInput{
			Expiration:        expirationInput(input.Body.Expiration),
			Principal:         principal,
			Source:            audit.SourceAPI,
			RequestID:         input.RequestID,
			TenantID:          tenant.ID(input.TenantID),
			InventoryID:       inventory.InventoryID(input.InventoryID),
			Kind:              input.Body.Kind,
			Title:             input.Body.Title,
			Description:       input.Body.Description,
			ParentAssetID:     input.Body.ParentAssetID,
			CustomAssetTypeID: input.Body.CustomAssetTypeID,
			CustomFields:      input.Body.CustomFields,
			TagIDs:            input.Body.TagIDs,
		}
		var result app.AssetMutationResult
		status := http.StatusCreated
		printJobID := ""
		if selection := input.Body.PrintLabel; selection != nil {
			created, createErr := application.CreateAssetAndPrint(ctx, command, printingapp.JobSelection{PrinterID: printing.PrinterID(selection.PrinterID), ExpectedMediaFingerprint: selection.ExpectedMediaFingerprint, Template: printing.TemplateSelection{ID: printing.TemplateID(selection.TemplateID), Version: selection.TemplateVersion, Options: printing.TemplateOptions{ShowReference: selection.TemplateOptions.ShowReference}}, Copies: selection.Copies}, input.IdempotencyKey)
			err = createErr
			result = app.AssetMutationResult{Asset: created.Asset, UndoableOperationID: created.Job.AssetCreationOperationID}
			printJobID = string(created.Job.ID)
			if !created.Created {
				status = http.StatusOK
			}
		} else {
			result, err = application.CreateAssetWithOperation(ctx, command)
		}
		if err != nil {
			if errors.Is(err, printingapp.ErrLabelsUnavailable) {
				return nil, huma.Error503ServiceUnavailable("Labels are not configured.")
			}
			return nil, shared.ToHumaError(err)
		}
		item := result.Asset
		tags, err := application.GetAssetAssignedTags(ctx, app.GetAssetAssignedTagsInput{
			Principal:   principal,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			AssetID:     item.ID,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		response := mapper.AssetToResponseWithTags(item, tags, nil, nil, nil)
		response.UndoableOperationID = result.UndoableOperationID
		response.PrintJobID = printJobID
		return &dto.CreateAssetOutput{Status: status,
			Body: shared.SuccessEnvelope[dto.AssetResponse]{
				Data: response,
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("assets"), shared.CreatedOperation, shared.SecuredOperation, func(op *huma.Operation) {
		op.Responses = map[string]*huma.Response{"200": {Description: "Existing asset and print job returned for an identical idempotent retry", Content: map[string]*huma.MediaType{"application/json": {Schema: api.OpenAPI().Components.Schemas.Schema(reflect.TypeFor[shared.SuccessEnvelope[dto.AssetResponse]](), true, "AssetResponseEnvelope")}}}}
	})
}
