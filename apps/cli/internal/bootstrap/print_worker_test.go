//go:build linux

package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type connectorPeer struct {
	mu         sync.Mutex
	session    string
	revoked    bool
	heartbeats int
}

func (p *connectorPeer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if r.Header.Get("Authorization") != "Bearer restricted-secret" || p.revoked {
		w.WriteHeader(401)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	switch r.URL.Path {
	case "/print-consumer/heartbeat":
		var body struct {
			Session string `json:"sessionId"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil || len(body.Session) < 16 {
			w.WriteHeader(400)
			return
		}
		p.session = body.Session
		p.heartbeats++
		w.Write([]byte(`{"data":{"id":"connector","state":"active"},"meta":{}}`))
	case "/print-consumer/printers":
		p.revoked = true
		w.Write([]byte(`{"data":[],"meta":{}}`))
	default:
		w.WriteHeader(404)
	}
}
func TestForegroundWorkerUsesSeparateCredentialAndStopsOnRevocation(t *testing.T) {
	peer := &connectorPeer{}
	server := httptest.NewServer(peer)
	defer server.Close()
	store := credentials.ConnectorFile{Path: filepath.Join(t.TempDir(), "private", "connector.json")}
	if err := store.Save(context.Background(), ports.ConnectorRegistration{Server: server.URL, ConnectorID: "connector", TenantID: "tenant", InventoryID: "inventory", Credential: "restricted-secret", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{"STUFF_STASH_CLI_SERVER": server.URL, "STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP": "true", "STUFF_STASH_CLI_CONNECTOR_CREDENTIAL_FILE": store.Path, "STUFF_STASH_CLI_PRINT_HEARTBEAT_INTERVAL": "100ms", "STUFF_STASH_CLI_PRINT_STATE_DIRECTORY": filepath.Join(t.TempDir(), "state")}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var out, errors bytes.Buffer
	code := Run(ctx, []string{"connectors", "print", "run", "--connector", "connector"}, func(key string) string { return env[key] }, &out, &errors)
	if code != 1 || !strings.Contains(errors.String(), "Pair this connector again") {
		t.Fatalf("revoked connector did not terminate clearly: %d %s", code, errors.String())
	}
	peer.mu.Lock()
	defer peer.mu.Unlock()
	if peer.heartbeats != 1 || peer.session == "" {
		t.Fatal("connector did not authenticate its process session")
	}
	if strings.Contains(out.String()+errors.String(), "restricted-secret") {
		t.Fatal("connector credential leaked to output")
	}
}
