package blobstore

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
)

func TestInternalBlobHTTPKeepsPublicUploadsHTTPS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Last-Modified", time.Unix(0, 0).UTC().Format(http.TimeFormat))
		w.Header().Set("Content-Length", "7")
		w.Header().Set("Content-Type", "image/jpeg")
		if r.Method != http.MethodHead {
			_, _ = w.Write([]byte("fixture"))
		}
	}))
	defer server.Close()
	publicSecure := true
	store, err := NewS3Store(S3Config{Endpoint: strings.TrimPrefix(server.URL, "http://"), PublicEndpoint: "uploads.example.test", Secure: false, PublicSecure: &publicSecure, AccessKey: "access", SecretKey: "secret", Bucket: "images", Region: "garage"})
	if err != nil {
		t.Fatal(err)
	}
	key, _ := media.NewStorageKey("tenant/inventory/asset/attachment")
	content, err := store.GetBlob(context.Background(), key)
	if err != nil || string(content) != "fixture" {
		t.Fatalf("internal read failed: %v", err)
	}
	upload, err := NewS3DirectAttachmentUploader(store).CreateDirectAttachmentUpload(context.Background(), directUploadRequest(t, time.Now().Add(time.Hour)))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(upload.URL, "https://uploads.example.test/") {
		t.Fatal("public upload lost HTTPS")
	}
}
