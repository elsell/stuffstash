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

func TestPrinterInspectionPreservesConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer owner" || !strings.HasPrefix(r.URL.Path, "/tenants/home/inventories/garage/printers") {
			w.WriteHeader(403)
			return
		}
		if r.Method != "GET" {
			t.Error("wrong method")
		}
		data := `{"id":"printer","adapterId":"brother","name":"Desk printer","revision":5,"readiness":"unavailable","readinessReason":"Load paper","reportedAt":"2026-10-05T10:00:00Z","retired":false,"mediaFingerprint":"fingerprint","media":{"name":"Label","presetId":"preset","version":2,"widthMicrometers":62000,"heightMicrometers":29000,"marginsMicrometers":{"top":1,"bottom":2,"left":3,"right":4},"rasterWidth":732,"rasterHeight":342,"resolutionDpi":300,"displayRotation":90,"orientation":"landscape","colorMode":"monochrome","cutPolicy":"cut"}}`
		if strings.HasSuffix(r.URL.Path, "/printers") {
			if r.URL.Query().Get("limit") != "1" || r.URL.Query().Get("cursor") != "before" {
				t.Error("lost paging")
			}
			data = "[" + data + "]"
		}
		io.WriteString(w, `{"$schema":"printer-schema","data":`+data+`,"meta":{"requestId":"trace","pagination":{"hasMore":true,"nextCursor":"after"}}}`)
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

	for _, action := range [][]string{{"show", "printer"}, {"list", "--limit", "1", "--cursor", "before"}} {
		command := append([]string{"printers"}, action...)
		command = append(command, "--tenant", "home", "--inventory", "garage", "--json", "--no-input")
		var out, diagnostic bytes.Buffer
		if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 {
			t.Fatalf("read: %d %s", code, &diagnostic)
		}
		for _, field := range []string{`"readinessReason":"Load paper"`, `"reportedAt":"2026-10-05T10:00:00Z"`, `"media":{`, `"widthMicrometers":62000`, `"heightMicrometers":29000`, `"top":1`, `"bottom":2`, `"left":3`, `"right":4`, `"rasterWidth":732`, `"rasterHeight":342`, `"resolutionDpi":300`, `"displayRotation":90`, `"orientation":"landscape"`, `"colorMode":"monochrome"`, `"cutPolicy":"cut"`, `"requestId":"trace"`, `"$schema":"printer-schema"`} {
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
}
