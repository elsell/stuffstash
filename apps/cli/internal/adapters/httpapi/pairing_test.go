package httpapi

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/pairingkeys"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// pairingServer implements key-bound, once-only exchange and activation over
// real HTTP. It holds state rather than matching a scripted sequence of calls.
type pairingServer struct {
	public                     ed25519.PublicKey
	approved, consumed, active bool
	activationDeadline         time.Time
}

func (s *pairingServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	respond := func(value any) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": value, "meta": map[string]any{}})
	}
	switch r.URL.Path {
	case "/print-connector-pairings":
		if r.Header.Get("Authorization") != "" {
			http.Error(w, "human credential not allowed", 403)
			return
		}
		var input struct {
			PublicKey []byte `json:"publicKey"`
		}
		if json.NewDecoder(r.Body).Decode(&input) != nil || len(input.PublicKey) != ed25519.PublicKeySize {
			http.Error(w, "key", 400)
			return
		}
		s.public = input.PublicKey
		w.WriteHeader(201)
		respond(map[string]any{"id": "pair", "pollToken": "poll-secret", "userCode": "ABCD1234", "verificationUrl": "https://web.example/pair/pair", "expiresAt": time.Now().Add(time.Minute)})
	case "/print-connector-pairings/pair":
		if r.Header.Get("X-Pairing-Token") != "poll-secret" {
			http.Error(w, "denied", 401)
			return
		}
		state := "pending"
		if s.approved {
			state = "approved"
		}
		if s.consumed {
			state = "consumed"
		}
		respond(map[string]any{"id": "pair", "state": state, "expiresAt": time.Now().Add(time.Minute)})
	case "/print-connector-pairings/pair/credential":
		var input struct {
			Signature []byte `json:"signature"`
		}
		_ = json.NewDecoder(r.Body).Decode(&input)
		if !s.approved || s.consumed || r.Header.Get("X-Pairing-Token") != "poll-secret" || !ed25519.Verify(s.public, []byte("stuffstash-print-pairing-v1\npair\npoll-secret"), input.Signature) {
			http.Error(w, "denied", 403)
			return
		}
		s.consumed = true
		respond(map[string]any{"connectorId": "connector", "credential": "machine-secret", "tenantId": "tenant", "inventoryId": "inventory", "expiresAt": time.Now().Add(time.Hour), "activationDeadline": s.activationDeadline})
	case "/print-consumer/heartbeat":
		if !s.consumed || r.Header.Get("Authorization") != "Bearer machine-secret" {
			http.Error(w, "denied", 401)
			return
		}
		s.active = true
		respond(map[string]any{"id": "connector"})
	default:
		http.NotFound(w, r)
	}
}
func TestGeneratedPairingClientPreservesKeyAndMachineCredentialBoundary(t *testing.T) {
	ctx := context.Background()
	state := &pairingServer{activationDeadline: time.Date(2030, 1, 2, 3, 4, 5, 678, time.UTC)}
	server := httptest.NewTLSServer(state)
	defer server.Close()
	api, err := NewPairing(server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	receipts := &pairingReceiptRecorder{}
	api.Receipts = receipts
	key, _ := (pairingkeys.Keys{}).NewKey()
	challenge, err := api.Start(ctx, ports.PairingRequest{Name: "Garage", PublicKey: key.PublicKey()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := api.Exchange(ctx, challenge, key.Sign(challenge.ID, challenge.PollToken)); err == nil {
		t.Fatal("unapproved exchange succeeded")
	}
	state.approved = true
	if status, err := api.Poll(ctx, challenge); err != nil || status != ports.PairingApproved {
		t.Fatalf("poll: %v", err)
	}
	wrong, _ := (pairingkeys.Keys{}).NewKey()
	if _, err := api.Exchange(ctx, challenge, wrong.Sign(challenge.ID, challenge.PollToken)); err == nil {
		t.Fatal("foreign key exchanged")
	}
	registration, err := api.Exchange(ctx, challenge, key.Sign(challenge.ID, challenge.PollToken))
	if err != nil {
		t.Fatal(err)
	}
	if !registration.ActivationDeadline.Equal(state.activationDeadline) {
		t.Fatalf("activation deadline lost: %v", registration.ActivationDeadline)
	}
	if _, err := api.Exchange(ctx, challenge, key.Sign(challenge.ID, challenge.PollToken)); err == nil {
		t.Fatal("consumed proof exchanged again")
	}
	if len(receipts.receipts) != 3 {
		t.Fatalf("failed exchanges produced receipts: %d", len(receipts.receipts))
	}
	if err := api.Activate(ctx, registration, "session"); err != nil || !state.active {
		t.Fatalf("activation: %v", err)
	}
	if len(receipts.receipts) != 4 || receipts.receipts[3].Operation != "heartbeat" {
		t.Fatal("activation heartbeat receipt missing")
	}
}
func TestPairingClientDoesNotForwardSecretAcrossRedirect(t *testing.T) {
	reached := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached = true }))
	defer target.Close()
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer origin.Close()
	api, err := NewPairing(origin.URL, origin.Client())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := api.Poll(context.Background(), ports.PairingChallenge{ID: "pair", PollToken: "secret"}); err == nil || reached {
		t.Fatal("followed secret-bearing redirect")
	}
}
