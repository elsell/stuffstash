package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/inputfiles"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/presentation"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/uploadtransfer"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func TestUncertainUploadIdentifiesAssetWithoutRepeatingWrites(t *testing.T) {
	var starts, transfers, completions int
	storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		transfers++
		if req.Header.Get("Authorization") != "" {
			t.Error("API credentials reached file storage")
		}
		body, _ := io.ReadAll(req.Body)
		if string(body) != "%PDF-1.7\ncontent" {
			t.Errorf("file bytes changed: %q", body)
		}
	}))
	defer storage.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Bearer owner" || req.Method != "POST" {
			t.Error("incorrect authenticated upload request")
		}
		switch req.URL.Path {
		case "/tenants/home/inventories/garage/assets/asset-123/attachments/direct-uploads":
			starts++
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"data": ports.DirectUpload{UploadID: "secret-upload-token", Method: "PUT", URL: storage.URL, Headers: map[string]string{"Content-Type": "application/pdf"}}})
		case "/tenants/home/inventories/garage/assets/asset-123/attachments/direct-uploads/secret-upload-token/complete":
			completions++
			// A completion response is lost after the server receives the mutation.
			connection, _, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			_ = connection.Close()
		default:
			t.Errorf("unexpected API path: %s", req.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	file := filepath.Join(t.TempDir(), "photo.pdf")
	if err := os.WriteFile(file, []byte("%PDF-1.7\ncontent"), 0600); err != nil {
		t.Fatal(err)
	}
	var diagnostic bytes.Buffer
	r := Runner{
		UploadFiles:    inputfiles.Files{},
		UploadTransfer: uploadtransfer.Client{HTTP: storage.Client(), AllowLoopbackHTTP: true},
		AttachmentUploads: func(server, token string) (ports.AttachmentUploads, error) {
			return httpapi.New(server, token, http.DefaultClient)
		},
		Output: presentation.Output{Stdout: io.Discard, Stderr: &diagnostic},
	}
	err := r.uploadAttachment(context.Background(), Options{Server: server.URL, Scope: ports.Scope{Tenant: "home", Inventory: "garage"}, Command: []string{"attachments", "upload", "asset-123"}, FilePath: file}, "owner")
	var failure *ports.Error
	if !errors.As(err, &failure) || failure.Category != "network" {
		t.Fatalf("expected uncertain network result: %v", err)
	}
	if !strings.Contains(failure.Message, `attachments list "asset-123"`) {
		t.Fatalf("recovery command must identify the uploaded asset: %s", failure.Message)
	}
	for _, scope := range []string{`--server "` + server.URL + `"`, `--tenant "home"`, `--inventory "garage"`} {
		if !strings.Contains(failure.Message, scope) {
			t.Errorf("recovery lost effective scope %s: %s", scope, failure.Message)
		}
	}
	if starts != 1 || transfers != 1 || completions != 1 {
		t.Fatalf("upload writes repeated: start=%d transfer=%d completion=%d", starts, transfers, completions)
	}
	if strings.Contains(failure.Message+diagnostic.String(), "secret-upload-token") || strings.Contains(failure.Message+diagnostic.String(), "Bearer owner") {
		t.Fatal("recovery output exposed a credential")
	}
}
