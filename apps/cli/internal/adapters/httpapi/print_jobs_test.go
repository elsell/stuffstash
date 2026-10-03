package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// printPeer is a controlled API implementation with owned claims, revisions,
// settlement and protected artifact reads; it never asserts a call sequence.
type printPeer struct {
	mu       sync.Mutex
	attempt  generated.PrintConsumerAttempt
	token    string
	revoked  bool
	redirect string
	bytes    []byte
}

func (p *printPeer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.revoked || r.Header.Get("Authorization") != "Bearer connector-secret" {
		w.WriteHeader(401)
		return
	}
	if r.URL.Path == "/print-consumer/claims" && r.Method == "POST" {
		var body generated.PrintClaimRequest
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			w.WriteHeader(400)
			return
		}
		if p.attempt.AttemptId != "" && p.attempt.SettledAt == nil {
			w.WriteHeader(409)
			return
		}
		p.token = body.ClaimToken
		p.attempt = generated.PrintConsumerAttempt{AttemptId: body.AttemptId, SessionId: body.SessionId, PrinterId: body.PrinterId, JobId: "job", Revision: 1, Status: "claimed", ProtocolVersion: 1, LeaseValid: true, LeaseExpiresAt: time.Now().Add(time.Minute), Copies: 1, Media: &generated.PrintConsumerMedia{PresetId: "media", Version: 1}, Artifact: &generated.PrintArtifact{ContentType: "image/png", ByteLength: int64(len(p.bytes))}}
	} else if r.URL.Path == "/print-consumer/attempts/"+p.attempt.AttemptId && r.Method == "GET" {
		// Recovery returns only status/evidence, never artifact content.
		value := p.attempt
		value.Artifact = nil
		value.Media = nil
		json.NewEncoder(w).Encode(generated.SuccessEnvelopePrintConsumerAttempt{Data: value})
		return
	} else if r.URL.Path == "/print-consumer/claims/"+p.attempt.AttemptId+"/content" {
		if r.Header.Get("X-Print-Claim-Token") != p.token || r.Header.Get("X-Print-Session-ID") != p.attempt.SessionId || r.Header.Get("X-Print-Revision") != strconv.FormatInt(p.attempt.Revision, 10) {
			w.WriteHeader(403)
			return
		}
		if p.redirect != "" {
			http.Redirect(w, r, p.redirect, 307)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.(http.Flusher).Flush()
		w.Write(p.bytes)
		return
	} else if strings.HasPrefix(r.URL.Path, "/print-consumer/claims/"+p.attempt.AttemptId+"/") {
		var body generated.PrintOutcomeRequest
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			w.WriteHeader(400)
			return
		}
		if body.ClaimToken != p.token || body.SessionId != p.attempt.SessionId || body.Revision != p.attempt.Revision {
			w.WriteHeader(409)
			return
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/start"):
			if p.attempt.Status != "claimed" {
				w.WriteHeader(409)
				return
			}
			now := time.Now()
			p.attempt.StartedAt = &now
			p.attempt.Status = "printing"
		case strings.HasSuffix(r.URL.Path, "/outcome"):
			p.attempt.Outcome = body.Outcome
			now := time.Now()
			p.attempt.SettledAt = &now
			if body.Outcome.Kind == "no_output" {
				p.attempt.Status = "queued"
			} else {
				p.attempt.Status = string(body.Outcome.Kind)
			}
		default:
			w.WriteHeader(404)
			return
		}
		p.attempt.Revision++
	} else {
		w.WriteHeader(404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(generated.SuccessEnvelopePrintConsumerAttempt{Data: p.attempt})
}
func TestPrintSDKPreservesOwnedProofAndSettledAttemptDespiteRequeuedJob(t *testing.T) {
	peer := &printPeer{bytes: []byte("rendered-png")}
	server := httptest.NewServer(peer)
	defer server.Close()
	client, err := httpapi.New(server.URL, "connector-secret", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	requested := printing.AttemptControl{AttemptID: "attempt", SessionID: "process", ClaimToken: "secret-claim-token"}
	claim, err := client.Claim(ctx, "printer", requested)
	if err != nil {
		t.Fatal(err)
	}
	wrong := claim.Control
	wrong.ClaimToken = "another-attempt-token"
	if _, _, err = client.Artifact(ctx, wrong, 100); err == nil {
		t.Fatal("wrong attempt proof fetched content")
	}
	body, _, err := client.Artifact(ctx, claim.Control, 100)
	if err != nil || string(body) != "rendered-png" {
		t.Fatalf("owned artifact: %v", err)
	}
	started, err := client.Start(ctx, claim.Control)
	if err != nil {
		t.Fatal(err)
	}
	claim.Control.Revision = started.Revision
	if err = client.Outcome(ctx, claim.Control, printing.Evidence{Outcome: printing.NoOutput, Reason: printing.Disconnected}); err != nil {
		t.Fatal(err)
	}
	recovered, err := client.Attempt(ctx, "attempt")
	if err != nil || recovered.Phase != printing.RemoteFailed {
		t.Fatalf("requeued parent hid settled attempt: %+v %v", recovered, err)
	}
	peer.mu.Lock()
	peer.revoked = true
	peer.mu.Unlock()
	_, err = client.Attempt(ctx, "attempt")
	var failure *ports.Error
	if !errors.As(err, &failure) || failure.Category != "authentication" {
		t.Fatalf("revocation must terminate worker: %v", err)
	}
}
func TestPrintArtifactLimitAndRedirectNeverLeakClaimCredentials(t *testing.T) {
	peer := &printPeer{bytes: []byte("too-many-bytes")}
	server := httptest.NewServer(peer)
	defer server.Close()
	client, _ := httpapi.New(server.URL, "connector-secret", server.Client())
	claim, err := client.Claim(context.Background(), "printer", printing.AttemptControl{AttemptID: "attempt", SessionID: "process", ClaimToken: "claim-token"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = client.Artifact(context.Background(), claim.Control, 3); err == nil {
		t.Fatal("unbounded artifact accepted")
	}
	var leaked bool
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { leaked = true }))
	defer target.Close()
	peer.mu.Lock()
	peer.redirect = target.URL
	peer.mu.Unlock()
	if _, _, err = client.Artifact(context.Background(), claim.Control, 100); err == nil {
		t.Fatal("redirect artifact accepted")
	}
	if leaked {
		t.Fatal("connector or claim credentials followed a redirect")
	}
}

func TestResolvedUncertaintySettlesJournalWithoutFabricatingCompletion(t *testing.T) {
	now := time.Now().UTC()
	peer := &printPeer{attempt: generated.PrintConsumerAttempt{AttemptId: "attempt", SessionId: "session", JobId: "job", Revision: 8, Status: "failed", SettledAt: &now, ResolvedAt: &now, Outcome: generated.PrintOutcome{Kind: "uncertain"}}}
	server := httptest.NewServer(peer)
	defer server.Close()
	client, err := httpapi.New(server.URL, "connector-secret", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	status, err := client.Attempt(context.Background(), "attempt")
	if err != nil || status.Phase != printing.RemoteFailed || status.Outcome != printing.Uncertain {
		t.Fatal("resolved uncertainty remained blocked or fabricated evidence", status, err)
	}
}
