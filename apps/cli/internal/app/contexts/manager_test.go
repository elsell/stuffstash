package contexts

import (
	"context"
	"testing"
)

type memoryStore struct{ config Config }

func (s *memoryStore) Load(context.Context) (Config, error) {
	c := s.config
	c.Contexts = append([]Entry(nil), c.Contexts...)
	return c, nil
}
func (s *memoryStore) Update(ctx context.Context, change func(*Config) error) error {
	c, _ := s.Load(ctx)
	if err := change(&c); err != nil {
		return err
	}
	if err := Validate(c); err != nil {
		return err
	}
	s.config = c
	return nil
}
func TestContextLifecycleDoesNotSilentlyChangeScope(t *testing.T) {
	ctx := context.Background()
	store := &memoryStore{config: Config{Version: Version}}
	manager := Manager{Store: store}
	home := Entry{Name: "home", Server: "https://home.example", Principal: "alice", Tenant: "house", Inventory: "garage"}
	work := Entry{Name: "work", Server: "https://work.example", Principal: "alice", Tenant: "office", Inventory: "supplies"}
	for _, entry := range []Entry{home, work} {
		if err := manager.Remember(ctx, entry); err != nil {
			t.Fatal(err)
		}
	}
	if err := manager.Use(ctx, "home"); err != nil {
		t.Fatal(err)
	}
	current, err := manager.Current(ctx)
	if err != nil || current != home {
		t.Fatalf("current: %+v %v", current, err)
	}
	if err := manager.Use(ctx, "missing"); err == nil {
		t.Fatal("unknown context accepted")
	}
	current, _ = manager.Current(ctx)
	if current != home {
		t.Fatal("failed selection changed current")
	}
	for _, invalid := range []Entry{
		{Name: "home", Server: home.Server, Principal: "bob", Tenant: "other"},
		{Name: "home", Server: work.Server, Principal: "alice", Tenant: "other"},
		{Name: "anonymous", Server: home.Server, Tenant: "other"},
	} {
		if err := manager.Remember(ctx, invalid); err == nil {
			t.Fatal("unsafe identity replacement accepted")
		}
	}
	if err := manager.ClearServer(ctx, home.Server); err != nil {
		t.Fatal(err)
	}
	current, err = manager.Current(ctx)
	if err != nil || current.Principal != "" || current.Tenant != "" || current.Inventory != "" || current.Server != home.Server {
		t.Fatalf("logout retained scope: %+v %v", current, err)
	}
	entries, err := manager.List(ctx)
	if err != nil || len(entries) != 2 || entries[1] != work {
		t.Fatalf("logout changed other server: %+v %v", entries, err)
	}
	if err := manager.Delete(ctx, "home"); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Current(ctx); err == nil {
		t.Fatal("delete silently chose another context")
	}
	if store.config.Current != "" {
		t.Fatal("current context not cleared")
	}
}

func TestTrailingSlashDoesNotChangeContextIdentity(t *testing.T) {
	ctx := context.Background()
	store := &memoryStore{config: Config{Version: Version, Current: "home", Contexts: []Entry{{Name: "home", Server: "https://home.example/", Principal: "alice", Tenant: "home", Inventory: "garage"}}}}
	got, err := Resolve(store.config, Selection{Server: "https://home.example"}, "alice")
	if err != nil || got.Scope.Inventory != "garage" {
		t.Fatalf("scope lost: %+v %v", got, err)
	}
	if err := (Manager{Store: store}).ClearServer(ctx, "https://home.example"); err != nil {
		t.Fatal(err)
	}
	if store.config.Contexts[0].Principal != "" || store.config.Contexts[0].Tenant != "" {
		t.Fatal("logout retained trailing-slash account scope")
	}
}

func TestDeletedResourceScopeClearsOnlyMatchingAccountAndTarget(t *testing.T) {
	store := &memoryStore{config: Config{Version: Version, Contexts: []Entry{
		{Name: "selected", Server: "https://stash.example", Principal: "alice", Tenant: "home", Inventory: "garage"},
		{Name: "other-inventory", Server: "https://stash.example", Principal: "alice", Tenant: "home", Inventory: "loft"},
		{Name: "other-account", Server: "https://stash.example", Principal: "bob", Tenant: "home", Inventory: "garage"},
		{Name: "other-server", Server: "https://other.example", Principal: "alice", Tenant: "home", Inventory: "garage"},
	}}}
	manager := Manager{Store: store}
	if err := manager.ClearResource(context.Background(), "https://stash.example/", "alice", "home", "garage"); err != nil {
		t.Fatal(err)
	}
	entries := store.config.Contexts
	if entries[0].Tenant != "home" || entries[0].Inventory != "" || entries[1].Inventory != "loft" || entries[2].Inventory != "garage" || entries[3].Inventory != "garage" {
		t.Fatalf("wrong inventory cleanup: %+v", entries)
	}
	if err := manager.ClearResource(context.Background(), "https://stash.example", "alice", "home", ""); err != nil {
		t.Fatal(err)
	}
	entries = store.config.Contexts
	if entries[0].Tenant != "" || entries[1].Tenant != "" || entries[1].Inventory != "" || entries[2].Tenant != "home" || entries[3].Tenant != "home" {
		t.Fatalf("wrong household cleanup: %+v", entries)
	}
}
