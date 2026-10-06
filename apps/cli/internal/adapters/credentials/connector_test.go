package credentials

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func TestConnectorFileSeparatesIdentityAndProtectsSecret(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "private", "connector.json")
	store := ConnectorFile{Path: path}
	registration := ports.ConnectorRegistration{Server: "https://stash.example", TenantID: "tenant", InventoryID: "inventory", ConnectorID: "connector", Credential: "secret", ExpiresAt: time.Now().UTC(), ActivationDeadline: time.Date(2030, 1, 2, 3, 4, 5, 678, time.UTC)}
	if err := store.Save(ctx, registration); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load(ctx, registration.Server, registration.ConnectorID)
	if err != nil || got.Credential != registration.Credential || !got.ActivationDeadline.Equal(registration.ActivationDeadline) {
		t.Fatalf("saved credential unavailable: %v", err)
	}
	for _, identity := range [][2]string{{"https://other.example", "connector"}, {registration.Server, "other"}} {
		if _, err := store.Load(ctx, identity[0], identity[1]); !errors.Is(err, ports.ErrConnectorNotRegistered) {
			t.Fatal("identity mismatch disclosed credential")
		}
	}
	other := registration
	other.ConnectorID = "other"
	if err := store.Save(ctx, other); err == nil {
		t.Fatal("overwrote another connector")
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(ctx, registration.Server, registration.ConnectorID); err == nil {
		t.Fatal("read world-readable secret")
	}
}

func TestConnectorFileRejectsHumanSessionAndSymlink(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "private", "credential.json")
	if err := (File{Path: path}).Save(ctx, ports.Session{Server: "https://stash.example", IDToken: "human-token"}); err != nil {
		t.Fatal(err)
	}
	registration := ports.ConnectorRegistration{Server: "https://stash.example", TenantID: "t", InventoryID: "i", ConnectorID: "c", Credential: "machine", ExpiresAt: time.Now().Add(time.Hour)}
	if err := (ConnectorFile{Path: path}).Save(ctx, registration); err == nil {
		t.Fatal("overwrote human credential")
	}
	link := path + ".link"
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	if err := (ConnectorFile{Path: link}).Save(ctx, registration); err == nil {
		t.Fatal("followed symlink")
	}
}

func TestConcurrentConnectorRegistrationsKeepOneIdentity(t *testing.T) {
	ctx := context.Background()
	store := ConnectorFile{Path: filepath.Join(t.TempDir(), "private", "connector.json")}
	start := make(chan struct{})
	results := make(chan error, 16)
	for i := 0; i < 16; i++ {
		go func(i int) {
			<-start
			results <- store.Save(ctx, ports.ConnectorRegistration{Server: "https://stash.example", TenantID: "t", InventoryID: "i", ConnectorID: fmt.Sprintf("connector-%d", i), Credential: "secret", ExpiresAt: time.Now().Add(time.Hour)})
		}(i)
	}
	close(start)
	success := 0
	for i := 0; i < 16; i++ {
		if <-results == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("%d concurrent registrations succeeded; want exactly one", success)
	}
}
