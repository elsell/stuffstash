package printregistry

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ConsumerPrinter struct {
	Printer           printing.Printer
	DeviceID          string
	BindingGeneration uint64
}

func (s ConnectorService) ConsumerPrinters(ctx context.Context, c printing.Connector) ([]ConsumerPrinter, error) {
	if c.State != printing.ConnectorActive {
		return nil, apperrors.ErrUnauthorized
	}
	registration, err := s.Repository.GetPrintConnector(ctx, c.Scope, c.ID)
	if err != nil {
		return nil, connectorError(err)
	}
	result := []ConsumerPrinter{}
	for _, binding := range registration.Bindings {
		if binding.Revoked {
			continue
		}
		authority, err := s.AuthorizePrinter(ctx, c, binding.PrinterID, ports.PrinterPermissionViewConsumer)
		if errors.Is(err, ports.ErrForbidden) {
			continue
		}
		if err != nil {
			return nil, err
		}
		p, err := s.Registry.Printers.GetPrinter(ctx, c.Scope, binding.PrinterID)
		if err != nil {
			return nil, registryError(err)
		}
		result = append(result, ConsumerPrinter{Printer: p, DeviceID: binding.DeviceID, BindingGeneration: authority.BindingGeneration})
	}
	return result, nil
}
func (s ConnectorService) ReportPrinter(ctx context.Context, c printing.Connector, id printing.PrinterID, state printing.PrinterReadiness, reason string) error {
	switch state {
	case printing.PrinterReady, printing.PrinterUnknown, printing.PrinterUnavailable, printing.PrinterError:
	default:
		return apperrors.ErrInvalidInput
	}
	switch reason {
	case "", "device_unavailable", "device_busy", "paper_empty", "cover_open", "hardware_error", "unknown":
	default:
		return apperrors.ErrInvalidInput
	}
	if state == printing.PrinterReady && reason != "" {
		return apperrors.ErrInvalidInput
	}
	authority, err := s.AuthorizePrinter(ctx, c, id, ports.PrinterPermissionReport)
	if err != nil {
		return err
	}
	return connectorError(s.Repository.ReportPrintPrinter(ctx, authority, printing.PrinterReport{Scope: c.Scope, ConnectorID: c.ID, PrinterID: id, State: state, Reason: reason}, s.Registry.Clock.Now()))
}
