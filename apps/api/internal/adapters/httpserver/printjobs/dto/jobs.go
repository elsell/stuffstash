package dto

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"time"
)

type Scope struct {
	Authorization string `header:"Authorization"`
	RequestID     string `header:"X-Request-ID"`
	TenantID      string `path:"tenantId"`
	InventoryID   string `path:"inventoryId"`
}
type PrintJobTemplateOptions struct {
	ShowReference bool `json:"showReference"`
}
type PrintJobSelection struct {
	PrinterID                string                  `json:"printerId" minLength:"1"`
	ExpectedMediaFingerprint string                  `json:"expectedMediaFingerprint" minLength:"1"`
	TemplateID               string                  `json:"templateId"`
	TemplateVersion          uint32                  `json:"templateVersion" minimum:"1"`
	TemplateOptions          PrintJobTemplateOptions `json:"templateOptions"`
	Copies                   int                     `json:"copies" minimum:"1"`
	PreviewFingerprint       string                  `json:"previewFingerprint,omitempty"`
}
type CreateInput struct {
	Scope
	AssetID        string `path:"assetId"`
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"200"`
	Body           PrintJobSelection
}
type ReprintInput struct {
	JobInput
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"200"`
	Body           PrintJobSelection
}
type TestJobInput struct {
	Scope
	PrinterID      string `path:"printerId"`
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"200"`
	Body           PrintJobSelection
}
type JobInput struct {
	Scope
	JobID string `path:"jobId"`
}
type PrintJobRevision struct {
	Revision uint64 `json:"revision" minimum:"1"`
}
type CancelInput struct {
	JobInput
	Body PrintJobRevision
}
type ListInput struct {
	Scope
	PrinterID string `query:"printerId"`
	Limit     int    `query:"limit" default:"50" minimum:"1" maximum:"100"`
	Cursor    string `query:"cursor"`
}
type PrintJobAttempt struct {
	ID              string     `json:"id"`
	ConnectorID     string     `json:"connectorId"`
	ClaimedAt       time.Time  `json:"claimedAt"`
	LeaseExpiresAt  time.Time  `json:"leaseExpiresAt"`
	StartedAt       *time.Time `json:"startedAt,omitempty"`
	SettledAt       *time.Time `json:"settledAt,omitempty"`
	Outcome         string     `json:"outcome"`
	Reason          string     `json:"reason"`
	CompletedCopies int        `json:"completedCopies"`
}
type PrintJob struct {
	ID               string            `json:"id"`
	Predecessor      string            `json:"predecessor,omitempty"`
	PrinterID        string            `json:"printerId"`
	AssetID          string            `json:"assetId,omitempty"`
	Kind             string            `json:"kind"`
	Status           string            `json:"status"`
	Revision         uint64            `json:"revision"`
	Copies           int               `json:"copies"`
	MediaFingerprint string            `json:"mediaFingerprint"`
	RequestedBy      string            `json:"requestedBy"`
	CreatedAt        time.Time         `json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`
	Attempts         []PrintJobAttempt `json:"attempts"`
}
type Output struct {
	Status       int
	CacheControl string `header:"Cache-Control"`
	Body         shared.SuccessEnvelope[PrintJob]
}
type ListOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         shared.SuccessEnvelope[[]PrintJob]
}
