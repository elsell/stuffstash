package bootstrap

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPrintJobInspectionPreservesEvidence(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer owner" || !strings.HasPrefix(r.URL.Path, "/tenants/home/inventories/garage/print-jobs") {
			w.WriteHeader(403)
			return
		}
		if r.Method != "GET" {
			t.Error("wrong method")
		}
		data := `{"id":"job","printerId":"printer","assetId":"asset","predecessor":"previous","kind":"asset_label","requestedBy":"owner","mediaFingerprint":"media","revision":5,"copies":2,"status":"uncertain","createdAt":"2026-10-05T10:00:00Z","updatedAt":"2026-10-05T11:00:00Z","resolution":{"reportedOutcome":"printed","resolvedBy":"owner","resolvedAt":"2026-10-05T11:00:00Z"},"attempts":[{"id":"attempt","connectorId":"connector","outcome":"uncertain","reason":"Check printer","completedCopies":1,"claimedAt":"2026-10-05T10:01:00Z","leaseExpiresAt":"2026-10-05T10:02:00Z","startedAt":"2026-10-05T10:01:01Z","settledAt":"2026-10-05T10:01:30Z","idleConfirmedAt":"2026-10-05T10:03:00Z"}]}`
		if strings.HasSuffix(r.URL.Path, "/print-jobs") {
			if r.URL.Query().Get("printerId") != "printer" || r.URL.Query().Get("cursor") != "before" || r.URL.Query().Get("limit") != "1" {
				t.Error("lost list options")
			}
			data = "[" + data + "]"
		}
		io.WriteString(w, `{"$schema":"job-schema","data":`+data+`,"meta":{"requestId":"trace","pagination":{"hasMore":true,"nextCursor":"after"}}}`)
	}))
	defer server.Close()
	dir := t.TempDir()
	os.Chmod(dir, 0700)
	store := credentials.File{Path: filepath.Join(dir, "session.json")}
	if err := store.Save(context.Background(), ports.Session{Server: server.URL, Issuer: "https://id.example", Subject: "owner", IDToken: "owner", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	getenv := func(k string) string {
		switch k {
		case "STUFF_STASH_CLI_SERVER":
			return server.URL
		case "STUFF_STASH_CLI_CREDENTIAL_FILE":
			return store.Path
		case "STUFF_STASH_CLI_CONFIG_FILE":
			return filepath.Join(dir, "config", "contexts.json")
		case "STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP":
			return "true"
		}
		return ""
	}

	for _, action := range [][]string{{"show", "job"}, {"list", "--printer", "printer", "--cursor", "before", "--limit", "1"}} {
		command := append([]string{"print-jobs"}, action...)
		command = append(command, "--tenant", "home", "--inventory", "garage", "--json", "--no-input")
		var out, diagnostic bytes.Buffer
		if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 {
			t.Fatalf("read: %d %s", code, &diagnostic)
		}
		for _, field := range []string{`"requestedBy":"owner"`, `"mediaFingerprint":"media"`, `"updatedAt":"2026-10-05T11:00:00Z"`, `"reportedOutcome":"printed"`, `"claimedAt":"2026-10-05T10:01:00Z"`, `"leaseExpiresAt":"2026-10-05T10:02:00Z"`, `"startedAt":"2026-10-05T10:01:01Z"`, `"settledAt":"2026-10-05T10:01:30Z"`, `"idleConfirmedAt":"2026-10-05T10:03:00Z"`, `"requestId":"trace"`, `"$schema":"job-schema"`} {
			if !strings.Contains(out.String(), field) {
				t.Fatalf("missing %s: %s", field, &out)
			}
		}
		for _, override := range [][]string{{"--tenant", "other"}, {"--inventory", "other"}} {
			if code := Run(context.Background(), append(command, override...), getenv, &out, &diagnostic); code == 0 {
				t.Fatal("scope bypass")
			}
		}
	}
	var out, diagnostic bytes.Buffer
	if code := Run(context.Background(), []string{"print-jobs", "show", "job", "--tenant", "home", "--inventory", "garage", "--no-input"}, getenv, &out, &diagnostic); code != 0 || !strings.Contains(out.String(), "Check printer") || !strings.Contains(out.String(), "printed") {
		t.Fatalf("missing human evidence: %d %s %s", code, &out, &diagnostic)
	}
}
