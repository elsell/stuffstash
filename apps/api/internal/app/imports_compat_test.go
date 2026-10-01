package app

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/app/dataportability"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const maxImportRequestIDLength = dataportability.MaxImportRequestIDLength

type importSourceIdentity = dataportability.ImportSourceIdentity
type importImportedResourceInput = dataportability.ImportedResourceInput
type importCancelledError = dataportability.ImportCancelledError

func importAttachmentSessionFailureMessage(err error) importplan.Message {
	return dataportability.ImportAttachmentSessionFailureMessage(err)
}
func importAttachmentReadFailureMessage(err error, attachment importplan.Attachment) importplan.Message {
	return dataportability.ImportAttachmentReadFailureMessage(err, attachment)
}
func (a App) updateImportProgress(ctx context.Context, command ports.ImportJobCommand, phase importjob.Phase, done int, total int, message string) error {
	return a.importService().UpdateImportProgress(ctx, command, phase, done, total, message)
}
func sourceFingerprint(plan importplan.Plan) (string, error) {
	return dataportability.SourceFingerprint(plan)
}
func importAttachmentSourceLinkKey(tenantID tenant.ID, inventoryID inventory.InventoryID, sourceIdentity importSourceIdentity, planned importplan.Attachment) ports.ImportSourceLinkKey {
	return dataportability.ImportAttachmentSourceLinkKey(tenantID, inventoryID, sourceIdentity, planned)
}
func (a App) importedResourceRecords(input importImportedResourceInput) (ports.ImportSourceLink, ports.ImportJobResource, error) {
	return a.importService().ImportedResourceRecords(input)
}
func (a App) importSourceRequest(input ImportSourceInput) (ports.ImportSourceRequest, error) {
	return a.importService().ImportSourceRequest(input)
}
func importSourceInputError(err error) error { return dataportability.ImportSourceInputError(err) }
func (a App) normalizedImportPlanForJob(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, plan importplan.Plan) (importplan.Plan, error) {
	return a.importService().NormalizedImportPlanForJob(ctx, tenantID, inventoryID, plan)
}
