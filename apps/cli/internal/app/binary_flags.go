package app

import (
	"flag"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func binaryFlags(flags *flag.FlagSet, o Options) error {
	if !isBinaryCommand(o) {
		return nil
	}
	invalid := ""
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "server", "tenant", "inventory", "context", "credential-file", "allow-loopback-http", "json", "no-input", "request-id", "color", "help":
			return
		case "file", "transfer":
			if o.Command[1] == "upload" {
				return
			}
		case "output":
			if o.Command[1] != "upload" {
				return
			}
		case "variant":
			if o.Command[1] == "thumbnail" {
				return
			}
		case "format":
			if o.Command[0] == "inventories" {
				return
			}
		}
		invalid = f.Name
	})
	if invalid != "" {
		return ports.Failure("usage", "This file command does not accept --"+invalid+". Remove the option.")
	}
	return nil
}
