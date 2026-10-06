package ports

import "context"

type AttachmentAction string

const (
	ArchiveAttachment AttachmentAction = "archive"
	RestoreAttachment AttachmentAction = "restore"
)

type Attachment struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenantId"`
	InventoryID string `json:"inventoryId"`
	AssetID     string `json:"assetId"`
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
	SHA256      string `json:"sha256"`
	CreatedAt   string `json:"createdAt"`
	Lifecycle   string `json:"lifecycleState"`
}
type AttachmentsAPI interface {
	Attachments(context.Context, Scope, string, Page) (Result[[]Attachment], error)
	Attachment(context.Context, Scope, string, string) (Result[Attachment], error)
	ChangeAttachment(context.Context, Scope, string, string, AttachmentAction) (Result[Attachment], error)
	DeleteAttachment(context.Context, Scope, string, string) error
}
