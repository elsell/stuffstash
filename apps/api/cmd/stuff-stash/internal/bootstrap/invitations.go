package bootstrap

import (
	"fmt"
	"net/netip"
	"net/url"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/config"
)

func validateInvitationPublicBaseURL(cfg config.Config) error {
	value := strings.TrimSpace(cfg.InvitationPublicBaseURL)
	if value == "" {
		if cfg.AuthMode == "" || cfg.AuthMode == "local-dev" {
			return nil
		}
		return fmt.Errorf("invitation public base URL is required outside local development")
	}
	if strings.Contains(value, "\\") {
		return fmt.Errorf("invitation public base URL contains an invalid path separator")
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("invitation public base URL must be an absolute URL without credentials, query, or fragment")
	}
	if parsed.Scheme == "https" {
		return nil
	}
	if parsed.Scheme == "http" && cfg.InvitationAllowInsecureLocalHTTP && isLocalDevelopmentInvitationHost(parsed.Hostname()) {
		return nil
	}
	return fmt.Errorf("invitation public base URL must use HTTPS outside local development")
}

func isLocalDevelopmentInvitationHost(host string) bool {
	if host == "localhost" {
		return true
	}
	address, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	if address.IsLoopback() {
		return true
	}
	if !address.Is4() {
		return false
	}
	octets := address.As4()
	return octets[0] == 10 ||
		(octets[0] == 172 && octets[1] >= 16 && octets[1] <= 31) ||
		(octets[0] == 192 && octets[1] == 168)
}
