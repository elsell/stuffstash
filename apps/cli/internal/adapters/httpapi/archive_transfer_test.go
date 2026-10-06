package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestArchiveUploadStreamsWithoutReadingAhead(t *testing.T) {
	first := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tenants/home/archive-restores" || r.Header.Get("Authorization") != "Bearer owner" {
			w.WriteHeader(403)
			return
		}
		if r.Method != "POST" || r.Header.Get("Content-Type") != "application/zip" || r.Header.Get("Idempotency-Key") != "retry-key" {
			t.Error("wrong archive headers")
		}
		b := make([]byte, 1)
		if _, err := io.ReadFull(r.Body, b); err != nil {
			t.Error(err)
			return
		}
		close(first)
		rest, err := io.ReadAll(r.Body)
		if err != nil || b[0] != 80 || !bytes.Equal(rest, []byte{75, 0, 255}) {
			t.Error("changed archive")
		}
		io.WriteString(w, `{"data":{"id":"restore","createdAt":"2026-10-05T12:00:00Z","expiresAt":"2026-10-06T12:00:00Z","kind":"restore","state":"validating","phase":"validate","photos":true,"otherFiles":true},"meta":{"requestId":"trace"}}`)
	}))
	defer server.Close()
	api, _ := New(server.URL, "owner", server.Client())
	reader, writer := io.Pipe()
	defer reader.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go func() {
		defer writer.Close()
		if _, err := writer.Write([]byte{80}); err != nil {
			return
		}
		select {
		case <-first:
			writer.Write([]byte{75, 0, 255})
		case <-ctx.Done():
			writer.CloseWithError(ctx.Err())
		}
	}()
	result, err := api.UploadArchive(ctx, "home", "retry-key", reader)
	if err != nil || result.Data.ID != "restore" || result.Data.State != "validating" {
		t.Fatalf("upload: %+v %v", result, err)
	}
	if _, err := api.UploadArchive(ctx, "other", "retry-key", bytes.NewReader([]byte{80})); err == nil {
		t.Fatal("wrong household accepted")
	}
}
func TestArchiveDownloadBytesAndBoundaries(t *testing.T) {
	visited := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { visited = true }))
	defer target.Close()
	for _, status := range []int{200, 206, 302, 401, 403, 404} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/tenants/home/archive-jobs/job/content" || r.URL.Query().Get("inventoryId") != "garage" || r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					return
				}
				w.Header().Set("Content-Type", "application/zip")
				w.Header().Set("Content-Disposition", `attachment; filename="../../outside.zip"`)
				w.Header().Set("Location", target.URL)
				w.WriteHeader(status)
				w.Write([]byte{80, 75, 0, 255})
			}))
			defer server.Close()
			api, _ := New(server.URL, "owner", server.Client())
			for _, scope := range []ports.Scope{{Tenant: "home", Inventory: "garage"}, {Tenant: "other", Inventory: "garage"}, {Tenant: "home", Inventory: "other"}} {
				result, err := api.DownloadArchive(context.Background(), scope, "job")
				success := status == 200 && scope.Tenant == "home" && scope.Inventory == "garage"
				if !success {
					if err == nil {
						result.Body.Close()
						t.Fatal("error returned as archive")
					}
					continue
				}
				if err != nil {
					t.Fatal(err)
				}
				data, err := io.ReadAll(result.Body)
				result.Body.Close()
				if err != nil || !bytes.Equal(data, []byte{80, 75, 0, 255}) || result.ContentType != "application/zip" || result.ContentDisposition != `attachment; filename="../../outside.zip"` {
					t.Fatal("changed archive response")
				}
			}
		})
	}
	if visited {
		t.Fatal("followed archive redirect")
	}
}
