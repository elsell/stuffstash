package routes

import (
	"context"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/archives/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/archives/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/app/dataportability"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

const jobsPath = "/tenants/{tenantId}/archive-jobs"

func Register(api huma.API, application app.App, service *dataportability.ArchiveService, timeout time.Duration) {
	registerTransfers(api, application, service, timeout)
	huma.Post(api, jobsPath, func(ctx context.Context, in *dto.CreateInput) (*dto.JobOutput, error) {
		access, err := authenticate(ctx, application, service, dto.Access{Authorization: in.Authorization, TenantID: in.TenantID, InventoryID: in.Body.InventoryID})
		if err != nil {
			return nil, err
		}
		job, err := service.CreateExport(ctx, access, in.IdempotencyKey, in.Body.Photos, in.Body.OtherFiles)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.JobOutput{Body: shared.SuccessEnvelope[dto.Job]{Data: mapper.Job(job), Meta: shared.Meta{TenantID: in.TenantID}}}, nil
	}, huma.OperationTags("archives"), shared.SecuredOperation, shared.CreatedOperation)
	huma.Get(api, jobsPath, func(ctx context.Context, in *dto.ListInput) (*dto.ListOutput, error) {
		access, err := authenticate(ctx, application, service, in.Access)
		if err != nil {
			return nil, err
		}
		jobs, err := service.List(ctx, access, in.After, in.Limit+1)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		hasMore := len(jobs) > in.Limit
		if hasMore {
			jobs = jobs[:in.Limit]
		}
		items := make([]dto.Job, 0, len(jobs))
		for _, job := range jobs {
			items = append(items, mapper.Job(job))
		}
		var cursor *string
		if hasMore {
			last := jobs[len(jobs)-1].ID
			cursor = &last
		}
		return &dto.ListOutput{Body: shared.SuccessEnvelope[[]dto.Job]{Data: items, Meta: shared.PaginatedMeta(in.TenantID, in.Limit, cursor, hasMore)}}, nil
	}, huma.OperationTags("archives"), shared.SecuredOperation)
	huma.Get(api, jobsPath+"/{jobId}", func(ctx context.Context, in *dto.JobInput) (*dto.JobOutput, error) {
		access, err := authenticate(ctx, application, service, in.Access)
		if err != nil {
			return nil, err
		}
		job, err := service.Job(ctx, access, in.JobID)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.JobOutput{Body: shared.SuccessEnvelope[dto.Job]{Data: mapper.Job(job), Meta: shared.Meta{TenantID: in.TenantID}}}, nil
	}, huma.OperationTags("archives"), shared.SecuredOperation)
	huma.Delete(api, jobsPath+"/{jobId}", func(ctx context.Context, in *dto.JobInput) (*dto.JobOutput, error) {
		access, err := authenticate(ctx, application, service, in.Access)
		if err != nil {
			return nil, err
		}
		job, err := service.Cancel(ctx, access, in.JobID)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.JobOutput{Body: shared.SuccessEnvelope[dto.Job]{Data: mapper.Job(job), Meta: shared.Meta{TenantID: in.TenantID}}}, nil
	}, huma.OperationTags("archives"), shared.SecuredOperation)
	huma.Post(api, jobsPath+"/{jobId}/retry", func(ctx context.Context, in *dto.JobInput) (*dto.JobOutput, error) {
		access, err := authenticate(ctx, application, service, in.Access)
		if err != nil {
			return nil, err
		}
		job, err := service.Retry(ctx, access, in.JobID)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.JobOutput{Body: shared.SuccessEnvelope[dto.Job]{Data: mapper.Job(job), Meta: shared.Meta{TenantID: in.TenantID}}}, nil
	}, huma.OperationTags("archives"), shared.SecuredOperation)
	huma.Post(api, jobsPath+"/{jobId}/approve", func(ctx context.Context, in *dto.ApproveInput) (*dto.JobOutput, error) {
		access, err := authenticate(ctx, application, service, in.Access)
		if err != nil {
			return nil, err
		}
		job, err := service.Approve(ctx, access, in.JobID, in.Body.Name)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.JobOutput{Body: shared.SuccessEnvelope[dto.Job]{Data: mapper.Job(job), Meta: shared.Meta{TenantID: in.TenantID}}}, nil
	}, huma.OperationTags("archives"), shared.SecuredOperation)
	huma.Get(api, jobsPath+"/{jobId}/preview", func(ctx context.Context, in *dto.JobInput) (*dto.PreviewOutput, error) {
		access, err := authenticate(ctx, application, service, in.Access)
		if err != nil {
			return nil, err
		}
		preview, err := service.Preview(ctx, access, in.JobID)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.PreviewOutput{Body: shared.SuccessEnvelope[dto.Preview]{Data: mapper.Preview(preview), Meta: shared.Meta{TenantID: in.TenantID}}}, nil
	}, huma.OperationTags("archives"), shared.SecuredOperation)
}
func authenticate(ctx context.Context, application app.App, service *dataportability.ArchiveService, in dto.Access) (dataportability.ArchiveAccess, error) {
	principal, err := shared.Authenticate(ctx, application, in.Authorization)
	if err != nil {
		return dataportability.ArchiveAccess{}, err
	}
	if service == nil {
		return dataportability.ArchiveAccess{}, huma.Error503ServiceUnavailable("Archive jobs require persistent storage.")
	}
	return dataportability.ArchiveAccess{Principal: principal, TenantID: tenant.ID(in.TenantID), InventoryID: inventory.InventoryID(in.InventoryID)}, nil
}
