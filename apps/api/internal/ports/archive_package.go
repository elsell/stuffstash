package ports

import (
	"context"
	"errors"
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

// The supplied ReaderAt remains open and immutable while content is consumed.
type ArchivePackageReader interface {
	ReadPackage(context.Context, io.ReaderAt, int64, ArchivePackageLimits) (ArchivePackageContents, error)
}
type ArchivePackageContents struct {
	Metadata           []byte
	Photos, OtherFiles bool
	ExportedAt         time.Time
	Media              []ArchivePackageMedia
	Source             ArchiveContentSource
}
type ArchivePlanCodec interface {
	EncodePlan(context.Context, ArchiveRestorePlan, int) ([]byte, error)
	DecodePlan(context.Context, []byte, int) (ArchiveRestorePlan, error)
}

var ErrArchivePackageInvalid = errors.New("invalid inventory archive")
var ErrArchivePackageLimit = errors.New("inventory archive exceeds configured limit")
