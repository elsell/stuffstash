package ports

import "context"

type PairingReview struct {
	ID                   string                   `json:"id"`
	Name                 string                   `json:"name"`
	Rotation             bool                     `json:"rotation"`
	PublicKeyFingerprint string                   `json:"publicKeyFingerprint"`
	Candidates           []PublicPairingCandidate `json:"candidates"`
}
type PublicPairingCandidate struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AdapterID string `json:"adapterId"`
}
type PairingBinding struct {
	CandidateID string `json:"candidateId"`
	PrinterID   string `json:"printerId"`
}

// ApprovedConnector contains only the public approval response, never credentials.
type ApprovedConnector struct {
	PrintConnector
	LastSeenAt       Optional[string]               `json:"lastSeenAt,omitempty"`
	ReportReceivedAt Optional[string]               `json:"reportReceivedAt,omitempty"`
	Report           Optional[PrintConnectorReport] `json:"report,omitempty"`
}

type ApprovedPairing struct {
	ID        string `json:"id"`
	State     string `json:"state"`
	ExpiresAt string `json:"expiresAt"`
}

type PairingApprovalAPI interface {
	ReviewPrintPairing(context.Context, string, []byte) (Result[PairingReview], error)
	ApprovePrintPairing(context.Context, string, []byte) (Result[ApprovedConnector], error)
	ApproveConnectorRotation(context.Context, Scope, string, []byte) (Result[ApprovedPairing], error)
}
