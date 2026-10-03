package config

import "testing"

func TestCLIClientExplicitlyExtendsOIDCAudiences(t *testing.T) {
	t.Setenv("STUFF_STASH_OIDC_CLIENT_ID", "web")
	t.Setenv("STUFF_STASH_OIDC_CLIENT_IDS", "")
	t.Setenv("STUFF_STASH_OIDC_MOBILE_CLIENT_ID", "")
	t.Setenv("STUFF_STASH_OIDC_CLI_CLIENT_ID", "cli")
	cfg := Load()
	if len(cfg.OIDCClientIDs) != 2 || cfg.OIDCClientIDs[1] != "cli" {
		t.Fatalf("CLI audience missing: %v", cfg.OIDCClientIDs)
	}
}
