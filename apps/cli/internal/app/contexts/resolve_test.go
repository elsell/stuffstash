package contexts

import (
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"testing"
)

func TestResolveNeverMixesServerTenantOrPrincipalScopes(t *testing.T) {
	saved := Config{Version: 1, Current: "home", Contexts: []Entry{{Name: "home", Server: "https://home.example", Principal: "alice", Tenant: "house", Inventory: "garage"}, {Name: "work", Server: "https://work.example", Principal: "alice", Tenant: "office", Inventory: "supplies"}}}
	tests := []struct {
		name      string
		input     Selection
		principal string
		want      ports.Scope
	}{
		{"remembered", Selection{}, "alice", ports.Scope{Tenant: "house", Inventory: "garage"}},
		{"server override", Selection{Server: "https://work.example"}, "alice", ports.Scope{Tenant: "office", Inventory: "supplies"}},
		{"new server", Selection{Server: "https://new.example"}, "alice", ports.Scope{}},
		{"tenant override", Selection{Tenant: "other"}, "alice", ports.Scope{Tenant: "other"}},
		{"explicit inventory", Selection{Inventory: "attic"}, "alice", ports.Scope{Tenant: "house", Inventory: "attic"}},
		{"different principal", Selection{}, "bob", ports.Scope{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Resolve(saved, tt.input, tt.principal)
			if err != nil || got.Scope != tt.want {
				t.Fatalf("got %+v, %v; want %+v", got, err, tt.want)
			}
		})
	}
}

func TestRememberedContextRequiresUnambiguousServerChoice(t *testing.T) {
	saved := Config{Version: 1, Current: "a", Contexts: []Entry{{Name: "a", Server: "https://one.example", Principal: "alice", Tenant: "t", Inventory: "i"}, {Name: "b", Server: "https://two.example", Principal: "alice", Tenant: "t2", Inventory: "i2"}, {Name: "c", Server: "https://two.example", Principal: "alice", Tenant: "t3", Inventory: "i3"}}}
	if _, err := Resolve(saved, Selection{Server: "https://two.example"}, "alice"); err == nil {
		t.Fatal("ambiguous server silently chose context")
	}
	got, err := Resolve(saved, Selection{Context: "c"}, "alice")
	if err != nil || got.Scope.Inventory != "i3" {
		t.Fatalf("explicit context: %+v %v", got, err)
	}
	if _, err := Resolve(saved, Selection{Context: "missing"}, "alice"); err == nil {
		t.Fatal("unknown explicit context silently ignored")
	}
}

func TestFlagScopeOverridesEnvironmentWithoutInheritingItsInventory(t *testing.T) {
	env := Selection{Server: "https://home.example", Tenant: "one", Inventory: "old"}
	got := Overlay(env, Selection{Tenant: "two"})
	if got.Tenant != "two" || got.Inventory != "" {
		t.Fatalf("cross-tenant environment inventory retained: %+v", got)
	}
	got = Overlay(env, Selection{Server: "https://other.example"})
	if got.Tenant != "" || got.Inventory != "" {
		t.Fatalf("cross-server environment scope retained: %+v", got)
	}
}

func TestContextFlagPreservesExplicitEnvironmentScope(t *testing.T) {
	saved := Config{Version: Version, Current: "home", Contexts: []Entry{{Name: "home", Server: "https://home.example", Principal: "alice", Tenant: "saved", Inventory: "saved"}}}
	selection := Overlay(Selection{Server: "https://prod.example", Tenant: "explicit", Inventory: "chosen"}, Selection{Context: "home"})
	got, err := Resolve(saved, selection, "alice")
	if err != nil || got.Server != "https://prod.example" || got.Scope.Tenant != "explicit" || got.Scope.Inventory != "chosen" {
		t.Fatalf("explicit environment scope lost: %+v %v", got, err)
	}
}

func TestPrincipalUsesVerifiedIdentityNotRotatingToken(t *testing.T) {
	session := ports.Session{Issuer: "https://identity.example", Subject: "alice", IDToken: "old"}
	first := Principal(session)
	session.IDToken = "new"
	if first == "" || Principal(session) != first {
		t.Fatal("token renewal changed principal")
	}
	session.Subject = "bob"
	if Principal(session) == first {
		t.Fatal("accounts share a principal")
	}
	session.Subject = ""
	if Principal(session) != "" {
		t.Fatal("legacy session reused identity without verification")
	}
	session.Subject = "alice"
	session.Issuer = ""
	if Principal(session) != "" {
		t.Fatal("identity without issuer accepted")
	}
}
