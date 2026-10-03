package printregistry

import (
	"context"
	"math"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s Service) PrintSettings(ctx context.Context, actor Actor) (printing.InventoryPrintSettings, error) {
	if err := s.access(ctx, actor, ports.InventoryPermissionView); err != nil {
		return printing.InventoryPrintSettings{}, err
	}
	settings, err := s.Settings.GetPrintSettings(ctx, actor.Scope)
	if err != nil {
		return printing.InventoryPrintSettings{}, registryError(err)
	}
	record, err := s.auditRecord(actor, audit.ActionPrintSettingsViewed, string(settings.DefaultPrinterID))
	if err != nil {
		return printing.InventoryPrintSettings{}, err
	}
	if err = s.Audit.SaveAuditRecord(ctx, record); err != nil {
		return printing.InventoryPrintSettings{}, err
	}
	return settings, nil
}
func (s Service) ReplacePrintSettings(ctx context.Context, actor Actor, settings printing.InventoryPrintSettings) (printing.InventoryPrintSettings, error) {
	if err := s.access(ctx, actor, ports.InventoryPermissionConfigure); err != nil {
		return printing.InventoryPrintSettings{}, err
	}
	if settings.Revision == math.MaxUint64 || (settings.PrintOnCreateDefault && settings.DefaultPrinterID == "") {
		return printing.InventoryPrintSettings{}, apperrors.ErrInvalidInput
	}
	settings.Scope = actor.Scope
	var destination *printing.SettingsDestination
	var media *printing.MediaSnapshot
	if settings.DefaultPrinterID != "" {
		printer, err := s.Printers.GetPrinter(ctx, actor.Scope, settings.DefaultPrinterID)
		if err != nil {
			return printing.InventoryPrintSettings{}, registryError(err)
		}
		if printer.Retired {
			return printing.InventoryPrintSettings{}, apperrors.ErrInvalidInput
		}
		destination = &printing.SettingsDestination{ID: printer.ID, Revision: printer.Revision, MediaFingerprint: printer.MediaFingerprint}
		media = &printer.Media
	}
	if err := s.SelectionValidator.ValidateDefaultLabelSelection(ctx, settings.Template, media); err != nil {
		return printing.InventoryPrintSettings{}, err
	}
	// Permission can change during CPU-bound compatibility rendering.
	if err := s.access(ctx, actor, ports.InventoryPermissionConfigure); err != nil {
		return printing.InventoryPrintSettings{}, err
	}
	expected := settings.Revision
	settings.Revision++
	settings.UpdatedAt = s.Clock.Now()
	record, err := s.auditRecord(actor, audit.ActionPrintSettingsUpdated, string(settings.DefaultPrinterID))
	if err != nil {
		return printing.InventoryPrintSettings{}, err
	}
	result, err := s.Settings.SavePrintSettings(ctx, settings, expected, destination, record)
	if err != nil {
		return printing.InventoryPrintSettings{}, registryError(err)
	}
	s.observe(ctx, ports.EventPrintSettingsUpdated)
	return result, nil
}
