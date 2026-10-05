package app

import "testing"

func TestScopeFlagsDoNotCarryEnvironmentAcrossBoundaries(t *testing.T) {
	env := map[string]string{"STUFF_STASH_CLI_SERVER": "https://home.example", "STUFF_STASH_CLI_TENANT": "home", "STUFF_STASH_CLI_INVENTORY": "old", "STUFF_STASH_CLI_CONTEXT": "home"}
	getenv := func(key string) string { return env[key] }
	for _, test := range []struct {
		args                               []string
		server, tenant, inventory, context string
	}{
		{[]string{"assets", "list", "--server", "https://work.example"}, "https://work.example", "", "", "home"},
		{[]string{"assets", "list", "--tenant", "work"}, "https://home.example", "work", "", "home"},
		{[]string{"assets", "list", "--context", "work"}, "https://home.example", "home", "old", "work"},
	} {
		got, err := Parse(test.args, getenv)
		if err != nil || got.Server != test.server || got.Scope.Tenant != test.tenant || got.Scope.Inventory != test.inventory || got.Selection.Context != test.context {
			t.Fatalf("%v: %+v %v", test.args, got, err)
		}
	}
}

func TestEmptyExplicitScopeCannotFallBackToEnvironment(t *testing.T) {
	getenv := func(key string) string {
		if key == "STUFF_STASH_CLI_SERVER" || key == "STUFF_STASH_CLI_TENANT" || key == "STUFF_STASH_CLI_INVENTORY" || key == "STUFF_STASH_CLI_CONTEXT" {
			return "saved"
		}
		return ""
	}
	for _, name := range []string{"server", "tenant", "inventory", "context"} {
		for _, args := range [][]string{{"assets", "list", "--" + name, ""}, {"assets", "list", "--" + name + "="}} {
			if _, err := Parse(args, getenv); err == nil {
				t.Fatalf("empty explicit %s accepted", name)
			}
		}
	}
}
