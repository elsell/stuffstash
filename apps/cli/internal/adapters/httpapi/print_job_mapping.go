package httpapi

import (
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"time"
)

func proof(control printing.AttemptControl) generated.PrintClaimProof {
	return generated.PrintClaimProof{SessionId: control.SessionID, ClaimToken: control.ClaimToken, Revision: int64(control.Revision)}
}
func consumerMedia(value generated.PrintConsumerMedia) printing.Media {
	return printing.Media{PresetID: value.PresetId, Version: uint32(value.Version), WidthMicrometers: int(value.WidthMicrometers), HeightMicrometers: int(value.HeightMicrometers), Margins: printing.Margins{Left: int(value.MarginsMicrometers.Left), Right: int(value.MarginsMicrometers.Right), Top: int(value.MarginsMicrometers.Top), Bottom: int(value.MarginsMicrometers.Bottom)}, ResolutionDPI: int(value.ResolutionDpi), RasterWidth: int(value.RasterWidth), RasterHeight: int(value.RasterHeight), Orientation: value.Orientation, ColorMode: value.ColorMode, CutPolicy: value.CutPolicy, DisplayRotation: int(value.DisplayRotation)}
}
func attemptStatus(value generated.PrintConsumerAttempt) (printing.AttemptStatus, error) {
	if value.Revision <= 0 || value.AttemptId == "" || value.SessionId == "" || value.JobId == "" {
		return printing.AttemptStatus{}, ports.Failure("protocol", "invalid print attempt response")
	}
	result := printing.AttemptStatus{AttemptID: value.AttemptId, SessionID: value.SessionId, JobID: value.JobId, Revision: uint64(value.Revision), LeaseExpiresAt: value.LeaseExpiresAt, Outcome: printing.Outcome(value.Outcome.Kind), CompletedCopies: int(value.Outcome.CompletedCopies)}
	// Attempt settlement takes precedence over the parent job, which may already
	// be queued or claimed again after a proven no-output attempt.
	if value.SettledAt != nil {
		switch result.Outcome {
		case printing.Completed:
			result.Phase = printing.RemoteCompleted
		case printing.NoOutput:
			result.Phase = printing.RemoteFailed
		case printing.Uncertain:
			result.Phase = printing.RemoteUncertain
		default:
			return printing.AttemptStatus{}, ports.Failure("protocol", "invalid settled print outcome")
		}
		return result, nil
	}
	switch value.Status {
	case "claimed":
		result.Phase = printing.RemoteClaimed
	case "printing":
		result.Phase = printing.RemotePrinting
	case "uncertain":
		result.Phase = printing.RemoteUncertain
	default:
		return printing.AttemptStatus{}, ports.Failure("protocol", "unrecognized active print attempt")
	}
	if !value.LeaseValid {
		result.LeaseExpiresAt = time.Time{}
	}
	return result, nil
}
func outcome(evidence printing.Evidence) generated.PrintOutcome {
	result := generated.PrintOutcome{Kind: generated.PrintOutcomeKind(evidence.Outcome), CompletedCopies: int64(evidence.CompletedCopies)}
	switch evidence.Reason {
	case printing.NoReason:
		result.Reason = ""
	case printing.InvalidArtifact:
		result.Reason = "invalid_artifact"
	case printing.Disconnected, printing.PermissionDenied, printing.NoMedia, printing.CoverOpen:
		result.Reason = "device_unavailable"
	case printing.CutterJam, printing.MediaError, printing.DeviceError:
		result.Reason = "device_failure"
	default:
		result.Reason = "unknown"
	}
	if evidence.Outcome == printing.Uncertain && evidence.CompletedCopies > 0 {
		result.Reason = "partial_output"
	}
	result.Retryable = evidence.Outcome == printing.NoOutput && result.Reason != "invalid_artifact"
	return result
}
