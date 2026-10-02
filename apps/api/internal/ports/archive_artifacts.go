package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"time"
)

type ArchiveArtifactKind string

const (
	ArchiveArtifactSource        ArchiveArtifactKind = "source"
	ArchiveArtifactPlan          ArchiveArtifactKind = "plan"
	ArchiveArtifactExport        ArchiveArtifactKind = "export"
	ArchiveArtifactRestoredMedia ArchiveArtifactKind = "restored_media"
)

type ArchiveArtifact struct {
	Key                                media.StorageKey
	JobID, TenantID, SourceInventoryID string
	Kind                               ArchiveArtifactKind
	ExpiresAt                          time.Time
}
type ArchiveArtifactRepository interface {
	RegisterArchiveArtifact(context.Context, archivejob.Record, media.StorageKey, ArchiveArtifactKind, time.Time) error
	ListDueArchiveArtifacts(context.Context, time.Time, int) ([]ArchiveArtifact, error)
	RetireArchiveArtifact(context.Context, ArchiveArtifact, time.Time, string) error
	ListExpiredArchiveJobs(context.Context, time.Time, int) ([]archivejob.Record, error)
}
