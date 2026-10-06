package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBinaryDownloadCommands(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Header.Get("Authorization") != "Bearer owner" || r.URL.Query().Get("variant") == "invalid" {
			w.WriteHeader(403)
			return
		}
		io.WriteString(w, "binary payload")
	}))
	defer server.Close()
	env := binaryEnvironment(t, server.URL)
	for _, command := range [][]string{{"attachments", "download", "asset", "photo"}, {"attachments", "thumbnail", "asset", "photo", "--variant", "small"}, {"labels", "download", "render"}, {"inventories", "export", "--format", "json"}} {
		var out, diagnostic bytes.Buffer
		args := append(command, "--tenant", "home", "--inventory", "garage", "--output", "-", "--no-input")
		if code := Run(context.Background(), args, env, &out, &diagnostic); code != 0 {
			t.Fatalf("%v: %d %s", command, code, diagnostic.String())
		}
		if out.String() != "binary payload" {
			t.Fatal(out.String())
		}
	}
	before := requests
	var out, diagnostic bytes.Buffer
	if code := Run(context.Background(), []string{"attachments", "download", "asset", "photo", "--tenant", "home", "--inventory", "garage", "--output", "-", "--json"}, env, &out, &diagnostic); code == 0 || requests != before {
		t.Fatal("JSON binary stdout accepted")
	}
}
func binaryEnvironment(t *testing.T, server string) func(string) string {
	t.Helper()
	dir := t.TempDir()
	os.Chmod(dir, 0700)
	store := credentials.File{Path: filepath.Join(dir, "session.json")}
	if err := store.Save(context.Background(), ports.Session{Server: server, Issuer: "https://id.example", Subject: "owner", IDToken: "owner", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	return func(k string) string {
		switch k {
		case "STUFF_STASH_CLI_SERVER":
			return server
		case "STUFF_STASH_CLI_CREDENTIAL_FILE":
			return store.Path
		case "STUFF_STASH_CLI_CONFIG_FILE":
			return filepath.Join(dir, "config", "contexts.json")
		case "STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP":
			return "true"
		}
		return ""
	}
}
func TestDirectUploadNoCredentialsAndFailureStopsCompletion(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "failure"}[fail], func(t *testing.T) {
			completed := false
			storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "" {
					t.Error("API credential leaked")
				}
				data, _ := io.ReadAll(r.Body)
				if string(data) != "%PDF-1.7\ncontent" {
					t.Errorf("body %q", data)
				}
				if fail {
					w.WriteHeader(500)
				}
			}))
			defer storage.Close()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(401)
					return
				}
				if r.URL.Path == "/tenants/home/inventories/garage/assets/asset/attachments/direct-uploads" {
					var input map[string]any
					if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
						t.Fatal(err)
					}
					w.WriteHeader(201)
					json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"uploadId": "upload", "attachmentId": "photo", "method": "PUT", "url": storage.URL, "headers": map[string]string{"Content-Type": "application/pdf"}}})
					return
				}
				completed = true
				w.WriteHeader(201)
				io.WriteString(w, `{"data":{"id":"photo","fileName":"test.pdf","contentType":"application/pdf","sizeBytes":16}}`)
			}))
			defer server.Close()
			file := filepath.Join(t.TempDir(), "test.pdf")
			os.WriteFile(file, []byte("%PDF-1.7\ncontent"), 0600)
			var out, diagnostic bytes.Buffer
			code := Run(context.Background(), []string{"attachments", "upload", "asset", "--file", file, "--tenant", "home", "--inventory", "garage", "--json", "--no-input"}, binaryEnvironment(t, server.URL), &out, &diagnostic)
			if (code == 0) == fail || completed == fail {
				t.Fatalf("code=%d completed=%v: %s", code, completed, diagnostic.String())
			}
		})
	}
}

func TestAPIUploadAndDownloadDenial(t *testing.T) {
	uploaded := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.WriteHeader(403)
			io.WriteString(w, `{"error":{"code":"forbidden","message":"secret upstream details"}}`)
			return
		}
		var body struct{ FileName, ContentType, ContentBase64 string }
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body.FileName != "paper.pdf" || body.ContentType != "application/pdf" || body.ContentBase64 != "JVBERi0xLjdcblRlc3Q=" {
			t.Errorf("unexpected payload: %+v", body)
		}
		uploaded = true
		w.WriteHeader(201)
		io.WriteString(w, `{"data":{"id":"paper","fileName":"paper.pdf","contentType":"application/pdf","sizeBytes":14}}`)
	}))
	defer server.Close()
	env := binaryEnvironment(t, server.URL)
	p := filepath.Join(t.TempDir(), "paper.pdf")
	os.WriteFile(p, []byte(`%PDF-1.7\nTest`), 0600)
	var out, diagnostic bytes.Buffer
	if code := Run(context.Background(), []string{"attachments", "upload", "asset", "--file", p, "--transfer", "api", "--tenant", "home", "--inventory", "garage", "--no-input"}, env, &out, &diagnostic); code != 0 || !uploaded {
		t.Fatalf("%d %s", code, diagnostic.String())
	}
	out.Reset()
	diagnostic.Reset()
	dest := filepath.Join(t.TempDir(), "denied")
	if code := Run(context.Background(), []string{"attachments", "download", "asset", "private", "--tenant", "other", "--inventory", "garage", "--output", dest, "--no-input"}, env, &out, &diagnostic); code == 0 {
		t.Fatal("allowed denied download")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatal("denied download published file")
	}
}

func TestUploadRejectionPreservesAuthorizationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, `{"error":{"code":"forbidden","message":"private server reason"}}`)
	}))
	defer server.Close()
	file := filepath.Join(t.TempDir(), "paper.pdf")
	os.WriteFile(file, []byte("%PDF-1.7\ncontent"), 0600)
	var out, diagnostic bytes.Buffer
	code := Run(context.Background(), []string{"attachments", "upload", "asset", "--file", file, "--transfer", "api", "--tenant", "home", "--inventory", "garage", "--json", "--no-input"}, binaryEnvironment(t, server.URL), &out, &diagnostic)
	var result struct {
		Error struct{ Category, Message string }
	}
	if err := json.Unmarshal(diagnostic.Bytes()[bytes.LastIndex(diagnostic.Bytes(), []byte("{\"error\"")):], &result); err != nil {
		t.Fatal(err)
	}
	if code == 0 || result.Error.Category != "forbidden" {
		t.Fatalf("code=%d category=%s diagnostic=%s", code, result.Error.Category, diagnostic.String())
	}
}
