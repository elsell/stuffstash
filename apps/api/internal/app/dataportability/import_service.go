package dataportability

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ImportDependencies struct {
	Targets                    ports.ImportTargets
	AssetTags                  ports.AssetTagRepository
	Assets                     ports.AssetRepository
	Attachments                ports.AttachmentRepository
	Audit                      ports.AuditRepository
	Blobs                      ports.BlobStorage
	Clock                      ports.Clock
	CustomFields               ports.CustomFieldDefinitionRepository
	IDs                        ports.IDGenerator
	ImportAssetUnitOfWork      ports.ImportAssetUnitOfWork
	ImportAttachmentSources    ports.ImportAttachmentSource
	ImportAttachmentUnitOfWork ports.ImportAttachmentUnitOfWork
	ImportJobTimeout           time.Duration
	ImportJobs                 ports.ImportJobRepository
	ImportLinks                ports.ImportLinkRepository
	ImportSourceVault          ports.ImportJobSourceVault
	ImportSources              ports.ImportSourceReader
	ImportWorker               ports.ImportWorker
	MaxAttachmentBytes         int
	Observer                   ports.Observer
}
type ImportService struct{ deps ImportDependencies }

func NewImportService(deps ImportDependencies) ImportService { return ImportService{deps: deps} }
