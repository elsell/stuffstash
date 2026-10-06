package bootstrap

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLabelRenderFullInputAndFileEnvelope(t *testing.T) {
	body := `{"$schema":"request-schema","format":"png","media":{"preset_id":"preset","version":4294967295,"width_micrometers":29000,"height_micrometers":90000,"margins_micrometers":{"left":0,"right":0,"top":0,"bottom":0},"resolution_dpi":300,"raster_width":300,"raster_height":1000,"orientation":"portrait","color_mode":"monochrome","cut_policy":"end","display_rotation":0},"template":{"id":"template","version":4294967295,"options":{"show_reference":false}}}`
	content := []byte("\x89PNG\r\n\x1a\nlabel")
	digest := sha256.Sum256(content)
	checksum := hex.EncodeToString(digest[:])
	response := `{"$schema":"render-schema","data":{"id":"render","selectionFingerprint":"selection","mediaFingerprint":"media","contentType":"image/png","sha256":"` + checksum + `","contentPath":"https://untrusted.invalid/steal","expiresAt":"2026-01-01T00:00:00Z","widthPixels":300,"heightPixels":1000,"displayRotation":0},"meta":{"requestId":"trace","tenantId":"home"}}`
	calls, status := 0, 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if status != 200 {
			w.WriteHeader(status)
			return
		}
		prefix := "/tenants/home/inventories/garage/"
		if !strings.HasPrefix(r.URL.Path, prefix) || r.Header.Get("Authorization") != "Bearer owner" {
			w.WriteHeader(403)
			return
		}
		switch r.URL.Path {
		case prefix + "assets/asset/label":
			if r.Method != "POST" {
				t.Error("identity method")
			}
			io.WriteString(w, `{"data":{},"meta":{}}`)
		case prefix + "assets/asset/label-renders":
			got, _ := io.ReadAll(r.Body)
			if r.Method != "POST" || string(got) != body {
				t.Errorf("changed render input: %s", got)
			}
			io.WriteString(w, response)
		case prefix + "label-renders/render/content":
			w.Header().Set("Content-Type", "image/png")
			w.Write(content)
		default:
			t.Errorf("explicit input fetched selection: %s", r.URL)
			w.WriteHeader(404)
		}
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

	input := filepath.Join(dir, "render.json")
	os.WriteFile(input, []byte(body), 0600)
	destination := filepath.Join(dir, "label.png")
	args := []string{"labels", "render", "asset", "--tenant", "home", "--inventory", "garage", "--input", input, "--output", destination, "--json", "--no-input"}
	var out, diag bytes.Buffer
	if code := Run(context.Background(), args, getenv, &out, &diag); code != 0 || calls != 3 {
		t.Fatalf("render: %d %s", code, &diag)
	}
	var result struct {
		Path, Format, SHA256 string
		Render               json.RawMessage
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	var expected, actual any
	json.Unmarshal([]byte(response), &expected)
	json.Unmarshal(result.Render, &actual)
	if !reflect.DeepEqual(expected, actual) || result.Path != destination || result.Format != "png" || result.SHA256 != checksum {
		t.Fatalf("render metadata lost: %s", &out)
	}
	saved, err := os.ReadFile(destination)
	if err != nil || !bytes.Equal(saved, content) {
		t.Fatalf("file: %v", err)
	}
	if code := Run(context.Background(), args, getenv, &out, &diag); code == 0 {
		t.Fatal("overwrote existing output")
	}
	for _, denial := range []int{401, 403, 409} {
		status = denial
		before := calls
		if code := Run(context.Background(), args, getenv, &out, &diag); code == 0 || calls != before+1 {
			t.Fatalf("denial/retry: %d %s", code, &diag)
		}
	}
	status = 200
	before := calls
	if code := Run(context.Background(), append(args, "--inventory", "other"), getenv, &out, &diag); code == 0 || calls != before+1 {
		t.Fatal("cross inventory render succeeded")
	}
	before = calls
	invalid := strings.Replace(body, "4294967295", "4294967296", 1)
	if err := os.WriteFile(input, []byte(invalid), 0600); err != nil {
		t.Fatal(err)
	}
	if code := Run(context.Background(), args, getenv, &out, &diag); code != 2 || calls != before {
		t.Fatalf("overflow reached API: %d %s", code, &diag)
	}

}
