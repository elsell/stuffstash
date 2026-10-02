package dto

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
)

type ArchiveAccessInput struct {
	Authorization string `header:"Authorization"`
	TenantID      string `path:"tenantId"`
	InventoryID   string `query:"inventoryId"`
}
type ArchiveJobInput struct {
	ArchiveAccessInput
	JobID string `path:"jobId"`
}
type ArchiveListInput struct {
	ArchiveAccessInput
	After string `query:"after"`
	Limit int    `query:"limit" default:"25" minimum:"1" maximum:"100"`
}
type CreateArchiveInput struct {
	Authorization  string `header:"Authorization"`
	TenantID       string `path:"tenantId"`
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"200"`
	Body           struct {
		InventoryID string `json:"inventoryId" minLength:"1"`
		Photos      bool   `json:"photos"`
		OtherFiles  bool   `json:"otherFiles"`
	}
}
type ApproveArchiveInput struct {
	ArchiveJobInput
	Body struct {
		Name string `json:"name" minLength:"1" maxLength:"200"`
	}
}
type ArchiveJob struct {
	ID                     string    `json:"id"`
	Kind                   string    `json:"kind"`
	State                  string    `json:"state"`
	Phase                  string    `json:"phase"`
	InventoryID            string    `json:"inventoryId,omitempty"`
	DestinationInventoryID string    `json:"destinationInventoryId,omitempty"`
	Photos                 bool      `json:"photos"`
	OtherFiles             bool      `json:"otherFiles"`
	CreatedAt              time.Time `json:"createdAt"`
	ExpiresAt              time.Time `json:"expiresAt"`
	Failure                string    `json:"failure,omitempty"`
}
type ArchiveKeyRemapping struct {
	Family         string `json:"family"`
	SourceKey      string `json:"sourceKey"`
	DestinationKey string `json:"destinationKey"`
}
type ArchivePreview struct {
	InventoryName      string                `json:"inventoryName"`
	Assets             int                   `json:"assets"`
	Tags               int                   `json:"tags"`
	CustomAssetTypes   int                   `json:"customAssetTypes"`
	CustomFields       int                   `json:"customFields"`
	Photos             int                   `json:"photos"`
	OtherFiles         int                   `json:"otherFiles"`
	OmittedAttachments int                   `json:"omittedAttachments"`
	KeyRemappings      []ArchiveKeyRemapping `json:"keyRemappings"`
}
type ArchiveJobOutput struct {
	Body shared.SuccessEnvelope[ArchiveJob]
}
type ArchiveListOutput struct {
	Body shared.SuccessEnvelope[[]ArchiveJob]
}
type ArchivePreviewOutput struct {
	Body shared.SuccessEnvelope[ArchivePreview]
}
