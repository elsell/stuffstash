package printing

import "fmt"

type State string

const (
	Ready          State = "ready"
	Unavailable    State = "unavailable"
	NeedsAttention State = "needs_attention"
	Busy           State = "busy"
	Unknown        State = "unknown"
)

type Outcome string

const (
	Pending   Outcome = "pending"
	Completed Outcome = "completed"
	NoOutput  Outcome = "no_output"
	Uncertain Outcome = "uncertain"
)

type Reason string

const (
	NoReason               Reason = ""
	Disconnected           Reason = "disconnected"
	PermissionDenied       Reason = "permission_denied"
	UnsupportedTransport   Reason = "unsupported_transport"
	NoMedia                Reason = "no_media"
	CoverOpen              Reason = "cover_open"
	CutterJam              Reason = "cutter_jam"
	MediaError             Reason = "media_error"
	DeviceError            Reason = "device_error"
	InvalidArtifact        Reason = "invalid_artifact"
	InvalidStatus          Reason = "invalid_status"
	Interrupted            Reason = "interrupted"
	ActiveSubmission       Reason = "active_submission"
	UnrecognizedSubmission Reason = "unrecognized_submission"
)

type Margins struct {
	Left   int `json:"left"`
	Right  int `json:"right"`
	Top    int `json:"top"`
	Bottom int `json:"bottom"`
}
type Media struct {
	PresetID          string  `json:"preset_id"`
	Version           uint32  `json:"version"`
	WidthMicrometers  int     `json:"width_micrometers"`
	HeightMicrometers int     `json:"height_micrometers"`
	Margins           Margins `json:"margins_micrometers"`
	ResolutionDPI     int     `json:"resolution_dpi"`
	RasterWidth       int     `json:"raster_width"`
	RasterHeight      int     `json:"raster_height"`
	Orientation       string  `json:"orientation"`
	ColorMode         string  `json:"color_mode"`
	CutPolicy         string  `json:"cut_policy"`
	DisplayRotation   int     `json:"display_rotation"`
}
type Descriptor struct {
	ID                 string   `json:"id"`
	Model              string   `json:"model"`
	Platforms          []string `json:"platforms"`
	Transport          string   `json:"transport"`
	ContractVersions   []int    `json:"contract_versions"`
	Formats            []string `json:"formats"`
	Media              []Media  `json:"media_presets"`
	CompletionEvidence string   `json:"completion_evidence"`
	Wake               bool     `json:"wake"`
	PhysicallyVerified bool     `json:"physically_verified"`
}
type Device struct {
	ID           string `json:"id"`
	Model        string `json:"model"`
	Serial       string `json:"serial,omitempty"`
	PhysicalPort string `json:"physical_port"`
	Path         string `json:"path,omitempty"`
	State        State  `json:"state"`
	Reason       Reason `json:"reason,omitempty"`
}
type Readiness struct {
	State  State  `json:"state"`
	Reason Reason `json:"reason,omitempty"`
}

// Label is one already-rendered copy. The worker validates source/digest and
// journals intent before calling Submit under its physical-device lock.
type Label struct {
	AttemptID string
	Copy      int
	PNG       []byte
	Media     Media
}
type Submission struct {
	Reference string
	AttemptID string
	Copy      int
}
type Observation struct {
	Outcome Outcome
	Reason  Reason
}
type SubmissionError struct {
	Outcome Outcome
	Reason  Reason
}

func (e *SubmissionError) Error() string {
	return fmt.Sprintf("printer submission %s: %s", e.Outcome, e.Reason)
}
