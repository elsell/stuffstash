package dto

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
)

type Access struct {
	Authorization string `header:"Authorization"`
	TenantID      string `path:"tenantId"`
	InventoryID   string `query:"inventoryId"`
}
type JobInput struct {
	Access
	JobID string `path:"jobId"`
}
type ListInput struct {
	Access
	After string `query:"after"`
	Limit int    `query:"limit" default:"25" minimum:"1" maximum:"100"`
}
type CreateInput struct {
	Authorization  string `header:"Authorization"`
	TenantID       string `path:"tenantId"`
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"200"`
	Body           struct {
		InventoryID string `json:"inventoryId" minLength:"1"`
		Photos      bool   `json:"photos"`
		OtherFiles  bool   `json:"otherFiles"`
	}
}
type ApproveInput struct {
	JobInput
	Body struct {
		Name string `json:"name" minLength:"1" maxLength:"200"`
	}
}
type Job struct {
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
type KeyRemapping struct {
	Family         string `json:"family"`
	SourceKey      string `json:"sourceKey"`
	DestinationKey string `json:"destinationKey"`
}
type Preview struct {
	InventoryName      string         `json:"inventoryName"`
	Assets             int            `json:"assets"`
	Tags               int            `json:"tags"`
	CustomAssetTypes   int            `json:"customAssetTypes"`
	CustomFields       int            `json:"customFields"`
	Photos             int            `json:"photos"`
	OtherFiles         int            `json:"otherFiles"`
	OmittedAttachments int            `json:"omittedAttachments"`
	KeyRemappings      []KeyRemapping `json:"keyRemappings"`
}
type JobOutput struct{ Body shared.SuccessEnvelope[Job] }
type ListOutput struct{ Body shared.SuccessEnvelope[[]Job] }
type PreviewOutput struct {
	Body shared.SuccessEnvelope[Preview]
}
