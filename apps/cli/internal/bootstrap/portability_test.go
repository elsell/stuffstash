package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
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

const archiveFixture = `{"id":"job","createdAt":"2026-01-01T00:00:00Z","expiresAt":"2026-01-02T00:00:00Z","destinationInventoryId":"new","inventoryId":"inv","failure":null,"kind":"backup","otherFiles":false,"photos":true,"phase":"ready","state":"complete"}`

func portabilityEnvironment(t *testing.T, server string) (string, func(string) string) {
	t.Helper()
	dir := t.TempDir()
	os.Chmod(dir, 0700)
	store := credentials.File{Path: filepath.Join(dir, "session.json")}
	if err := store.Save(context.Background(), ports.Session{Server: server, Issuer: "https://id.example", Subject: "owner", IDToken: "owner", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	return dir, func(k string) string {
		switch k {
		case "STUFF_STASH_CLI_SERVER":
			return server
		case "STUFF_STASH_CLI_CREDENTIAL_FILE":
			return store.Path
		case "STUFF_STASH_CLI_CONFIG_FILE":
			return filepath.Join(dir, "contexts.json")
		case "STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP":
			return "true"
		}
		return ""
	}
}
func TestArchiveCommandsScopePayloadAndDenials(t *testing.T) {
	for _, tc := range []struct {
		command            []string
		method, path, body string
	}{
		{[]string{"list"}, "GET", "", ""}, {[]string{"show", "job"}, "GET", "/job", ""},
		{[]string{"create"}, "POST", "", `{"$schema":"schema","inventoryId":"inv","photos":false,"otherFiles":true}`},
		{[]string{"approve", "job"}, "POST", "/job/approve", `{"$schema":"schema","name":"Restored home"}`},
		{[]string{"retry", "job"}, "POST", "/job/retry", ""}, {[]string{"delete", "job"}, "DELETE", "/job", ""},
	} {
		t.Run(strings.Join(tc.command, "_"), func(t *testing.T) {
			calls, status := 0, 200
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != tc.method {
					t.Error("unexpected method")
				}
				if r.URL.Path != "/tenants/home/archive-jobs"+tc.path || r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					return
				}
				if r.URL.Query().Get("inventoryId") != "" {
					t.Error("implicit inventory filter")
				}
				body, _ := io.ReadAll(r.Body)
				if string(body) != tc.body {
					t.Errorf("changed input: %q", body)
				}
				if tc.command[0] == "create" && r.Header.Get("Idempotency-Key") == "" {
					t.Error("missing retry key")
				}
				if status != 200 {
					w.WriteHeader(status)
					io.WriteString(w, "private-denial")
					return
				}
				if tc.method == "DELETE" {
					w.WriteHeader(204)
					return
				}
				data := archiveFixture
				if tc.command[0] == "list" {
					data = "[" + data + "]"
				}
				io.WriteString(w, `{"$schema":"archive-schema","data":`+data+`,"meta":{"requestId":"trace","tenantId":"home"}}`)
			}))
			defer server.Close()
			dir, env := portabilityEnvironment(t, server.URL)
			args := append([]string{"archive-jobs"}, tc.command...)
			args = append(args, "--tenant", "home", "--json", "--no-input")
			if tc.body != "" {
				path := filepath.Join(dir, "input.json")
				os.WriteFile(path, []byte(tc.body), 0600)
				args = append(args, "--input", path)
			}
			var out, diag bytes.Buffer
			if tc.method != "GET" {
				if code := Run(context.Background(), args, env, &out, &diag); code != 2 || calls != 0 {
					t.Fatalf("unconfirmed %d %s", code, &diag)
				}
				args = append(args, "--yes")
			}
			out.Reset()
			diag.Reset()
			if code := Run(context.Background(), args, env, &out, &diag); code != 0 || calls != 1 {
				t.Fatalf("command %d %s", code, &diag)
			}
			if tc.method != "DELETE" {
				for _, field := range []string{`"$schema":"archive-schema"`, `"otherFiles":false`, `"destinationInventoryId":"new"`, `"requestId":"trace"`} {
					if !strings.Contains(out.String(), field) {
						t.Errorf("missing %s: %s", field, &out)
					}
				}
			}
			for _, denial := range []int{401, 403, 409, 500} {
				status = denial
				before := calls
				out.Reset()
				diag.Reset()
				if code := Run(context.Background(), args, env, &out, &diag); code == 0 || calls != before+1 || strings.Contains(out.String()+diag.String(), "private-denial") {
					t.Fatalf("denial/retry %d %s", code, &diag)
				}
			}
			status = 200
			before := calls
			if code := Run(context.Background(), append(args, "--tenant", "other"), env, &out, &diag); code == 0 || calls != before+1 {
				t.Fatal("wrong scope succeeded")
			}
		})
	}
}
func TestPortabilityRejectsMalformedInputBeforeAuth(t *testing.T) {
	for _, tc := range []struct {
		command []string
		body    string
	}{
		{[]string{"archive-jobs", "create"}, `{"inventoryId":"inv","photos":false}`},
		{[]string{"archive-jobs", "approve", "job"}, `{"name":false}`},
		{[]string{"import-jobs", "preview"}, `{"sourceType":"invalid","password":"secret-marker"}`},
		{[]string{"import-jobs", "start", "job"}, `{"sourceType":"legacy_homebox","allowInsecureTLS":"secret-marker"}`},
	} {
		path := filepath.Join(t.TempDir(), "input.json")
		os.WriteFile(path, []byte(tc.body), 0600)
		args := append(tc.command, "--input", path, "--tenant", "home", "--inventory", "inv", "--server", "https://stash.example", "--no-input", "--yes")
		var out, diag bytes.Buffer
		if code := Run(context.Background(), args, func(string) string { return "" }, &out, &diag); code != 2 || strings.Contains(out.String()+diag.String(), "secret-marker") {
			t.Fatalf("invalid input %d %s", code, &diag)
		}
	}
}
func TestImportSourceCommandsPreserveProtectedInput(t *testing.T) {
	for _, verb := range []string{"preview", "start"} {
		t.Run(verb, func(t *testing.T) {
			body := `{"$schema":"source-schema","sourceType":"legacy_homebox","baseUrl":"https://source.example","username":"protected-user","password":"protected-password","includeImages":false,"allowInsecureTLS":false,"allowPrivateNetwork":false,"fileName":null,"contentBase64":null}`
			calls, status := 0, 200
			path := "/tenants/home/inventories/inv/imports/jobs/preview"
			if verb == "start" {
				path = "/tenants/home/inventories/inv/imports/jobs/job/start"
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != "POST" || r.URL.Path != path || r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					return
				}
				got, _ := io.ReadAll(r.Body)
				if string(got) != body {
					t.Error("source JSON changed")
				}
				w.WriteHeader(status)
				if status != 200 {
					io.WriteString(w, "protected-password")
					return
				}
				io.WriteString(w, `{"$schema":"job-schema","data":{"id":"job","status":"previewed","createdAt":"today","updatedAt":"today","source":{"type":"legacy_homebox","name":"Source","imageImport":"none","allowPrivateNetwork":false,"allowInsecureTLS":false},"counts":{"assetsCreated":9007199254740993},"preview":{},"progress":{}},"meta":{"requestId":"trace"}}`)
			}))
			defer server.Close()
			dir, env := portabilityEnvironment(t, server.URL)
			input := filepath.Join(dir, "source.json")
			os.WriteFile(input, []byte(body), 0600)
			args := []string{"import-jobs", verb}
			if verb == "start" {
				args = append(args, "job")
			}
			args = append(args, "--tenant", "home", "--inventory", "inv", "--input", input, "--no-input", "--json")
			var out, diag bytes.Buffer
			if code := Run(context.Background(), args, env, &out, &diag); code != 2 || calls != 0 {
				t.Fatalf("unconfirmed %d %s", code, &diag)
			}
			args = append(args, "--yes")
			out.Reset()
			diag.Reset()
			if code := Run(context.Background(), args, env, &out, &diag); code != 0 || calls != 1 || !strings.Contains(out.String(), `"assetsCreated":9007199254740993`) {
				t.Fatalf("source %d %s %s", code, &out, &diag)
			}
			if strings.Contains(out.String()+diag.String(), "protected-") {
				t.Fatal("source credentials exposed")
			}
			for _, s := range []int{401, 403, 409, 500} {
				status = s
				before := calls
				out.Reset()
				diag.Reset()
				if code := Run(context.Background(), args, env, &out, &diag); code == 0 || calls != before+1 || strings.Contains(out.String()+diag.String(), "protected-") {
					t.Fatalf("unsafe denial %d %s", code, &diag)
				}
			}
		})
	}
}

