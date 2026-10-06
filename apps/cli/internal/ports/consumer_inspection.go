package ports

import "context"

type ConsumerInspectionAPI interface {
	ConsumerPrinters(context.Context) (Result[[]ConsumerPrinter], error)
	ConsumerAttempts(context.Context, Page, string, string) (Result[[]ConsumerAttempt], error)
	ConsumerAttempt(context.Context, string) (Result[*ConsumerAttempt], error)
}
type ConsumerPrinter struct {
	BindingGeneration uint64                    `json:"bindingGeneration"`
	DeviceID          string                    `json:"deviceId"`
	Printer           ConsumerRegisteredPrinter `json:"printer"`
}
type ConsumerRegisteredPrinter struct {
	ID               string               `json:"id"`
	Name             string               `json:"name"`
	AdapterID        string               `json:"adapterId"`
	Revision         uint64               `json:"revision"`
	Retired          bool                 `json:"retired"`
	Media            ConsumerPrinterMedia `json:"media"`
	MediaFingerprint string               `json:"mediaFingerprint"`
	Readiness        string               `json:"readiness"`
	ReadinessReason  Optional[string]     `json:"readinessReason,omitempty"`
	ReportedAt       Optional[string]     `json:"reportedAt,omitempty"`
}
type ConsumerPrinterMedia struct {
	ConsumerMedia
	Name string `json:"name"`
}
type ConsumerMedia struct {
	MarginsMicrometers PrinterMargins `json:"marginsMicrometers"`
	DisplayRotation    int64          `json:"displayRotation"`
	PresetID           string         `json:"presetId"`
	Version            uint32         `json:"version"`
	WidthMicrometers   int64          `json:"widthMicrometers"`
	HeightMicrometers  int64          `json:"heightMicrometers"`
	ResolutionDPI      int64          `json:"resolutionDpi"`
	RasterWidth        int64          `json:"rasterWidth"`
	RasterHeight       int64          `json:"rasterHeight"`
	Orientation        string         `json:"orientation"`
	ColorMode          string         `json:"colorMode"`
	CutPolicy          string         `json:"cutPolicy"`
}
type ConsumerAttempt struct {
	ResolvedAt       Optional[string]           `json:"resolvedAt,omitempty"`
	ProtocolVersion  int64                      `json:"protocolVersion"`
	JobID            string                     `json:"jobId"`
	PrinterID        string                     `json:"printerId"`
	AttemptID        string                     `json:"attemptId"`
	SessionID        string                     `json:"sessionId"`
	Status           string                     `json:"status"`
	Revision         uint64                     `json:"revision"`
	Copies           int64                      `json:"copies"`
	LeaseExpiresAt   string                     `json:"leaseExpiresAt"`
	LeaseValid       bool                       `json:"leaseValid"`
	StartedAt        Optional[string]           `json:"startedAt,omitempty"`
	SettledAt        Optional[string]           `json:"settledAt,omitempty"`
	Outcome          ConsumerOutcome            `json:"outcome"`
	MediaFingerprint string                     `json:"mediaFingerprint"`
	Media            Optional[ConsumerMedia]    `json:"media,omitempty"`
	Artifact         Optional[ConsumerArtifact] `json:"artifact,omitempty"`
}
type ConsumerOutcome struct {
	Kind            string `json:"kind"`
	CompletedCopies int64  `json:"completedCopies"`
	Retryable       bool   `json:"retryable"`
	Reason          string `json:"reason"`
}
type ConsumerArtifact struct {
	SHA256       string `json:"sha256"`
	ContentType  string `json:"contentType"`
	ByteLength   int64  `json:"byteLength"`
	WidthPixels  int64  `json:"widthPixels"`
	HeightPixels int64  `json:"heightPixels"`
	ExpiresAt    string `json:"expiresAt"`
}
