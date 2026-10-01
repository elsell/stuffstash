package app

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/app/dataportability"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const MaxImportCSVBytes = dataportability.MaxImportCSVBytes

type ImportSourceInput = dataportability.ImportSourceInput
type CreateImportJobPreviewInput = dataportability.CreateImportJobPreviewInput
type StartImportJobInput = dataportability.StartImportJobInput
type GetImportJobInput = dataportability.GetImportJobInput
type ListImportJobsInput = dataportability.ListImportJobsInput
type CancelImportJobInput = dataportability.CancelImportJobInput
type RemoveImportJobFromHistoryInput = dataportability.RemoveImportJobFromHistoryInput
type ImportResult = dataportability.ImportResult

func (a App) importService() dataportability.ImportService {
	return dataportability.NewImportService(dataportability.ImportDependencies{
		Targets:                    importTargets{app: a},
		AssetTags:                  a.assetTags,
		Assets:                     a.assets,
		Attachments:                a.attachments,
		Audit:                      a.audit,
		Blobs:                      a.blobs,
		Clock:                      a.clock,
		CustomFields:               a.customFields,
		IDs:                        a.ids,
		ImportAssetUnitOfWork:      a.importAssetUnitOfWork,
		ImportAttachmentSources:    a.importAttachmentSources,
		ImportAttachmentUnitOfWork: a.importAttachmentUnitOfWork,
		ImportJobTimeout:           a.importJobTimeout,
		ImportJobs:                 a.importJobs,
		ImportLinks:                a.importLinks,
		ImportSourceVault:          a.importSourceVault,
		ImportSources:              a.importSources,
		ImportWorker:               a.importWorker,
		MaxAttachmentBytes:         a.maxAttachmentBytes,
		Observer:                   a.observer,
	})
}
func (a App) CreateImportJobPreview(ctx context.Context, input CreateImportJobPreviewInput) (importjob.Record, error) {
	return a.importService().CreateImportJobPreview(ctx, input)
}
func (a App) ListImportJobs(ctx context.Context, input ListImportJobsInput) ([]importjob.Record, error) {
	return a.importService().ListImportJobs(ctx, input)
}
func (a App) GetImportJob(ctx context.Context, input GetImportJobInput) (importjob.Record, error) {
	return a.importService().GetImportJob(ctx, input)
}
func (a App) GetImportJobForWorker(ctx context.Context, command ports.ImportJobCommand) (importjob.Record, error) {
	return a.importService().GetImportJobForWorker(ctx, command)
}
func (a App) StartImportJob(ctx context.Context, input StartImportJobInput) (importjob.Record, error) {
	return a.importService().StartImportJob(ctx, input)
}
func (a App) CancelImportJob(ctx context.Context, input CancelImportJobInput) (importjob.Record, error) {
	return a.importService().CancelImportJob(ctx, input)
}
func (a App) RemoveImportJobFromHistory(ctx context.Context, input RemoveImportJobFromHistoryInput) error {
	return a.importService().RemoveImportJobFromHistory(ctx, input)
}
func (a App) ExecuteImportJob(ctx context.Context, command ports.ImportJobCommand) (importjob.Record, error) {
	return a.importService().ExecuteImportJob(ctx, command)
}
func (a App) ResumeRunningImportJobs(ctx context.Context, limit int) (int, error) {
	return a.importService().ResumeRunningImportJobs(ctx, limit)
}
func (a App) VacuumImportJobCredentials(ctx context.Context) (int, error) {
	return a.importService().VacuumImportJobCredentials(ctx)
}
