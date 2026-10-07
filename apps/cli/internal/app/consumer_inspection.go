package app

import (
	"context"
	"flag"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func IsConsumerInspection(o Options) bool {
	return len(o.Command) >= 3 && o.Command[0] == "connectors" && o.Command[1] == "print" && (o.Command[2] == "printers" || o.Command[2] == "attempts")
}
func consumerAttemptsList(o Options) bool {
	return len(o.Command) == 4 && o.Command[2] == "attempts" && o.Command[3] == "list"
}
func validateConsumerInspection(flags *flag.FlagSet, o Options) error {
	if !IsConsumerInspection(o) {
		return nil
	}
	if !(len(o.Command) == 3 && o.Command[2] == "printers") && !consumerAttemptsList(o) && !(len(o.Command) == 5 && o.Command[2] == "attempts" && o.Command[3] == "show" && o.Command[4] != "") {
		return ports.Failure("usage", "Use connectors print printers, connectors print attempts list, or connectors print attempts show ATTEMPT_ID.")
	}
	invalid := ""
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "server", "connector", "tenant", "inventory", "allow-loopback-http", "json", "no-input", "request-id", "color", "help":
		case "printer", "status", "limit", "cursor":
			if !consumerAttemptsList(o) {
				invalid = f.Name
			}
		default:
			invalid = f.Name
		}
	})
	if invalid != "" {
		return ports.Failure("usage", "Connector read commands do not accept --"+invalid+". Remove the option.")
	}
	if o.ConnectorID == "" {
		return ports.Failure("usage", "Select the stored connector with --connector ID or STUFF_STASH_CLI_CONNECTOR_ID.")
	}
	if consumerAttemptsList(o) && (o.Page.Limit < 1 || o.Page.Limit > 100) {
		return ports.Failure("usage", "Supply --limit from 1 through 100.")
	}
	if o.InvitationStatus != "" && o.InvitationStatus != "unsettled" {
		return ports.Failure("usage", "Use --status unsettled for connector attempts.")
	}
	return nil
}

type ConsumerInspector struct {
	Credentials ports.ConnectorCredentials
	API         func(string, string) (ports.ConsumerInspectionAPI, error)
	Clock       ports.Clock
	Output      ports.Output
	Observer    ports.Observer
}

func (r ConsumerInspector) Run(ctx context.Context, o Options) error {
	registration, err := r.Credentials.Load(ctx, o.Server, o.ConnectorID)
	if err != nil {
		return err
	}
	if registration.Server != o.Server || registration.ConnectorID != o.ConnectorID || registration.TenantID == "" || registration.InventoryID == "" || registration.Credential == "" {
		return ports.Failure("authentication", "The stored connector identity does not match. Select the registered server and connector.")
	}
	if (o.Scope.Tenant != "" && o.Scope.Tenant != registration.TenantID) || (o.Scope.Inventory != "" && o.Scope.Inventory != registration.InventoryID) {
		return ports.Failure("usage", "Household or inventory scope conflicts with the stored connector. Remove the scope override or select the matching connector.")
	}
	if !registration.ExpiresAt.After(r.Clock.Now()) {
		return ports.Failure("authentication", "The connector credential expired. Pair this connector again.")
	}
	api, err := r.API(registration.Server, registration.Credential)
	if err != nil {
		return err
	}
	active, cancel := context.WithDeadline(ctx, registration.ExpiresAt)
	defer cancel()
	var result any
	switch {
	case o.Command[2] == "printers":
		result, err = api.ConsumerPrinters(active)
	case consumerAttemptsList(o):
		status := o.InvitationStatus
		if status == "" {
			status = "unsettled"
		}
		result, err = api.ConsumerAttempts(active, o.Page, o.PrinterID, status)
	default:
		result, err = api.ConsumerAttempt(active, o.Command[4])
	}
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.print_consumer.read.completed")
	return r.Output.Result(result)
}
