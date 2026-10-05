package ports

import (
	"context"
	"io"
)

type ArchiveTransfers interface {
	// The upload caller owns and closes the input stream.
	UploadArchive(context.Context, string, string, io.Reader) (Result[ArchiveJob], error)
	DownloadArchive(context.Context, Scope, string) (BinaryContent, error)
}
