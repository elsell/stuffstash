package dto

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"time"
)

type PrintConsumerInput struct {
	Authorization string `header:"Authorization"`
}
type PrintAttemptInput struct {
	PrintConsumerInput
	AttemptID string `path:"attemptId"`
}
type PrintClaimProof struct {
	SessionID  string `json:"sessionId" minLength:"16" maxLength:"100"`
	ClaimToken string `json:"claimToken" minLength:"43" maxLength:"43"`
	Revision   uint64 `json:"revision" minimum:"1"`
}
type PrintClaimRequest struct {
	PrinterID  string `json:"printerId" minLength:"1"`
	AttemptID  string `json:"attemptId" minLength:"16" maxLength:"100"`
	SessionID  string `json:"sessionId" minLength:"16" maxLength:"100"`
	ClaimToken string `json:"claimToken" minLength:"43" maxLength:"43"`
}
type PrintClaimInput struct {
	PrintConsumerInput
	Body PrintClaimRequest
}
type PrintAttemptMutation struct {
	PrintAttemptInput
	Body PrintClaimProof
}
type PrintOutcome struct {
	Kind            string `json:"kind" enum:"completed,no_output,uncertain"`
	CompletedCopies int    `json:"completedCopies" minimum:"0"`
	Retryable       bool   `json:"retryable"`
	Reason          string `json:"reason" enum:",lease_expired,canceled,device_unavailable,invalid_artifact,device_failure,partial_output,unknown"`
}
type PrintOutcomeRequest struct {
	PrintClaimProof
	Outcome PrintOutcome `json:"outcome"`
}
type PrintOutcomeInput struct {
	PrintAttemptInput
	Body PrintOutcomeRequest
}
type PrintContentInput struct {
	PrintAttemptInput
	SessionID  string `header:"X-Print-Session-ID" required:"true"`
	ClaimToken string `header:"X-Print-Claim-Token" required:"true"`
	Revision   uint64 `header:"X-Print-Revision" required:"true" minimum:"1"`
}
type PrintArtifact struct {
	SHA256       string    `json:"sha256"`
	ContentType  string    `json:"contentType"`
	ByteLength   int64     `json:"byteLength"`
	WidthPixels  int       `json:"widthPixels"`
	HeightPixels int       `json:"heightPixels"`
	ExpiresAt    time.Time `json:"expiresAt"`
}
type PrintMediaMargins struct {
	Left   int `json:"left"`
	Right  int `json:"right"`
	Top    int `json:"top"`
	Bottom int `json:"bottom"`
}
type PrintConsumerMedia struct {
	MarginsMicrometers PrintMediaMargins `json:"marginsMicrometers"`
	DisplayRotation    int               `json:"displayRotation"`
	PresetID           string            `json:"presetId"`
	Version            uint32            `json:"version"`
	WidthMicrometers   int               `json:"widthMicrometers"`
	HeightMicrometers  int               `json:"heightMicrometers"`
	ResolutionDPI      int               `json:"resolutionDpi"`
	RasterWidth        int               `json:"rasterWidth"`
	RasterHeight       int               `json:"rasterHeight"`
	Orientation        string            `json:"orientation"`
	ColorMode          string            `json:"colorMode"`
	CutPolicy          string            `json:"cutPolicy"`
}
type PrintConsumerAttempt struct {
	ProtocolVersion  int                 `json:"protocolVersion"`
	JobID            string              `json:"jobId"`
	PrinterID        string              `json:"printerId"`
	AttemptID        string              `json:"attemptId"`
	SessionID        string              `json:"sessionId"`
	Status           string              `json:"status"`
	Revision         uint64              `json:"revision"`
	Copies           int                 `json:"copies"`
	LeaseExpiresAt   time.Time           `json:"leaseExpiresAt"`
	LeaseValid       bool                `json:"leaseValid"`
	StartedAt        *time.Time          `json:"startedAt,omitempty"`
	SettledAt        *time.Time          `json:"settledAt,omitempty"`
	Outcome          PrintOutcome        `json:"outcome"`
	MediaFingerprint string              `json:"mediaFingerprint"`
	Media            *PrintConsumerMedia `json:"media,omitempty"`
	Artifact         *PrintArtifact      `json:"artifact,omitempty"`
}
type PrintConsumerOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         shared.SuccessEnvelope[*PrintConsumerAttempt]
}
type PrintContentOutput struct {
	CacheControl string `header:"Cache-Control"`
	ContentType  string `header:"Content-Type"`
	Body         []byte
}

type PrintRecoveryInput struct {
	PrintConsumerInput
	PrinterID string `query:"printerId"`
	Status    string `query:"status" default:"unsettled" enum:"unsettled"`
	Limit     int    `query:"limit" default:"50" minimum:"1" maximum:"100"`
	Cursor    string `query:"cursor"`
}
type PrintRecoveryOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         shared.SuccessEnvelope[[]PrintConsumerAttempt]
}
type PrintReconciliation struct {
	Revision uint64       `json:"revision" minimum:"1"`
	Outcome  PrintOutcome `json:"outcome"`
}
type PrintReconcileInput struct {
	PrintAttemptInput
	Body PrintReconciliation
}
