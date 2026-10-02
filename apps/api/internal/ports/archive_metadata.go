package ports

import "context"

// ArchiveMetadataCodec handles the separately versioned inventory document.
// Decoding never grants authority: the application must validate the graph and
// remap source identifiers before publishing any destination resources.
type ArchiveMetadataCodec interface {
	EncodeMetadata(context.Context, InventoryExportDocument, int) ([]byte, error)
	DecodeMetadata(context.Context, []byte, int) (InventoryExportDocument, error)
}
