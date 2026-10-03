package dto

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"time"
)

type PairingCandidate struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AdapterID string `json:"adapterId"`
	DeviceID  string `json:"deviceId"`
}
type BeginPairingInput struct {
	Body struct {
		Rotation   bool               `json:"rotation,omitempty"`
		Name       string             `json:"name" minLength:"1" maxLength:"100"`
		PublicKey  []byte             `json:"publicKey"`
		Candidates []PairingCandidate `json:"candidates" maxItems:"16"`
	}
}
type PairingStarted struct {
	VerificationURL string    `json:"verificationUrl"`
	ID              string    `json:"id"`
	PollToken       string    `json:"pollToken"`
	UserCode        string    `json:"userCode"`
	ExpiresAt       time.Time `json:"expiresAt"`
}
type PairingStartedOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         shared.SuccessEnvelope[PairingStarted]
}
type PairingInput struct {
	PairingID string `path:"pairingId"`
	PollToken string `header:"X-Pairing-Token" required:"true"`
}
type PairingStatus struct {
	ID        string    `json:"id"`
	State     string    `json:"state"`
	ExpiresAt time.Time `json:"expiresAt"`
}
type PairingStatusOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         shared.SuccessEnvelope[PairingStatus]
}
type PairingBinding struct {
	CandidateID string `json:"candidateId"`
	PrinterID   string `json:"printerId"`
}
type ApprovePairingInput struct {
	Authorization string `header:"Authorization"`
	RequestID     string `header:"X-Request-ID"`
	PairingID     string `path:"pairingId"`
	Body          struct {
		UserCode    string           `json:"userCode"`
		TenantID    string           `json:"tenantId"`
		InventoryID string           `json:"inventoryId"`
		Bindings    []PairingBinding `json:"bindings" minItems:"1" maxItems:"16"`
	}
}
type Connector struct {
	Availability         string     `json:"availability" enum:"online,offline,unknown"`
	Generation           uint64     `json:"generation"`
	PrinterIDs           []string   `json:"printerIds"`
	ID                   string     `json:"id"`
	Name                 string     `json:"name"`
	State                string     `json:"state"`
	AuthorizationPending bool       `json:"authorizationPending"`
	LastSeenAt           *time.Time `json:"lastSeenAt,omitempty"`
}
type ConnectorOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         shared.SuccessEnvelope[Connector]
}
type ExchangePairingInput struct {
	PairingInput
	Body struct {
		Signature []byte `json:"signature"`
	}
}
type PairingCredential struct {
	Credential         string    `json:"credential"`
	ConnectorID        string    `json:"connectorId"`
	TenantID           string    `json:"tenantId"`
	InventoryID        string    `json:"inventoryId"`
	ExpiresAt          time.Time `json:"expiresAt"`
	ActivationDeadline time.Time `json:"activationDeadline"`
}
type PairingCredentialOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         shared.SuccessEnvelope[PairingCredential]
}
type HeartbeatInput struct {
	Authorization string `header:"Authorization"`
	Body          struct {
		SessionID string `json:"sessionId" minLength:"1" maxLength:"100"`
	}
}

type ConnectorInput struct {
	PrinterScope
	ConnectorID string `path:"connectorId"`
}
type UpdateConnectorInput struct {
	ConnectorInput
	Body struct {
		Generation uint64    `json:"generation" minimum:"1"`
		Name       *string   `json:"name,omitempty"`
		Revoked    *bool     `json:"revoked,omitempty"`
		PrinterIDs *[]string `json:"printerIds,omitempty"`
	}
}

type ConsumerInput struct {
	Authorization string `header:"Authorization"`
}
type ConsumerPrinter struct {
	Printer           Printer `json:"printer"`
	DeviceID          string  `json:"deviceId"`
	BindingGeneration uint64  `json:"bindingGeneration"`
}
type ConsumerPrintersOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         shared.SuccessEnvelope[[]ConsumerPrinter]
}
type PrinterReportInput struct {
	ConsumerInput
	Body struct {
		PrinterID string `json:"printerId"`
		State     string `json:"state" enum:"ready,unavailable,error,unknown"`
		Reason    string `json:"reason,omitempty" enum:",device_unavailable,device_busy,paper_empty,cover_open,hardware_error,unknown"`
	}
}
type PrinterReportOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         shared.SuccessEnvelope[struct{}]
}

type ListConnectorsInput struct {
	PrinterScope
	Limit  int    `query:"limit" default:"50" minimum:"1" maximum:"100"`
	Cursor string `query:"cursor"`
}
type ConnectorsOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         shared.SuccessEnvelope[[]Connector]
}
type ReviewPairingInput struct {
	Authorization string `header:"Authorization"`
	RequestID     string `header:"X-Request-ID"`
	PairingID     string `path:"pairingId"`
	Body          struct {
		UserCode    string `json:"userCode"`
		TenantID    string `json:"tenantId"`
		InventoryID string `json:"inventoryId"`
	}
}
type PublicPairingCandidate struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AdapterID string `json:"adapterId"`
}
type PairingReview struct {
	Rotation             bool                     `json:"rotation"`
	ID                   string                   `json:"id"`
	Name                 string                   `json:"name"`
	PublicKeyFingerprint string                   `json:"publicKeyFingerprint"`
	Candidates           []PublicPairingCandidate `json:"candidates"`
}
type PairingReviewOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         shared.SuccessEnvelope[PairingReview]
}

type RotateCredentialInput struct {
	ConnectorInput
	Body struct {
		Generation uint64 `json:"generation" minimum:"1"`
		PairingID  string `json:"pairingId"`
		UserCode   string `json:"userCode"`
	}
}
