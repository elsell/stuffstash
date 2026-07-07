package importjob

import (
	"time"
)

type ID string

func (id ID) String() string {
	return string(id)
}

type Status string

const (
	StatusPreviewed          Status = "previewed"
	StatusRunning            Status = "running"
	StatusSucceeded          Status = "succeeded"
	StatusFailed             Status = "failed"
	StatusCancelRequested    Status = "cancel_requested"
	StatusCancelledKept      Status = "cancelled_kept"
	StatusCancelledDiscarded Status = "cancelled_discarded"
	StatusDiscardFailed      Status = "discard_failed"
)

type Phase string

const (
	PhasePreviewing  Phase = "previewing"
	PhaseReady       Phase = "ready"
	PhaseReading     Phase = "reading_source"
	PhaseFields      Phase = "creating_fields"
	PhaseTags        Phase = "creating_tags"
	PhaseLocations   Phase = "creating_locations"
	PhaseAssets      Phase = "creating_assets"
	PhaseAttachments Phase = "importing_attachments"
	PhaseFinalizing  Phase = "finalizing"
	PhaseTerminal    Phase = "terminal"
)

type CancellationMode string

const (
	CancellationModeKeepPartial    CancellationMode = "keep_partial_progress"
	CancellationModeDiscardPartial CancellationMode = "discard_partial_progress"
)

type SourceRef struct {
	Type                SourceType
	Name                string
	BaseURL             string
	Version             string
	ImageImport         string
	AllowPrivateNetwork bool
	AllowInsecureTLS    bool
	Fingerprint         string
}

type SourceType string

const (
	SourceTypeLegacyHomebox    SourceType = "legacy_homebox"
	SourceTypeLegacyHomeboxCSV SourceType = "legacy_homebox_csv"
)

type Counts struct {
	Fields               int
	Tags                 int
	Locations            int
	Assets               int
	Attachments          int
	Warnings             int
	Errors               int
	FieldsCreated        int
	FieldsExisting       int
	TagsCreated          int
	TagsExisting         int
	LocationsCreated     int
	AssetsCreated        int
	AssetsSkipped        int
	AttachmentsCreated   int
	AttachmentsSkipped   int
	RecordsDiscarded     int
	SourceLinksDiscarded int
}

type Progress struct {
	Phase     Phase
	Done      int
	Total     int
	Message   string
	UpdatedAt time.Time
}

type PreviewSummary struct {
	Fields               []PreviewField
	Tags                 []PreviewTag
	Locations            []PreviewAsset
	Assets               []PreviewAsset
	Attachments          []PreviewAttachment
	Messages             []Message
	FieldsTruncated      bool
	TagsTruncated        bool
	LocationsTruncated   bool
	AssetsTruncated      bool
	AttachmentsTruncated bool
	MessagesTruncated    bool
}

type Message struct {
	Code       string
	Severity   MessageSeverity
	Summary    string
	Detail     string
	SourceID   string
	SourceName string
}

type MessageSeverity string

const (
	MessageSeverityWarning MessageSeverity = "warning"
	MessageSeverityError   MessageSeverity = "error"
)

type PreviewField struct {
	Key         string
	DisplayName string
	Type        string
}

type PreviewTag struct {
	Key         string
	DisplayName string
	Color       string
}

type PreviewAsset struct {
	SourceID       string
	Kind           string
	Title          string
	ParentSourceID string
	Archived       bool
}

type PreviewAttachment struct {
	SourceID      string
	AssetSourceID string
	FileName      string
	ContentType   string
	SizeBytes     int
	Primary       bool
}

type ResourceSummary struct {
	ResourceType     string
	ResourceID       string
	DisplayName      string
	ResourceOwnerID  string
	SourceEntityType string
	SourceEntityID   string
	CreatedAt        time.Time
}

type Record struct {
	ID                    ID
	TenantID              TenantID
	InventoryID           InventoryID
	ActorID               PrincipalID
	Status                Status
	Source                SourceRef
	Counts                Counts
	Preview               PreviewSummary
	Resources             []ResourceSummary
	Messages              []Message
	Progress              Progress
	ProgressHistory       []Progress
	CancellationMode      CancellationMode
	CancellationRequestID string
	HistoryRemovedAt      time.Time
	CreatedAt             time.Time
	StartedAt             time.Time
	CompletedAt           time.Time
	UpdatedAt             time.Time
}

type TenantID string

func (id TenantID) String() string {
	return string(id)
}

type InventoryID string

func (id InventoryID) String() string {
	return string(id)
}

type PrincipalID string

func (id PrincipalID) String() string {
	return string(id)
}

func NewPreviewedRecord(id ID, tenantID TenantID, inventoryID InventoryID, actorID PrincipalID, source SourceRef, counts Counts, messages []Message, now time.Time) Record {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	progress := Progress{
		Phase:     PhaseReady,
		Done:      counts.Fields + counts.Tags + counts.Locations + counts.Assets + counts.Attachments,
		Total:     counts.Fields + counts.Tags + counts.Locations + counts.Assets + counts.Attachments,
		Message:   "Preview ready",
		UpdatedAt: now,
	}
	return Record{
		ID:          id,
		TenantID:    tenantID,
		InventoryID: inventoryID,
		ActorID:     actorID,
		Status:      StatusPreviewed,
		Source:      source,
		Counts:      counts,
		Messages:    append([]Message{}, messages...),
		Progress:    progress,
		ProgressHistory: []Progress{
			progress,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func AppendProgressHistory(history []Progress, progress Progress) []Progress {
	if progress.Phase == "" {
		return cloneProgressHistory(history)
	}
	out := cloneProgressHistory(history)
	if len(out) > 0 {
		last := out[len(out)-1]
		if last.Phase == progress.Phase && last.Message == progress.Message {
			out[len(out)-1] = progress
			return out
		}
	}
	out = append(out, progress)
	return out
}

func cloneProgressHistory(history []Progress) []Progress {
	return append([]Progress{}, history...)
}
