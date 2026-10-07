package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"math"
)

func (c *Client) Heartbeat(ctx context.Context, session string, report *printing.ConnectorReport) error {
	value, err := connectorReport(report)
	if err != nil {
		return err
	}
	body, err := workerRequest(workerHeartbeat{SessionID: session, Report: value})
	if err != nil {
		return err
	}
	result, err := read[ports.Result[ports.PrintConnector]](c.sdk.PostPrintConsumerHeartbeatWithBody(ctx, nil, "application/json", body))
	if err == nil {
		c.recordReceipt(ctx, "heartbeat", ports.WorkerConnectorReceipt(result))
	}
	return consumerError(err)
}
func (c *Client) Printers(ctx context.Context) ([]printing.RegisteredPrinter, error) {
	result, err := c.ConsumerPrinters(ctx)
	if err != nil {
		return nil, consumerError(err)
	}
	printers := make([]printing.RegisteredPrinter, 0, len(result.Data))
	for _, value := range result.Data {
		if value.BindingGeneration <= 0 || value.Printer.ID == "" || value.DeviceID == "" {
			return nil, ports.Failure("protocol", "The server returned invalid printer details. Check the printer settings with the server administrator.")
		}
		printers = append(printers, printing.RegisteredPrinter{ID: value.Printer.ID, AdapterID: value.Printer.AdapterID, DeviceID: value.DeviceID, MediaFingerprint: value.Printer.MediaFingerprint, BindingGeneration: value.BindingGeneration, Retired: value.Printer.Retired, Media: consumerMedia(value.Printer.Media.ConsumerMedia)})
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
	result, err := read[ports.Result[struct{}]](c.sdk.PostPrintConsumerPrinterReports(ctx, nil, generated.PrinterReportInputBody{PrinterId: id, State: state, Reason: &reason}))
	if err == nil {
		c.recordReceipt(ctx, "printer-report", ports.WorkerAcknowledgementReceipt(result))
	}
	return consumerError(err)
}

func connectorReport(report *printing.ConnectorReport) (*ports.PrintConnectorReport, error) {
	if report == nil {
		return nil, nil
	}
	adapters := []ports.PrintConnectorAdapter{}
	for _, a := range report.Adapters {
		versions := []uint32{}
		media := []ports.PrintConnectorMedia{}
		for _, v := range a.ContractVersions {
			if v < 0 || uint64(v) > math.MaxUint32 {
				return nil, ports.Failure("configuration", "The print adapter version is outside the supported range. Contact the connector administrator.")
			}
			versions = append(versions, uint32(v))
		}
		for _, m := range a.Media {
			media = append(media, ports.PrintConnectorMedia{ID: m.PresetID, Version: m.Version})
		}
		adapters = append(adapters, ports.PrintConnectorAdapter{ID: a.ID, Formats: a.Formats, CompletionEvidence: a.CompletionEvidence, Wake: a.Wake, ContractVersions: versions, Media: media})
	}
	return &ports.PrintConnectorReport{Version: report.Version, Commit: report.Commit, Platform: report.Platform, Architecture: report.Architecture, Adapters: adapters}, nil
}
