package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type LegacyHomeboxPreviewInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Body          LegacyHomeboxImportRequest
}

type LegacyHomeboxApplyInput = LegacyHomeboxPreviewInput

type LegacyHomeboxImportRequest struct {
	SourceType          string `json:"sourceType" enum:"legacy_homebox,legacy_homebox_csv" doc:"Import source type"`
	BaseURL             string `json:"baseUrl,omitempty" doc:"Homebox base URL for live imports"`
	Username            string `json:"username,omitempty" doc:"Homebox username for live imports"`
	Password            string `json:"password,omitempty" doc:"Homebox password for live imports"`
	IncludeImages       bool   `json:"includeImages,omitempty" doc:"Import supported image attachments for live imports"`
	AllowInsecureTLS    bool   `json:"allowInsecureTLS,omitempty" doc:"Allow self-signed or otherwise untrusted Homebox TLS certificates"`
	AllowPrivateNetwork bool   `json:"allowPrivateNetwork,omitempty" doc:"Allow Homebox URLs that resolve to private or local network addresses"`
	FileName            string `json:"fileName,omitempty" doc:"Uploaded CSV file name"`
	ContentBase64       string `json:"contentBase64,omitempty" doc:"Base64-encoded Homebox CSV content"`
}

type LegacyHomeboxPreviewOutput struct {
	Body shared.SuccessEnvelope[ImportPreviewResponse]
}

type LegacyHomeboxApplyOutput struct {
	Body shared.SuccessEnvelope[ImportApplyResponse]
}

type ImportPreviewResponse struct {
	Source       ImportSourceResponse    `json:"source"`
	Counts       ImportCountsResponse    `json:"counts"`
	Fields       []ImportFieldResponse   `json:"fields"`
	AssetSamples []ImportAssetSample     `json:"assetSamples"`
	ImageSamples []ImportImageSample     `json:"imageSamples"`
	Messages     []ImportMessageResponse `json:"messages"`
}

type ImportApplyResponse struct {
	Counts   ImportApplyCountsResponse `json:"counts"`
	Messages []ImportMessageResponse   `json:"messages"`
}

type ImportSourceResponse struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	BaseURL     string `json:"baseUrl,omitempty"`
	Version     string `json:"version,omitempty"`
	ImageImport string `json:"imageImport"`
}

type ImportCountsResponse struct {
	Fields      int `json:"fields"`
	Locations   int `json:"locations"`
	Assets      int `json:"assets"`
	Attachments int `json:"attachments"`
	Warnings    int `json:"warnings"`
	Errors      int `json:"errors"`
}

type ImportApplyCountsResponse struct {
	FieldsCreated      int `json:"fieldsCreated"`
	FieldsExisting     int `json:"fieldsExisting"`
	LocationsCreated   int `json:"locationsCreated"`
	AssetsCreated      int `json:"assetsCreated"`
	AssetsSkipped      int `json:"assetsSkipped"`
	AttachmentsCreated int `json:"attachmentsCreated"`
	AttachmentsSkipped int `json:"attachmentsSkipped"`
}

type ImportFieldResponse struct {
	Key         string `json:"key"`
	DisplayName string `json:"displayName"`
	Type        string `json:"type"`
}

type ImportAssetSample struct {
	SourceID       string         `json:"sourceId"`
	Kind           string         `json:"kind"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	ParentSourceID string         `json:"parentSourceId,omitempty"`
	CustomFields   map[string]any `json:"customFields"`
}

type ImportImageSample struct {
	SourceID      string `json:"sourceId"`
	AssetSourceID string `json:"assetSourceId"`
	FileName      string `json:"fileName"`
	ContentType   string `json:"contentType"`
	SizeBytes     int    `json:"sizeBytes"`
	Primary       bool   `json:"primary"`
}

type ImportMessageResponse struct {
	Code       string `json:"code"`
	Severity   string `json:"severity"`
	Summary    string `json:"summary"`
	Detail     string `json:"detail,omitempty"`
	SourceID   string `json:"sourceId,omitempty"`
	SourceName string `json:"sourceName,omitempty"`
}
