package contexts

import (
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/url"
	"strings"
)

func Validate(config Config) error {
	invalid := func() error {
		return ports.Failure("configuration", "The context file is not valid. Check its version, context names and server addresses.")
	}
	if config.Version != Version {
		return invalid()
	}
	seen := map[string]bool{}
	for _, entry := range config.Contexts {
		server, err := url.Parse(entry.Server)
		if strings.TrimSpace(entry.Name) == "" || seen[entry.Name] || err != nil || server.Host == "" || (server.Scheme != "https" && server.Scheme != "http") || server.User != nil || server.RawQuery != "" || server.Fragment != "" || (entry.Inventory != "" && entry.Tenant == "") {
			return invalid()
		}
		seen[entry.Name] = true
	}
	if config.Current != "" && !seen[config.Current] {
		return invalid()
	}
	return nil
}
