// Package contexts resolves local CLI choices; it never grants server authority.
package contexts

import (
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strings"
)

const Version = 1

type Entry struct {
	Name      string `json:"name"`
	Server    string `json:"server"`
	Principal string `json:"principal,omitempty"`
	Tenant    string `json:"tenant,omitempty"`
	Inventory string `json:"inventory,omitempty"`
}
type Config struct {
	Version  int     `json:"version"`
	Current  string  `json:"current,omitempty"`
	Contexts []Entry `json:"contexts"`
}
type Selection struct{ Context, Server, Tenant, Inventory string }
type Resolved struct {
	Context, Server string
	Scope           ports.Scope
}

// Overlay keeps dependent scope from crossing an explicit boundary change.
func Overlay(lower, higher Selection) Selection {
	if higher.Context != "" && higher.Context != lower.Context {
		lower.Context = higher.Context
	}
	if higher.Server != "" && ServerKey(higher.Server) != ServerKey(lower.Server) {
		lower.Server = higher.Server
		lower.Tenant = ""
		lower.Inventory = ""
	}
	if higher.Tenant != "" && higher.Tenant != lower.Tenant {
		lower.Tenant = higher.Tenant
		lower.Inventory = ""
	}
	if higher.Inventory != "" {
		lower.Inventory = higher.Inventory
	}
	return lower
}

// Resolve accepts an authenticated principal key. An empty key never reuses scope.
func Resolve(config Config, request Selection, principal string) (Resolved, error) {
	selected, err := selectEntry(config, request, principal)
	if err != nil {
		return Resolved{}, err
	}
	result := Resolved{Context: selected.Name, Server: selected.Server}
	if request.Server != "" {
		result.Server = request.Server
	}
	if principal != "" && selected.Principal == principal && ServerKey(selected.Server) == ServerKey(result.Server) {
		result.Scope = ports.Scope{Tenant: selected.Tenant, Inventory: selected.Inventory}
	}
	if request.Tenant != "" {
		if request.Tenant != result.Scope.Tenant {
			result.Scope.Inventory = ""
		}
		result.Scope.Tenant = request.Tenant
	}
	if request.Inventory != "" {
		result.Scope.Inventory = request.Inventory
	}
	return result, nil
}
func selectEntry(config Config, request Selection, principal string) (Entry, error) {
	if config.Version != 0 && config.Version != Version {
		return Entry{}, ports.Failure("configuration", "This context file version is not supported. Update Stuff Stash CLI.")
	}
	name := request.Context
	if name == "" {
		name = config.Current
	}
	var selected Entry
	for _, entry := range config.Contexts {
		if entry.Name == name {
			selected = entry
			break
		}
	}
	if request.Context != "" && selected.Name == "" {
		return Entry{}, ports.Failure("configuration", "The context does not exist. Use stuffstash context list to see the available contexts.")
	}
	if request.Server == "" || ServerKey(request.Server) == ServerKey(selected.Server) {
		return selected, nil
	}
	selected = Entry{}
	for _, entry := range config.Contexts {
		if ServerKey(entry.Server) != ServerKey(request.Server) || entry.Principal != principal {
			continue
		}
		if selected.Name != "" {
			return Entry{}, ports.Failure("configuration", "More than one context uses this server. Use --context NAME.")
		}
		selected = entry
	}
	return selected, nil
}

// ServerKey follows the CLI endpoint convention: trailing slashes are optional.
func ServerKey(server string) string { return strings.TrimRight(server, "/") }
