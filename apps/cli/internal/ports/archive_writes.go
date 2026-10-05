package ports

import "context"

type ArchiveWrites interface {
	CreateArchiveJob(context.Context, string, string, []byte) (Result[ArchiveJob], error)
	ApproveArchiveJob(context.Context, Scope, string, []byte) (Result[ArchiveJob], error)
}
