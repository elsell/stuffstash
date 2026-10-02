package inventoryarchive

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"io"
)

type PackageCodec struct{}

var _ ports.ArchivePackageWriter = PackageCodec{}

func (PackageCodec) WritePackage(ctx context.Context, w io.Writer, in ports.ArchivePackageInput, source ports.ArchiveContentSource, limits ports.ArchivePackageLimits) error {
	media := make([]Media, len(in.Media))
	for i, m := range in.Media {
		media[i] = Media{SHA256: m.SHA256, SizeBytes: m.SizeBytes}
	}
	return Write(ctx, w, in.Metadata, Selection{Photos: in.Photos, OtherFiles: in.OtherFiles}, in.ExportedAt, media, source, Limits{CompressedBytes: limits.CompressedBytes, ExpandedBytes: limits.ExpandedBytes, MetadataBytes: limits.MetadataBytes, EntryBytes: limits.EntryBytes, Entries: limits.Entries})
}

func (PackageCodec) ReadPackage(ctx context.Context, r io.ReaderAt, size int64, limits ports.ArchivePackageLimits) (ports.ArchivePackageContents, error) {
	a, err := Read(ctx, r, size, Limits{CompressedBytes: limits.CompressedBytes, ExpandedBytes: limits.ExpandedBytes, MetadataBytes: limits.MetadataBytes, EntryBytes: limits.EntryBytes, Entries: limits.Entries})
	if err != nil {
		return ports.ArchivePackageContents{}, err
	}
	result := ports.ArchivePackageContents{Metadata: a.InventoryJSON, Photos: a.Selection.Photos, OtherFiles: a.Selection.OtherFiles, ExportedAt: a.ExportedAt, Source: packageSource{a}}
	for _, m := range a.Media {
		result.Media = append(result.Media, ports.ArchivePackageMedia{SHA256: m.SHA256, SizeBytes: m.SizeBytes})
	}
	return result, nil
}

type packageSource struct{ archive *Archive }

func (s packageSource) Open(ctx context.Context, hash string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.archive.Open(hash)
}
