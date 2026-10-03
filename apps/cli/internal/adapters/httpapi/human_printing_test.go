package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// queuedLabelServer is a scoped, stateful HTTP queue with response-loss injection.
type queuedLabelServer struct {
	mu           sync.Mutex
	requests     map[string][]byte
	loseResponse bool
}

func (f *queuedLabelServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Authorization") != "Bearer editor" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "POST" || r.URL.Path != "/tenants/home/inventories/garage/assets/tool/print-jobs" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	var selection struct {
		PrinterID string `json:"printerId"`
		Media     string `json:"expectedMediaFingerprint"`
	}
	if json.Unmarshal(body, &selection) != nil || selection.PrinterID != "brother" || selection.Media != "29x90-v1" {
		w.WriteHeader(400)
		return
	}
	old, found := f.requests[key]
	if found && !bytes.Equal(old, body) {
		w.WriteHeader(http.StatusConflict)
		return
	}
	f.requests[key] = bytes.Clone(body)
	if f.loseResponse {
		f.loseResponse = false
		connection, _, err := w.(http.Hijacker).Hijack()
		if err == nil {
			_ = connection.Close()
		}
		return
	}
	if !found {
		w.WriteHeader(http.StatusCreated)
	}
	_, _ = io.WriteString(w, `{"data":{"id":"job-1","printerId":"brother","kind":"asset_label","copies":1,"status":"queued","revision":1,"requestedBy":"editor","mediaFingerprint":"29x90-v1","attempts":[],"createdAt":"2026-10-03T12:00:00Z","updatedAt":"2026-10-03T12:00:00Z"},"meta":{}}`)
}
func TestHumanPrintSDKRecoversLostResponseWithoutDuplicateAndPreservesScope(t *testing.T) {
	queue := &queuedLabelServer{requests: map[string][]byte{}, loseResponse: true}
	server := httptest.NewServer(queue)
	defer server.Close()
	client, err := New(server.URL, "editor", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	scope := ports.Scope{Tenant: "home", Inventory: "garage"}
	selection := ports.LabelPrintSelection{PrinterID: "brother", ExpectedMediaFingerprint: "29x90-v1", TemplateID: "qr-title", TemplateVersion: 1, Copies: 1}
	if _, err = client.QueueLabel(context.Background(), scope, "tool", selection, "stable-key"); err == nil {
		t.Fatal("response loss not reported")
	}
	result, err := client.QueueLabel(context.Background(), scope, "tool", selection, "stable-key")
	if err != nil || result.Data.ID != "job-1" || result.Data.Status != "queued" {
		t.Fatalf("retry %+v %v", result, err)
	}
	selection.Copies = 2
	if _, err = client.QueueLabel(context.Background(), scope, "tool", selection, "stable-key"); err == nil {
		t.Fatal("changed request reused key")
	}
	denied, _ := New(server.URL, "viewer", server.Client())
	if _, err = denied.QueueLabel(context.Background(), scope, "tool", selection, "other"); err == nil {
		t.Fatal("viewer enqueued")
	}
	if _, err = client.QueueLabel(context.Background(), ports.Scope{Tenant: "elsewhere", Inventory: "garage"}, "tool", selection, "other"); err == nil {
		t.Fatal("wrong tenant enqueued")
	}
	queue.mu.Lock()
	defer queue.mu.Unlock()
	if len(queue.requests) != 1 || !strings.Contains(string(queue.requests["stable-key"]), `"copies":1`) {
		t.Fatal("duplicate or rewritten output")
	}
}
