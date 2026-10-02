package ports

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
)

var ErrArchiveJobConflict = errors.New("archive job changed or request conflicts")
var ErrArchiveJobScope = errors.New("archive job scope is required")

type ArchiveJobScope struct{ TenantID, SourceInventoryID, PrincipalID string }

// Restore jobs have an empty SourceInventoryID and are tenant-owned. Export
// reads require the exact source inventory as well as the tenant.
type ArchiveJobRepository interface {
	CreateArchiveJob(context.Context, archivejob.Record) (archivejob.Record, error)
	ArchiveJobByID(context.Context, ArchiveJobScope, string) (archivejob.Record, bool, error)
	ListArchiveJobs(context.Context, ArchiveJobScope, string, int) ([]archivejob.Record, error)
	UpdateArchiveJob(context.Context, archivejob.Record, int64) (bool, error)
	// This operational queue is the only unscoped read, never a user listing.
	ListRunnableArchiveJobs(context.Context, time.Time, int) ([]archivejob.Record, error)
}
