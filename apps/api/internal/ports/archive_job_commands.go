package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
)

// User mutations persist their audit record in the same transaction as the job.
type ArchiveJobCommands interface {
	CreateArchiveJobAudited(context.Context, archivejob.Record, audit.Record) (archivejob.Record, error)
	UpdateArchiveJobAudited(context.Context, archivejob.Record, int64, audit.Record) (bool, error)
}
