package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type ImportSourceRequest struct {
	SourceType          string `json:"sourceType" enum:"legacy_homebox,legacy_homebox_csv" doc:"Import source type"`
	BaseURL             string `json:"baseUrl,omitempty" doc:"Source base URL for live-source imports"`
	Username            string `json:"username,omitempty" doc:"Source username for live-source imports"`
	Password            string `json:"password,omitempty" doc:"Source password for live-source imports"`
	IncludeImages       bool   `json:"includeImages,omitempty" doc:"Import supported image attachments when the source supports them"`
	AllowInsecureTLS    bool   `json:"allowInsecureTLS,omitempty" doc:"Allow self-signed or otherwise untrusted TLS certificates for live-source imports"`
	AllowPrivateNetwork bool   `json:"allowPrivateNetwork,omitempty" doc:"Allow source URLs that resolve to private or local network addresses"`
	FileName            string `json:"fileName,omitempty" doc:"Uploaded source file name"`
	ContentBase64       string `json:"contentBase64,omitempty" doc:"Base64-encoded source file content"`
}

type ImportJobListInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
}

type ImportJobDetailInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	JobID         string `path:"jobId" doc:"Import job ID"`
}

type RemoveImportJobInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	JobID         string `path:"jobId" doc:"Import job ID"`
}

type ImportJobPreviewInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Body          ImportSourceRequest
}

type ImportJobStartInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	JobID         string `path:"jobId" doc:"Import job ID"`
	Body          ImportSourceRequest
}

type ImportJobCancelInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	JobID         string `path:"jobId" doc:"Import job ID"`
	Body          ImportJobCancelRequest
}

type ImportJobCancelRequest struct {
	Mode string `json:"mode" enum:"keep_partial_progress,discard_partial_progress" doc:"Cancellation behavior for already imported records"`
}

type ImportJobListOutput struct {
	Body shared.SuccessEnvelope[ImportJobListResponse]
}

type ImportJobOutput struct {
	Body shared.SuccessEnvelope[ImportJobResponse]
}

type RemoveImportJobOutput struct{}

type ImportJobListResponse struct {
	Jobs []ImportJobResponse `json:"jobs"`
}

type ImportJobResponse struct {
	ID               string                  `json:"id"`
	Status           string                  `json:"status"`
	ActorID          string                  `json:"actorId,omitempty"`
	Source           ImportJobSourceResponse `json:"source"`
	Counts           ImportJobCountsResponse `json:"counts"`
	Preview          ImportJobPreview        `json:"preview"`
	Progress         ImportJobProgress       `json:"progress"`
	ProgressHistory  []ImportJobProgress     `json:"progressHistory"`
	CancellationMode string                  `json:"cancellationMode,omitempty"`
	CreatedAt        string                  `json:"createdAt"`
	StartedAt        string                  `json:"startedAt,omitempty"`
	CompletedAt      string                  `json:"completedAt,omitempty"`
	UpdatedAt        string                  `json:"updatedAt"`
	Resources        []ImportJobResource     `json:"resources"`
	Messages         []ImportMessageResponse `json:"messages"`
}

type ImportJobSourceResponse struct {
	Type                string `json:"type"`
	Name                string `json:"name"`
	BaseURL             string `json:"baseUrl,omitempty"`
	Version             string `json:"version,omitempty"`
	ImageImport         string `json:"imageImport"`
	AllowPrivateNetwork bool   `json:"allowPrivateNetwork"`
	AllowInsecureTLS    bool   `json:"allowInsecureTLS"`
	Fingerprint         string `json:"fingerprint,omitempty"`
}

type ImportJobCountsResponse struct {
	Fields               int `json:"fields"`
	Locations            int `json:"locations"`
	Assets               int `json:"assets"`
	Attachments          int `json:"attachments"`
	Warnings             int `json:"warnings"`
	Errors               int `json:"errors"`
	FieldsCreated        int `json:"fieldsCreated"`
	FieldsExisting       int `json:"fieldsExisting"`
	LocationsCreated     int `json:"locationsCreated"`
	AssetsCreated        int `json:"assetsCreated"`
	AssetsSkipped        int `json:"assetsSkipped"`
	AttachmentsCreated   int `json:"attachmentsCreated"`
	AttachmentsSkipped   int `json:"attachmentsSkipped"`
	RecordsDiscarded     int `json:"recordsDiscarded"`
	SourceLinksDiscarded int `json:"sourceLinksDiscarded"`
}

type ImportJobPreview struct {
	Fields               []ImportJobPreviewField      `json:"fields"`
	Locations            []ImportJobPreviewAsset      `json:"locations"`
	Assets               []ImportJobPreviewAsset      `json:"assets"`
	Attachments          []ImportJobPreviewAttachment `json:"attachments"`
	Messages             []ImportMessageResponse      `json:"messages"`
	FieldsTruncated      bool                         `json:"fieldsTruncated"`
	LocationsTruncated   bool                         `json:"locationsTruncated"`
	AssetsTruncated      bool                         `json:"assetsTruncated"`
	AttachmentsTruncated bool                         `json:"attachmentsTruncated"`
	MessagesTruncated    bool                         `json:"messagesTruncated"`
}

type ImportJobPreviewField struct {
	Key         string `json:"key"`
	DisplayName string `json:"displayName"`
	Type        string `json:"type"`
}

type ImportJobPreviewAsset struct {
	SourceID       string `json:"sourceId,omitempty"`
	Kind           string `json:"kind"`
	Title          string `json:"title"`
	ParentSourceID string `json:"parentSourceId,omitempty"`
	Archived       bool   `json:"archived"`
}

type ImportJobPreviewAttachment struct {
	SourceID      string `json:"sourceId,omitempty"`
	AssetSourceID string `json:"assetSourceId,omitempty"`
	FileName      string `json:"fileName"`
	ContentType   string `json:"contentType"`
	SizeBytes     int    `json:"sizeBytes"`
	Primary       bool   `json:"primary"`
}

type ImportJobProgress struct {
	Phase     string `json:"phase"`
	Done      int    `json:"done"`
	Total     int    `json:"total"`
	Message   string `json:"message,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

type ImportJobResource struct {
	ResourceType     string `json:"resourceType"`
	ResourceID       string `json:"resourceId"`
	DisplayName      string `json:"displayName,omitempty"`
	ResourceOwnerID  string `json:"resourceOwnerId,omitempty"`
	SourceEntityType string `json:"sourceEntityType"`
	SourceEntityID   string `json:"sourceEntityId"`
	CreatedAt        string `json:"createdAt"`
}

type ImportMessageResponse struct {
	Code       string `json:"code"`
	Severity   string `json:"severity"`
	Summary    string `json:"summary"`
	Detail     string `json:"detail,omitempty"`
	SourceID   string `json:"sourceId,omitempty"`
	SourceName string `json:"sourceName,omitempty"`
}
