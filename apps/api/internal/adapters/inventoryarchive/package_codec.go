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
