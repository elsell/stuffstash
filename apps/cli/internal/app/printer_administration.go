package app

import (
	"context"
	"errors"
	"flag"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func isPrinterCreation(o Options) bool {
	return len(o.Command) == 2 && o.Command[0] == "printers" && o.Command[1] == "create"
}
func isPrinterAdministration(o Options) bool {
	c := o.Command
	return len(c) >= 2 && (c[0] == "printers" && (c[1] == "create" || c[1] == "update") || c[0] == "print-settings" && c[1] == "update") || len(c) >= 3 && c[0] == "connectors" && c[1] == "print" && c[2] == "update"
}
func validatePrinterAdministration(o Options, scope bool) error {
	c := o.Command
	valid := isPrinterCreation(o) || len(c) == 3 && c[0] == "printers" && c[1] == "update" && c[2] != "" || len(c) == 4 && c[0] == "connectors" && c[1] == "print" && c[2] == "update" && c[3] != "" || len(c) == 2 && c[0] == "print-settings" && c[1] == "update"
	if !valid {
		return ports.Failure("usage", "Use printers create, printers update PRINTER_ID, connectors print update CONNECTOR_ID, or print-settings update.")
	}
	if scope && missingResourceScope(o) {
		return ports.Failure("usage", "Supply --tenant and --inventory, or select a saved inventory context.")
	}
	return nil
}
func validatePrinterAdministrationFlags(o Options, flags *flag.FlagSet) error {
	unsupported, mixed := "", false
	flags.Visit(func(f *flag.Flag) {
		if !isPrinterAdministration(o) {
			if f.Name == "adapter" || f.Name == "preset-version" {
				unsupported = f.Name
			}
			return
		}
		switch f.Name {
		case "server", "tenant", "inventory", "context", "json", "no-input", "request-id", "color", "help", "allow-loopback-http", "credential-file", "input", "yes":
		case "idempotency-key":
			if !isPrinterCreation(o) {
				unsupported = f.Name
			}
		case "name", "adapter", "label-size", "preset-version":
			if !isPrinterCreation(o) {
				unsupported = f.Name
			}
			if o.InputPath != "" {
				mixed = true
			}
		default:
			unsupported = f.Name
		}
	})
	if unsupported != "" {
		return ports.Failure("usage", "This command does not accept --"+unsupported+". Remove the option.")
	}
	if mixed {
		return ports.Failure("usage", "Use either printer field options or --input. Do not combine them.")
	}
	return nil
}
func (r Runner) printerAdministration(ctx context.Context, o Options, token string) error {
	if r.PrinterAdministrationAPI == nil {
		return ports.Failure("configuration", "Printer administration is not available. Update the CLI.")
	}
	api, err := r.PrinterAdministrationAPI(o.Server, token)
	if err != nil {
		return err
	}
	action, target, effect, inspect := "create", "new printer", "Register a printer using the selected adapter and media preset.", "printers list"
	if o.Command[0] == "printers" && !isPrinterCreation(o) {
		action = "update"
		target = o.Command[2]
		effect = "Update only the supplied printer fields using its current revision."
		inspect = "printers show PRINTER_ID"
	}
	if o.Command[0] == "connectors" {
		action = "connector.update"
		target = o.Command[3]
		effect = "Change supplied connector fields. Printer bindings and revocation can change connector access."
		inspect = "connectors print show CONNECTOR_ID"
	}
	if o.Command[0] == "print-settings" {
		action = "settings.update"
		target = "inventory defaults"
		effect = "Replace print settings. These defaults can change automatic label printing."
		inspect = "print-settings show"
	}
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant) + ". Inventory: " + strconv.Quote(o.Scope.Inventory) + ". Target: " + strconv.Quote(target) + ". " + effect); err != nil {
		return err
	}
	if err := r.confirmAction(ctx, o, "Printer administration", "Apply changes", effect); err != nil {
		return err
	}
	var result any
	switch action {
	case "create":
		result, err = api.CreatePrinter(ctx, o.Scope, o.IdempotencyKey, o.RequestBody)
	case "update":
		result, err = api.UpdatePrinter(ctx, o.Scope, target, o.RequestBody)
	case "connector.update":
		result, err = api.UpdatePrintConnector(ctx, o.Scope, target, o.RequestBody)
	case "settings.update":
		result, err = api.UpdatePrintSettings(ctx, o.Scope, o.RequestBody)
	}
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var failure *ports.Error
		if errors.As(err, &failure) {
			switch failure.Category {
			case "conflict":
				return ports.Failure("conflict", "The server rejected the change. Run "+inspect+" before you try again. Examine the server state. Do not overwrite another user's change.")
			case "network", "protocol", "unavailable", "api":
				advice := "The result is unknown. Run " + inspect + " before you try again."
				if isPrinterCreation(o) {
					advice += " Reuse the same idempotency key and unchanged request."
				}
				return ports.Failure(failure.Category, advice)
			}
		}
		return err
	}
	r.Observer.Event(ctx, "cli.printer_administration."+action+".completed")
	return r.Output.Result(result)
}
