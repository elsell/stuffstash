package app

import (
	"context"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func selectPrinter(ctx context.Context, api ports.PrintSelectionSource, o Options) (ports.LabelPrintSelection, error) {
	defaults, err := api.PrintDefaults(ctx, o.Scope)
	if err != nil {
		return ports.LabelPrintSelection{}, err
	}
	selection := ports.LabelPrintSelection{PrinterID: defaults.PrinterID, TemplateID: defaults.TemplateID, TemplateVersion: defaults.TemplateVersion, ShowReference: defaults.ShowReference, Copies: o.Copies}
	if o.PrinterID != "" {
		selection.PrinterID = o.PrinterID
	}
	if selection.PrinterID == "" {
		return selection, ports.Failure("usage", "choose --printer or configure an inventory default printer")
	}
	if o.TemplateID != "" {
		selection.TemplateID = o.TemplateID
	}
	if o.TemplateVersion != 0 {
		selection.TemplateVersion = uint32(o.TemplateVersion)
	}
	if o.ShowReferenceSet {
		selection.ShowReference = o.ShowReference
	}
	printer, err := api.RegisteredPrinter(ctx, o.Scope, selection.PrinterID)
	if err != nil {
		return selection, err
	}
	if printer.Retired {
		return selection, ports.Failure("conflict", "printer is retired; choose another destination")
	}
	selection.ExpectedMediaFingerprint = printer.MediaFingerprint
	return selection, nil
}
func isPrintingCommand(o Options) bool {
	return o.PrintLabel || (len(o.Command) > 0 && (o.Command[0] == "labels" || o.Command[0] == "printers" || o.Command[0] == "print-jobs"))
}
func validatePrintingCommand(o Options) error { return validatePrintingCommandOptions(o, true) }
func validatePrintingCommandOptions(o Options, requireScope bool) error {
	if len(o.Command) < 2 || requireScope && (o.Scope.Tenant == "" || o.Scope.Inventory == "") {
		return ports.Failure("usage", "printing requires --tenant and --inventory")
	}
	n := len(o.Command)
	switch o.Command[0] + " " + o.Command[1] {
	case "printers configure":
		if n == 3 && o.LabelSize != "" {
			return nil
		}
	case "printers profiles", "printers list", "print-jobs list":
		if n == 2 {
			return nil
		}
	case "print-jobs resolve", "printers show", "printers test", "labels print", "print-jobs show", "print-jobs cancel", "print-jobs reprint":
		if n == 3 {
			return nil
		}
	}
	return ports.Failure("usage", "invalid printing command; use --help")
}
func executePrinting(ctx context.Context, api ports.HumanPrintingAPI, o Options) (any, error) {
	switch o.Command[0] + " " + o.Command[1] {
	case "printers show":
		return api.Printer(ctx, o.Scope, o.Command[2])
	case "printers configure":
		return configurePrinter(ctx, api, o)
	case "printers profiles":
		return api.PrinterProfiles(ctx, o.Scope)
	case "printers list":
		return api.RegisteredPrinters(ctx, o.Scope, o.Page)
	case "print-jobs list":
		return api.PrintJobs(ctx, o.Scope, o.Page, o.PrinterID)
	case "print-jobs show":
		return api.PrintJob(ctx, o.Scope, o.Command[2])
	case "print-jobs cancel":
		job, err := api.PrintJob(ctx, o.Scope, o.Command[2])
		if err != nil {
			return nil, err
		}
		return api.CancelPrint(ctx, o.Scope, o.Command[2], job.Data.Revision)
	}
	if o.Command[0] == "printers" {
		o.PrinterID = o.Command[2]
		o.Copies = 1
	}
	selection, err := selectPrinter(ctx, api, o)
	if err != nil {
		return nil, err
	}
	switch o.Command[0] + " " + o.Command[1] {
	case "labels print":
		return api.QueueLabel(ctx, o.Scope, o.Command[2], selection, o.IdempotencyKey)
	case "printers test":
		return api.TestPrinter(ctx, o.Scope, selection, o.IdempotencyKey)
	case "print-jobs reprint":
		return api.Reprint(ctx, o.Scope, o.Command[2], selection, o.IdempotencyKey)
	}
	return nil, ports.Failure("usage", "unknown printing action")
}
