package ports

import "context"

type ImportJob struct {
	Actor            *ImportJobActor     `json:"actor,omitempty"`
	ActorID          *string             `json:"actorId,omitempty"`
	CancellationMode *string             `json:"cancellationMode,omitempty"`
	CompletedAt      *string             `json:"completedAt,omitempty"`
	Counts           ImportJobCounts     `json:"counts"`
	CreatedAt        string              `json:"createdAt"`
	ID               string              `json:"id"`
	Messages         []ImportMessage     `json:"messages"`
	Preview          ImportJobPreview    `json:"preview"`
	Progress         ImportJobProgress   `json:"progress"`
	ProgressHistory  []ImportJobProgress `json:"progressHistory"`
	Resources        []ImportJobResource `json:"resources"`
	Source           ImportJobSource     `json:"source"`
	StartedAt        *string             `json:"startedAt,omitempty"`
	Status           string              `json:"status"`
	UpdatedAt        string              `json:"updatedAt"`
}

type ImportJobActor struct {
	Email *string `json:"email,omitempty"`
	ID    string  `json:"id"`
}

type ImportJobCounts struct {
	Assets               int64 `json:"assets"`
	AssetsCreated        int64 `json:"assetsCreated"`
	AssetsSkipped        int64 `json:"assetsSkipped"`
	Attachments          int64 `json:"attachments"`
	AttachmentsCreated   int64 `json:"attachmentsCreated"`
	AttachmentsSkipped   int64 `json:"attachmentsSkipped"`
	Errors               int64 `json:"errors"`
	Fields               int64 `json:"fields"`
	FieldsCreated        int64 `json:"fieldsCreated"`
	FieldsExisting       int64 `json:"fieldsExisting"`
	Locations            int64 `json:"locations"`
	LocationsCreated     int64 `json:"locationsCreated"`
	RecordsDiscarded     int64 `json:"recordsDiscarded"`
	SourceLinksDiscarded int64 `json:"sourceLinksDiscarded"`
	Tags                 int64 `json:"tags"`
	TagsCreated          int64 `json:"tagsCreated"`
	TagsExisting         int64 `json:"tagsExisting"`
	Warnings             int64 `json:"warnings"`
}

type ImportMessage struct {
	Code       string  `json:"code"`
	Detail     *string `json:"detail,omitempty"`
	Severity   string  `json:"severity"`
	SourceID   *string `json:"sourceId,omitempty"`
	SourceName *string `json:"sourceName,omitempty"`
	Summary    string  `json:"summary"`
}

type ImportJobPreview struct {
	Assets               []ImportJobPreviewAsset      `json:"assets"`
	AssetsTruncated      bool                         `json:"assetsTruncated"`
	Attachments          []ImportJobPreviewAttachment `json:"attachments"`
	AttachmentsTruncated bool                         `json:"attachmentsTruncated"`
	Fields               []ImportJobPreviewField      `json:"fields"`
	FieldsTruncated      bool                         `json:"fieldsTruncated"`
	Locations            []ImportJobPreviewAsset      `json:"locations"`
	LocationsTruncated   bool                         `json:"locationsTruncated"`
	Messages             []ImportMessage              `json:"messages"`
	MessagesTruncated    bool                         `json:"messagesTruncated"`
	Tags                 []ImportJobPreviewTag        `json:"tags"`
	TagsTruncated        bool                         `json:"tagsTruncated"`
}

type ImportJobPreviewAsset struct {
	Archived       bool    `json:"archived"`
	Kind           string  `json:"kind"`
	ParentSourceID *string `json:"parentSourceId,omitempty"`
	SourceID       *string `json:"sourceId,omitempty"`
	Title          string  `json:"title"`
}

type ImportJobPreviewAttachment struct {
	AssetSourceID *string `json:"assetSourceId,omitempty"`
	ContentType   string  `json:"contentType"`
	FileName      string  `json:"fileName"`
	Primary       bool    `json:"primary"`
	SizeBytes     int64   `json:"sizeBytes"`
	SourceID      *string `json:"sourceId,omitempty"`
}

type ImportJobPreviewField struct {
	DisplayName string `json:"displayName"`
	Key         string `json:"key"`
	Type        string `json:"type"`
}

type ImportJobPreviewTag struct {
	Color       *string `json:"color,omitempty"`
	DisplayName string  `json:"displayName"`
	Key         string  `json:"key"`
}

type ImportJobProgress struct {
	Done      int64   `json:"done"`
	Message   *string `json:"message,omitempty"`
	Phase     string  `json:"phase"`
	Total     int64   `json:"total"`
	UpdatedAt *string `json:"updatedAt,omitempty"`
}

type ImportJobResource struct {
	CreatedAt        string  `json:"createdAt"`
	DisplayName      *string `json:"displayName,omitempty"`
	ResourceID       string  `json:"resourceId"`
	ResourceOwnerID  *string `json:"resourceOwnerId,omitempty"`
	ResourceType     string  `json:"resourceType"`
	SourceEntityID   string  `json:"sourceEntityId"`
	SourceEntityType string  `json:"sourceEntityType"`
}

type ImportJobSource struct {
	AllowInsecureTLS    bool    `json:"allowInsecureTLS"`
	AllowPrivateNetwork bool    `json:"allowPrivateNetwork"`
	BaseUrl             *string `json:"baseUrl,omitempty"`
	Fingerprint         *string `json:"fingerprint,omitempty"`
	ImageImport         string  `json:"imageImport"`
	Name                string  `json:"name"`
	Type                string  `json:"type"`
	Version             *string `json:"version,omitempty"`
}

type ImportJobList struct {
	Jobs []ImportJob `json:"jobs"`
}
type ImportJobsAPI interface {
	ImportJobs(context.Context, Scope) (Result[ImportJobList], error)
	ImportJob(context.Context, Scope, string) (Result[ImportJob], error)
	DeleteImportJob(context.Context, Scope, string) error
}
