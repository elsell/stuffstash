package ports

import (
	"context"
	"io"
	"time"
)

// ArchivePackageWriter owns the portable container format, independent of jobs.
type ArchivePackageWriter interface {
	WritePackage(context.Context, io.Writer, ArchivePackageInput, ArchiveContentSource, ArchivePackageLimits) error
}
type ArchiveContentSource interface {
	Open(context.Context, string) (io.ReadCloser, error)
}
type ArchivePackageMedia struct {
	SHA256    string
	SizeBytes int64
}
type ArchivePackageLimits struct {
	CompressedBytes, ExpandedBytes, MetadataBytes, EntryBytes int64
	Entries                                                   int
}
type ArchivePackageInput struct {
	Metadata           []byte
	Photos, OtherFiles bool
	ExportedAt         time.Time
	Media              []ArchivePackageMedia
}
