package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
	"github.com/stuffstash/stuff-stash/cli/internal/app/scopeselection"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (r Runner) chooseMissingScope(ctx context.Context, o Options, session ports.Session) (Options, error) {
	if !missingResourceScope(o) || r.Picker == nil || o.JSON || o.NoInput {
		return o, nil
	}
	if r.ScopeAPI == nil || r.Contexts == nil {
		return o, ports.Failure("configuration", "Scope selection is not available. Set the context file path and supply scope options.")
	}
	principal := contexts.Principal(session)
	if principal == "" {
		return o, ports.Failure("authentication", "The CLI cannot verify the account for this selection. Sign in again.")
	}
	api, err := r.ScopeAPI(o.Server, session.IDToken)
	if err != nil {
		return o, err
	}
	inventoryRequired := requiresInventory(o)
	selected, err := scopeselection.Choose(ctx, api, r.Picker, o.Scope, inventoryRequired)
	if err != nil {
		return o, err
	}
	config, err := r.Contexts.Load(ctx)
	if err != nil {
		return o, err
	}
	server := contexts.ServerKey(o.Server)
	name := server
	preferred := o.Selection.Context
	if preferred == "" {
		preferred = config.Current
	}
	for _, entry := range config.Contexts {
		if contexts.ServerKey(entry.Server) == server && entry.Principal == principal {
			if name == server {
				name = entry.Name
			}
			if entry.Name == preferred {
				name = entry.Name
				break
			}
		}
	}

	for _, entry := range config.Contexts {
		if entry.Name == name && (contexts.ServerKey(entry.Server) != contexts.ServerKey(o.Server) || entry.Principal != "" && entry.Principal != principal) {
			name = contexts.ServerKey(o.Server) + "#" + principal
			break
		}
	}
	entry := contexts.Entry{Name: name, Server: o.Server, Principal: principal, Tenant: selected.Tenant, Inventory: selected.Inventory}
	if err := (contexts.Manager{Store: r.Contexts}).Remember(ctx, entry); err != nil {
		return o, err
	}
	o.Scope = selected
	return o, nil
}
