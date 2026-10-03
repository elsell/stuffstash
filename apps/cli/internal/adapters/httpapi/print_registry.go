package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) Heartbeat(ctx context.Context, session string) error {
	_, err := read[generated.SuccessEnvelopeConnector](c.sdk.PostPrintConsumerHeartbeat(ctx, nil, generated.HeartbeatInputBody{SessionId: session}))
	return consumerError(err)
}
func (c *Client) Printers(ctx context.Context) ([]printing.RegisteredPrinter, error) {
	result, err := read[generated.SuccessEnvelopeListConsumerPrinter](c.sdk.GetPrintConsumerPrinters(ctx, nil))
	if err != nil {
		return nil, consumerError(err)
	}
	printers := make([]printing.RegisteredPrinter, 0, len(result.Data.GetOrEmpty()))
	for _, value := range result.Data.GetOrEmpty() {
		if value.BindingGeneration <= 0 || value.Printer.Id == "" || value.DeviceId == "" {
			return nil, ports.Failure("protocol", "invalid registered printer response")
		}
		media := value.Printer.Media
		printers = append(printers, printing.RegisteredPrinter{ID: value.Printer.Id, AdapterID: value.Printer.AdapterId, DeviceID: value.DeviceId, MediaFingerprint: value.Printer.MediaFingerprint, BindingGeneration: uint64(value.BindingGeneration), Retired: value.Printer.Retired, Media: printing.Media{PresetID: media.PresetId, Version: uint32(media.Version), WidthMicrometers: int(media.WidthMicrometers), HeightMicrometers: int(media.HeightMicrometers), Margins: printing.Margins{Left: int(media.MarginsMicrometers.Left), Right: int(media.MarginsMicrometers.Right), Top: int(media.MarginsMicrometers.Top), Bottom: int(media.MarginsMicrometers.Bottom)}, ResolutionDPI: int(media.ResolutionDpi), RasterWidth: int(media.RasterWidth), RasterHeight: int(media.RasterHeight), Orientation: media.Orientation, ColorMode: media.ColorMode, CutPolicy: media.CutPolicy, DisplayRotation: int(media.DisplayRotation)}})
	}
	return printers, nil
}
func (c *Client) Report(ctx context.Context, id string, readiness printing.Readiness) error {
	state := generated.PrinterReportInputBodyState("unknown")
	switch readiness.State {
	case printing.Ready:
		state = "ready"
	case printing.Unavailable:
		state = "unavailable"
	case printing.Busy, printing.NeedsAttention:
		state = "error"
	}
	reason := generated.PrinterReportInputBodyReason("")
	switch readiness.Reason {
	case printing.Disconnected, printing.PermissionDenied, printing.UnsupportedTransport:
		reason = "device_unavailable"
	case printing.ActiveSubmission:
		reason = "device_busy"
	case printing.NoMedia:
		reason = "paper_empty"
	case printing.CoverOpen:
		reason = "cover_open"
	case printing.CutterJam, printing.MediaError, printing.DeviceError:
		reason = "hardware_error"
	default:
		if readiness.Reason != printing.NoReason {
			reason = "unknown"
		}
	}
	_, err := read[generated.SuccessEnvelopeStruct](c.sdk.PostPrintConsumerPrinterReports(ctx, nil, generated.PrinterReportInputBody{PrinterId: id, State: state, Reason: &reason}))
	return consumerError(err)
}
