package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
)

type ArchiveRestorePublication struct {
	Job               archivejob.Record
	Plan              ArchiveRestorePlan
	OwnerGrantEventID string
	AuditRecords      []audit.Record
}
type ArchiveRestoreUnitOfWork interface {
	PublishArchiveRestore(context.Context, ArchiveRestorePublication) (archivejob.Record, error)
	// Read the exact, scoped grant receipt; never write authorization on this path.
	ArchiveRestoreGrantProcessed(context.Context, archivejob.Record) (bool, error)
}
