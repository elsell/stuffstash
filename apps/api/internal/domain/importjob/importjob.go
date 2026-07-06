package importjob

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
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
	Type        importplan.SourceType
	Name        string
	BaseURL     string
	Version     string
	ImageImport string
	Fingerprint string
}

type Counts struct {
	Fields               int
	Locations            int
	Assets               int
	Attachments          int
	Warnings             int
	Errors               int
	FieldsCreated        int
	FieldsExisting       int
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
	Locations            []PreviewAsset
	Assets               []PreviewAsset
	Attachments          []PreviewAttachment
	Messages             []importplan.Message
	FieldsTruncated      bool
	LocationsTruncated   bool
	AssetsTruncated      bool
	AttachmentsTruncated bool
	MessagesTruncated    bool
}

type PreviewField struct {
	Key         string
	DisplayName string
	Type        string
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
	ResourceOwnerID  string
	SourceEntityType string
	SourceEntityID   string
	CreatedAt        time.Time
}

type Record struct {
	ID                    ID
	TenantID              tenant.ID
	InventoryID           inventory.InventoryID
	ActorID               identity.PrincipalID
	Status                Status
	Source                SourceRef
	Counts                Counts
	Preview               PreviewSummary
	Resources             []ResourceSummary
	Messages              []importplan.Message
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

func NewPreviewedRecord(id ID, tenantID tenant.ID, inventoryID inventory.InventoryID, actorID identity.PrincipalID, source SourceRef, counts Counts, messages []importplan.Message, now time.Time) Record {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	progress := Progress{
		Phase:     PhaseReady,
		Done:      counts.Fields + counts.Locations + counts.Assets + counts.Attachments,
		Total:     counts.Fields + counts.Locations + counts.Assets + counts.Attachments,
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
		Messages:    append([]importplan.Message{}, messages...),
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
			return out
		}
	}
	out = append(out, progress)
	return out
}

func cloneProgressHistory(history []Progress) []Progress {
	return append([]Progress{}, history...)
}

func CountsFromPlan(plan importplan.Plan) Counts {
	counts := plan.Counts()
	return Counts{
		Fields:      counts.Fields,
		Locations:   counts.Locations,
		Assets:      counts.Assets,
		Attachments: counts.Attachments,
		Warnings:    counts.Warnings,
		Errors:      counts.Errors,
	}
}

func PreviewSummaryFromPlan(plan importplan.Plan, limit int) PreviewSummary {
	if limit <= 0 {
		limit = 12
	}
	summary := PreviewSummary{
		FieldsTruncated:      len(plan.Fields) > limit,
		AttachmentsTruncated: len(plan.Attachments) > limit,
		MessagesTruncated:    len(plan.Messages) > limit,
	}
	for _, field := range plan.Fields {
		if len(summary.Fields) >= limit {
			break
		}
		summary.Fields = append(summary.Fields, PreviewField{
			Key:         field.Key,
			DisplayName: field.DisplayName,
			Type:        field.Type,
		})
	}
	for _, item := range plan.Assets {
		if item.Kind != "location" {
			continue
		}
		if len(summary.Locations) >= limit {
			summary.LocationsTruncated = true
			continue
		}
		summary.Locations = append(summary.Locations, PreviewAsset{
			SourceID:       item.SourceID,
			Kind:           item.Kind,
			Title:          item.Title,
			ParentSourceID: item.ParentSourceID,
			Archived:       item.Archived,
		})
	}
	for _, item := range plan.Assets {
		if item.Kind == "location" {
			continue
		}
		if len(summary.Assets) >= limit {
			summary.AssetsTruncated = true
			continue
		}
		summary.Assets = append(summary.Assets, PreviewAsset{
			SourceID:       item.SourceID,
			Kind:           item.Kind,
			Title:          item.Title,
			ParentSourceID: item.ParentSourceID,
			Archived:       item.Archived,
		})
	}
	for _, attachment := range plan.Attachments {
		if len(summary.Attachments) >= limit {
			break
		}
		summary.Attachments = append(summary.Attachments, PreviewAttachment{
			SourceID:      attachment.SourceID,
			AssetSourceID: attachment.AssetSourceID,
			FileName:      attachment.FileName,
			ContentType:   attachment.ContentType,
			SizeBytes:     attachment.SizeBytes,
			Primary:       attachment.Primary,
		})
	}
	for _, message := range plan.Messages {
		if len(summary.Messages) >= limit {
			break
		}
		summary.Messages = append(summary.Messages, message)
	}
	return summary
}

func SourceRefFromPlan(plan importplan.Plan, fingerprint string) SourceRef {
	return SourceRef{
		Type:        plan.Source.Type,
		Name:        plan.Source.Name,
		BaseURL:     plan.Source.BaseURL,
		Version:     plan.Source.Version,
		ImageImport: plan.Source.ImageImport,
		Fingerprint: fingerprint,
	}
}