func TestArchivePreviewAndExplicitFilter(t *testing.T) {
	filter := ""
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/tenants/home/archive-jobs/job/preview" || r.URL.Query().Get("inventoryId") != filter {
			t.Errorf("scope: %s", r.URL)
		}
		io.WriteString(w, `{"$schema":"preview-schema","data":{"assets":9007199254740993,"customAssetTypes":2,"customFields":3,"inventoryName":"Restored","keyRemappings":[{"destinationKey":"new","sourceKey":"old","family":"field"}],"omittedAttachments":4,"otherFiles":5,"photos":6,"tags":7},"meta":{"requestId":"trace"}}`)
	}))
	defer server.Close()
	dir, env := portabilityEnvironment(t, server.URL)
	principal := contexts.Principal(ports.Session{Issuer: "https://id.example", Subject: "owner"})
	config := contexts.Config{Version: 1, Current: "saved", Contexts: []contexts.Entry{{Name: "saved", Server: server.URL, Principal: principal, Tenant: "home", Inventory: "remembered"}}}
	data, _ := json.Marshal(config)
	os.WriteFile(filepath.Join(dir, "contexts.json"), data, 0600)
	args := []string{"archive-jobs", "preview", "job", "--json", "--no-input"}
	var out, diag bytes.Buffer
	if code := Run(context.Background(), args, env, &out, &diag); code != 0 || calls != 1 || !strings.Contains(out.String(), `"assets":9007199254740993`) || !strings.Contains(out.String(), `"sourceKey":"old"`) {
		t.Fatalf("preview %d %s %s", code, &out, &diag)
	}
	filter = "explicit"
	args = append(args, "--inventory", "explicit")
	if code := Run(context.Background(), args, env, &out, &diag); code != 0 || calls != 2 {
		t.Fatalf("filter %d %s", code, &diag)
	}
	retained, _ := os.ReadFile(filepath.Join(dir, "contexts.json"))
	if !bytes.Equal(data, retained) {
		t.Fatal("read changed remembered context")
	}
}
func TestArchiveBinaryTransferCommands(t *testing.T) {
	payload := bytes.Repeat([]byte("PK\x03\x04archive-bytes"), 70000)
	calls := 0
	status := 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Bearer owner" {
			w.WriteHeader(403)
			return
		}
		if status != 200 {
			w.WriteHeader(status)
			return
		}
		switch r.Method {
		case "POST":
			if r.URL.Path != "/tenants/home/archive-restores" || r.Header.Get("Content-Type") != "application/zip" || r.Header.Get("Idempotency-Key") != "same-key" {
				t.Errorf("upload route/headers %s", r.URL)
			}
			body, _ := io.ReadAll(r.Body)
			if !bytes.Equal(body, payload) {
				t.Error("upload bytes changed")
			}
			io.WriteString(w, `{"data":`+archiveFixture+`,"meta":{}}`)
		case "GET":
			if r.URL.Path != "/tenants/home/archive-jobs/job/content" {
				t.Errorf("download route %s", r.URL)
			}
			w.Header().Set("Content-Type", "application/zip")
			w.Header().Set("Content-Disposition", `attachment; filename="../../unsafe.zip"`)
			w.Write(payload)
		}
	}))
	defer server.Close()
	dir, env := portabilityEnvironment(t, server.URL)
	file := filepath.Join(dir, "input.zip")
	os.WriteFile(file, payload, 0600)
	var out, diag bytes.Buffer
	args := []string{"archive-jobs", "upload", "--tenant", "home", "--file", file, "--idempotency-key", "same-key", "--yes", "--no-input", "--json"}
	if code := Run(context.Background(), args, env, &out, &diag); code != 0 || calls != 1 {
		t.Fatalf("upload %d %s", code, &diag)
	}
	out.Reset()
	diag.Reset()
	args = []string{"archive-jobs", "download", "job", "--tenant", "home", "--output", "-", "--json", "--no-input"}
	beforeJSON := calls
	if code := Run(context.Background(), args, env, &out, &diag); code != 2 || calls != beforeJSON {
		t.Fatalf("JSON binary rejection: %d %s", code, &diag)
	}
	out.Reset()
	diag.Reset()
	args = []string{"archive-jobs", "download", "job", "--tenant", "home", "--output", "-", "--no-input"}
	if code := Run(context.Background(), args, env, &out, &diag); code != 0 || !bytes.Equal(out.Bytes(), payload) {
		t.Fatalf("binary stdout mixed %d %s", code, &diag)
	}
	target := filepath.Join(dir, "output.zip")
	args = append(args, "--output", target)
	out.Reset()
	if code := Run(context.Background(), args, env, &out, &diag); code != 0 {
		t.Fatalf("file %d %s", code, &diag)
	}
	saved, _ := os.ReadFile(target)
	if !bytes.Equal(saved, payload) {
		t.Fatal("saved bytes changed")
	}
	if code := Run(context.Background(), args, env, &out, &diag); code == 0 {
		t.Fatal("existing path overwritten")
	}
	for _, denial := range []int{401, 403, 307} {
		status = denial
		out.Reset()
		diag.Reset()
		before := calls
		args = []string{"archive-jobs", "download", "job", "--tenant", "home", "--output", "-", "--no-input"}
		if code := Run(context.Background(), args, env, &out, &diag); code == 0 || out.Len() != 0 || calls != before+1 {
			t.Fatal("download denial/retry")
		}
	}
}
