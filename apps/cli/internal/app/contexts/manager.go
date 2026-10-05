package contexts

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type Manager struct{ Store Store }

func (m Manager) List(ctx context.Context) ([]Entry, error) {
	config, err := m.Store.Load(ctx)
	if err != nil {
		return nil, err
	}
	entries := append([]Entry{}, config.Contexts...)
	return entries, nil
}
func (m Manager) Current(ctx context.Context) (Entry, error) {
	config, err := m.Store.Load(ctx)
	if err != nil {
		return Entry{}, err
	}
	for _, entry := range config.Contexts {
		if entry.Name == config.Current {
			return entry, nil
		}
	}
	return Entry{}, ports.Failure("configuration", "No context is selected. Run stuffstash context use NAME.")
}
func (m Manager) Use(ctx context.Context, name string) error {
	return m.Store.Update(ctx, func(config *Config) error {
		for _, entry := range config.Contexts {
			if entry.Name == name {
				config.Current = name
				return nil
			}
		}
		return missingContext()
	})
}
func (m Manager) Delete(ctx context.Context, name string) error {
	return m.Store.Update(ctx, func(config *Config) error {
		for i, entry := range config.Contexts {
			if entry.Name == name {
				config.Contexts = append(config.Contexts[:i], config.Contexts[i+1:]...)
				if config.Current == name {
					config.Current = ""
				}
				return nil
			}
		}
		return missingContext()
	})
}
func (m Manager) Remember(ctx context.Context, entry Entry) error {
	entry.Server = ServerKey(entry.Server)
	if entry.Principal == "" {
		return ports.Failure("authentication", "Cannot save this selection without a verified account. Log in again.")
	}
	if err := Validate(Config{Version: Version, Current: entry.Name, Contexts: []Entry{entry}}); err != nil {
		return err
	}
	return m.Store.Update(ctx, func(config *Config) error {
		for i, saved := range config.Contexts {
			if saved.Name == entry.Name {
				if ServerKey(saved.Server) != ServerKey(entry.Server) || saved.Principal != "" && saved.Principal != entry.Principal {
					return ports.Failure("configuration", "This context belongs to another server or account. Use a different context name.")
				}
				config.Contexts[i] = entry
				config.Current = entry.Name
				return nil
			}
		}
		config.Contexts = append(config.Contexts, entry)
		config.Current = entry.Name
		return nil
	})
}
func (m Manager) ClearServer(ctx context.Context, server string) error {
	return m.Store.Update(ctx, func(config *Config) error {
		for i := range config.Contexts {
			entry := &config.Contexts[i]
			if ServerKey(entry.Server) == ServerKey(server) {
				entry.Principal = ""
				entry.Tenant = ""
				entry.Inventory = ""
			}
		}
		return nil
	})
}
func missingContext() error {
	return ports.Failure("configuration", "The context does not exist. Use stuffstash context list to see the available contexts.")
}
