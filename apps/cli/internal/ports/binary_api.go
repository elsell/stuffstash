package ports

import "context"

type BinaryAPI interface {
	AttachmentDownloads
	LabelContent(context.Context, Scope, string) (BinaryContent, error)
	ExportInventory(context.Context, Scope, string) (BinaryContent, error)
}
