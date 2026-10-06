package ports

import (
	"context"
	"time"
)

type ArchiveJob struct {
	ID                     string    `json:"id"`
	CreatedAt              time.Time `json:"createdAt"`
	ExpiresAt              time.Time `json:"expiresAt"`
	DestinationInventoryID *string   `json:"destinationInventoryId,omitempty"`
	InventoryID            *string   `json:"inventoryId,omitempty"`
	Failure                *string   `json:"failure,omitempty"`
	Kind                   string    `json:"kind"`
	OtherFiles             bool      `json:"otherFiles"`
	Photos                 bool      `json:"photos"`
	Phase                  string    `json:"phase"`
	State                  string    `json:"state"`
}
type ArchiveKeyRemapping struct {
	DestinationKey string `json:"destinationKey"`
	SourceKey      string `json:"sourceKey"`
	Family         string `json:"family"`
}
type ArchivePreview struct {
	Assets             int64                 `json:"assets"`
	CustomAssetTypes   int64                 `json:"customAssetTypes"`
	CustomFields       int64                 `json:"customFields"`
	InventoryName      string                `json:"inventoryName"`
	KeyRemappings      []ArchiveKeyRemapping `json:"keyRemappings"`
	OmittedAttachments int64                 `json:"omittedAttachments"`
	OtherFiles         int64                 `json:"otherFiles"`
	Photos             int64                 `json:"photos"`
	Tags               int64                 `json:"tags"`
}
type ArchiveJobsAPI interface {
	ArchiveJobs(context.Context, Scope, Page) (Result[[]ArchiveJob], error)
	ArchiveJob(context.Context, Scope, string) (Result[ArchiveJob], error)
	ArchivePreview(context.Context, Scope, string) (Result[ArchivePreview], error)
	RetryArchiveJob(context.Context, Scope, string) (Result[ArchiveJob], error)
	DeleteArchiveJob(context.Context, Scope, string) error
}
