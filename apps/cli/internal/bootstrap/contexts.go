package bootstrap

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/contextfile"
	"github.com/stuffstash/stuff-stash/cli/internal/app"
	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"os"
	"path/filepath"
)

func configuredContexts(getenv func(string) string, required bool) (contexts.Store, error) {
	path := getenv("STUFF_STASH_CLI_CONFIG_FILE")
	if path == "" {
		directory, err := os.UserConfigDir()
		if err != nil {
			if !required {
				return nil, nil
			}
			return nil, ports.Failure("configuration", "The CLI cannot find the configuration directory. Set STUFF_STASH_CLI_CONFIG_FILE to a private file path.")
		}
		path = filepath.Join(directory, "stuffstash", "contexts.json")
	}
	return contextfile.Store{Path: path}, nil
}
func contextServer(ctx context.Context, store contexts.Store, o app.Options) (string, error) {
	if store == nil {
		return o.Server, nil
	}
	if len(o.Command) == 1 && o.Command[0] == "logout" && o.Server != "" {
		return o.Server, nil
	}
	config, err := store.Load(ctx)
	if err != nil {
		return "", err
	}
	// Before authentication, only choose the endpoint. Account-bound resource
	// choices are resolved by the application after loading verified credentials.
	selection := o.Selection
	selection.Server = ""
	selection.Tenant = ""
	selection.Inventory = ""
	resolved, err := contexts.Resolve(config, selection, "")
	if err != nil {
		return "", err
	}
	if o.Server != "" {
		return o.Server, nil
	}
	return resolved.Server, nil
}
