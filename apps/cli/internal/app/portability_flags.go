package app

import (
	"flag"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func validatePortabilityFlags(o Options, flags *flag.FlagSet) error {
	if !isArchiveCommand(o) && !isImportSource(o) {
		return nil
	}
	action := ""
	if len(o.Command) > 1 {
		action = o.Command[1]
	}
	unsupported := ""
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "server", "tenant", "inventory", "context", "credential-file", "allow-loopback-http", "json", "no-input", "request-id", "color", "help":
		case "input":
			if !isArchiveBody(o) && !isImportSource(o) {
				unsupported = f.Name
			}
		case "yes":
			if !archiveMutation(o) && !isImportSource(o) {
				unsupported = f.Name
			}
		case "limit", "cursor":
			if !isArchiveCommand(o) || action != "list" {
				unsupported = f.Name
			}
		case "output":
			if !isArchiveCommand(o) || action != "download" {
				unsupported = f.Name
			}
		case "file":
			if !isArchiveCommand(o) || action != "upload" {
				unsupported = f.Name
			}
		case "idempotency-key":
			if !isArchiveCommand(o) || (action != "create" && action != "upload") {
				unsupported = f.Name
			}
		default:
			unsupported = f.Name
		}
	})
	if unsupported != "" {
		return ports.Failure("usage", "This portability command does not accept --"+unsupported+". Remove the option.")
	}
	return nil
}
