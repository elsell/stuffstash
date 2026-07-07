package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/imports/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/imports/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func Register(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/imports/jobs", func(ctx context.Context, input *dto.ImportJobListInput) (*dto.ImportJobListOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		jobs, err := application.ListImportJobs(ctx, app.ListImportJobsInput{
			Principal:   principal,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.ImportJobListOutput{
			Body: shared.SuccessEnvelope[dto.ImportJobListResponse]{
				Data: mapper.JobListToResponse(jobs, importJobUsers(ctx, application, jobs...)),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("imports"), shared.SecuredOperation)

	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/imports/jobs/preview", func(ctx context.Context, input *dto.ImportJobPreviewInput) (*dto.ImportJobOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		job, err := application.CreateImportJobPreview(ctx, app.CreateImportJobPreviewInput{
			Principal:   principal,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			Source:      sourceInput(input.Body),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.ImportJobOutput{
			Body: shared.SuccessEnvelope[dto.ImportJobResponse]{
				Data: mapper.JobToResponse(job, importJobUsers(ctx, application, job)),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("imports"), shared.SecuredOperation)

	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/imports/jobs/{jobId}", func(ctx context.Context, input *dto.ImportJobDetailInput) (*dto.ImportJobOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		job, err := application.GetImportJob(ctx, app.GetImportJobInput{
			Principal:   principal,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			JobID:       importjob.ID(input.JobID),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.ImportJobOutput{
			Body: shared.SuccessEnvelope[dto.ImportJobResponse]{
				Data: mapper.JobToResponse(job, importJobUsers(ctx, application, job)),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("imports"), shared.SecuredOperation)

	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/imports/jobs/{jobId}/start", func(ctx context.Context, input *dto.ImportJobStartInput) (*dto.ImportJobOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		job, err := application.StartImportJob(ctx, app.StartImportJobInput{
			Principal:   principal,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			JobID:       importjob.ID(input.JobID),
			Source:      sourceInput(input.Body),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.ImportJobOutput{
			Body: shared.SuccessEnvelope[dto.ImportJobResponse]{
				Data: mapper.JobToResponse(job, importJobUsers(ctx, application, job)),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("imports"), shared.SecuredOperation)

	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/imports/jobs/{jobId}/cancel", func(ctx context.Context, input *dto.ImportJobCancelInput) (*dto.ImportJobOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		job, err := application.CancelImportJob(ctx, app.CancelImportJobInput{
			Principal:   principal,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			JobID:       importjob.ID(input.JobID),
			Mode:        importjob.CancellationMode(input.Body.Mode),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.ImportJobOutput{
			Body: shared.SuccessEnvelope[dto.ImportJobResponse]{
				Data: mapper.JobToResponse(job, importJobUsers(ctx, application, job)),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("imports"), shared.SecuredOperation)

	huma.Delete(api, "/tenants/{tenantId}/inventories/{inventoryId}/imports/jobs/{jobId}", func(ctx context.Context, input *dto.RemoveImportJobInput) (*dto.RemoveImportJobOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		if err := application.RemoveImportJobFromHistory(ctx, app.RemoveImportJobFromHistoryInput{
			Principal:   principal,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			JobID:       importjob.ID(input.JobID),
		}); err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.RemoveImportJobOutput{}, nil
	}, huma.OperationTags("imports"), shared.NoContentOperation, shared.SecuredOperation)
}

func importJobUsers(ctx context.Context, application app.App, jobs ...importjob.Record) map[identity.PrincipalID]identity.User {
	ids := make([]identity.PrincipalID, 0, len(jobs))
	for _, job := range jobs {
		if job.ActorID.String() != "" {
			ids = append(ids, identity.PrincipalID(job.ActorID.String()))
		}
	}
	return application.ResolveUsersByID(ctx, ids)
}

func sourceInput(body dto.ImportSourceRequest) app.ImportSourceInput {
	return app.ImportSourceInput{
		SourceType:          body.SourceType,
		BaseURL:             body.BaseURL,
		Username:            body.Username,
		Password:            body.Password,
		IncludeImages:       body.IncludeImages,
		AllowInsecureTLS:    body.AllowInsecureTLS,
		AllowPrivateNetwork: body.AllowPrivateNetwork,
		FileName:            body.FileName,
		ContentBase64:       body.ContentBase64,
	}
}
