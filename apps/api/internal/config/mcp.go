package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type MCPConfiguration struct {
	Enabled   bool
	PublicURL string
	AuthMode  string
	Issuer    string
}

func LoadMCP(apiAuthMode, issuer string) (MCPConfiguration, error) {
	c := MCPConfiguration{}
	raw := os.Getenv("STUFF_STASH_MCP_ENABLED")
	if raw != "" {
		enabled, err := strconv.ParseBool(raw)
		if err != nil {
			return c, fmt.Errorf("STUFF_STASH_MCP_ENABLED must be a boolean")
		}
		c.Enabled = enabled
	}
	if !c.Enabled {
		return c, nil
	}
	c.AuthMode = strings.TrimSpace(os.Getenv("STUFF_STASH_MCP_AUTH_MODE"))
	if (c.AuthMode != "oidc" && c.AuthMode != "local-dev") || c.AuthMode != apiAuthMode {
		return MCPConfiguration{}, fmt.Errorf("STUFF_STASH_MCP_AUTH_MODE must explicitly match the API authentication mode")
	}
	c.PublicURL = strings.TrimSpace(os.Getenv("STUFF_STASH_MCP_PUBLIC_URL"))
	parsed, err := url.Parse(c.PublicURL)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" || parsed.EscapedPath() != "/mcp" {
		return MCPConfiguration{}, fmt.Errorf("STUFF_STASH_MCP_PUBLIC_URL must be an absolute endpoint URL ending in /mcp without credentials, query or fragment")
	}
	loopback := parsed.Hostname() == "localhost"
	if ip := net.ParseIP(parsed.Hostname()); ip != nil {
		loopback = ip.IsLoopback()
	}
	if parsed.Scheme != "https" && !(c.AuthMode == "local-dev" && parsed.Scheme == "http" && loopback) {
		return MCPConfiguration{}, fmt.Errorf("STUFF_STASH_MCP_PUBLIC_URL requires HTTPS except local-dev loopback HTTP")
	}
	if c.AuthMode == "oidc" {
		if strings.TrimSpace(issuer) == "" {
			return MCPConfiguration{}, fmt.Errorf("MCP OIDC authentication requires the configured API issuer")
		}
		c.Issuer = issuer
	}
	return c, nil
}
