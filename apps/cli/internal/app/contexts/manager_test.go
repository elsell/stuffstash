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
