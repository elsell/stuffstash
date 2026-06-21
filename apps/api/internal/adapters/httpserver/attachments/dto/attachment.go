package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type CreateAssetAttachmentInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	AssetID       string `path:"assetId" doc:"Asset ID"`
	Body          CreateAttachmentBody
}

type CreateAttachmentBody struct {
	FileName      string `json:"fileName" maxLength:"255" doc:"Original file name"`
	ContentType   string `json:"contentType" enum:"image/jpeg,image/png,image/webp,application/pdf" doc:"Media type"`
	ContentBase64 string `json:"contentBase64" doc:"Base64-encoded content"`
}

type CreateAssetAttachmentOutput struct {
	Body shared.SuccessEnvelope[AttachmentResponse]
}

type InitiateAssetAttachmentDirectUploadInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	AssetID       string `path:"assetId" doc:"Asset ID"`
	Body          InitiateDirectUploadBody
}

type InitiateDirectUploadBody struct {
	FileName    string `json:"fileName" maxLength:"255" doc:"Original file name"`
	ContentType string `json:"contentType" enum:"image/jpeg,image/png,image/webp,application/pdf" doc:"Media type"`
	SizeBytes   int64  `json:"sizeBytes" minimum:"1" doc:"Expected decoded content size"`
}

type InitiateAssetAttachmentDirectUploadOutput struct {
	Body shared.SuccessEnvelope[DirectUploadResponse]
}

type CompleteAssetAttachmentDirectUploadInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	AssetID       string `path:"assetId" doc:"Asset ID"`
	UploadID      string `path:"uploadId" doc:"Direct upload ID"`
}

type CompleteAssetAttachmentDirectUploadOutput struct {
	Body shared.SuccessEnvelope[AttachmentResponse]
}

type ListAssetAttachmentsInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	AssetID       string `path:"assetId" doc:"Asset ID"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type ListAssetAttachmentsOutput struct {
	Body shared.SuccessEnvelope[[]AttachmentResponse]
}

type DownloadAssetAttachmentInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	AssetID       string `path:"assetId" doc:"Asset ID"`
	AttachmentID  string `path:"attachmentId" doc:"Attachment ID"`
}

type DownloadAssetAttachmentOutput struct {
	ContentType        string `header:"Content-Type"`
	ContentDisposition string `header:"Content-Disposition"`
	Body               []byte
}

type DownloadAssetAttachmentThumbnailInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	AssetID       string `path:"assetId" doc:"Asset ID"`
	AttachmentID  string `path:"attachmentId" doc:"Attachment ID"`
	Variant       string `query:"variant" enum:"small" doc:"Thumbnail variant"`
}

type DownloadAssetAttachmentThumbnailOutput struct {
	ContentType string `header:"Content-Type"`
	Body        []byte
}

type GetAssetAttachmentInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	AssetID       string `path:"assetId" doc:"Asset ID"`
	AttachmentID  string `path:"attachmentId" doc:"Attachment ID"`
}

type GetAssetAttachmentOutput struct {
	Body shared.SuccessEnvelope[AttachmentResponse]
}

type UpdateAssetAttachmentLifecycleInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	AssetID       string `path:"assetId" doc:"Asset ID"`
	AttachmentID  string `path:"attachmentId" doc:"Attachment ID"`
}

type UpdateAssetAttachmentLifecycleOutput struct {
	Body shared.SuccessEnvelope[AttachmentResponse]
}

type DeleteAssetAttachmentOutput struct{}

type AttachmentResponse struct {
	ID             string `json:"id"`
	TenantID       string `json:"tenantId"`
	InventoryID    string `json:"inventoryId"`
	AssetID        string `json:"assetId"`
	FileName       string `json:"fileName"`
	ContentType    string `json:"contentType"`
	SizeBytes      int64  `json:"sizeBytes"`
	SHA256         string `json:"sha256"`
	CreatedAt      string `json:"createdAt"`
	LifecycleState string `json:"lifecycleState"`
}

type DirectUploadResponse struct {
	UploadID     string            `json:"uploadId"`
	AttachmentID string            `json:"attachmentId"`
	Method       string            `json:"method"`
	URL          string            `json:"url"`
	Headers      map[string]string `json:"headers"`
	FormFields   map[string]string `json:"formFields"`
	ExpiresAt    string            `json:"expiresAt"`
}
