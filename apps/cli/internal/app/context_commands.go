package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (r Runner) contextCommand(ctx context.Context, command []string) error {
	if r.Contexts == nil {
		return ports.Failure("configuration", "Context storage is not available. Examine the CLI configuration directory.")
	}
	manager := contexts.Manager{Store: r.Contexts}
	if len(command) == 2 {
		switch command[1] {
		case "list":
			entries, err := manager.List(ctx)
			if err != nil {
				return err
			}
			return r.Output.Result(entries)
		case "current":
			entry, err := manager.Current(ctx)
			if err != nil {
				return err
			}
			return r.Output.Result(entry)
		}
	}
	if len(command) == 3 {
		switch command[1] {
		case "use":
			if err := manager.Use(ctx, command[2]); err != nil {
				return err
			}
			return r.Output.Result(map[string]string{"context": command[2], "status": "selected"})
		case "delete":
			if err := manager.Delete(ctx, command[2]); err != nil {
				return err
			}
			return r.Output.Result(map[string]string{"context": command[2], "status": "deleted"})
		}
	}
	return ports.Failure("usage", "Use context list, context current, context use NAME, or context delete NAME.")
}
