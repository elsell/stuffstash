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

func TestPrintingCatalogContracts(t *testing.T) {
	empty := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prefix := "/tenants/home/inventories/garage/"
		if r.Header.Get("Authorization") != "Bearer owner" || !strings.HasPrefix(r.URL.Path, prefix) {
			w.WriteHeader(403)
			return
		}
		if r.Method != "GET" || r.URL.RawQuery != "" {
			t.Error("wrong request")
		}
		data := `[{"id":"template","version":2,"name":"Label","purpose":"asset","font":"Noto","glyphCoverage":"Latin","minimumQRModulePixels":4,"defaults":{"show_reference":true},"options":["showReference"]}]`
		if strings.HasSuffix(r.URL.Path, "printer-profiles") {
			data = `[{"adapterId":"brother","name":"QL","transport":"usb","physicallyVerified":true,"supportedPlatforms":["linux"],"media":[{"presetId":"preset","name":"Label","version":2,"widthMicrometers":62000,"heightMicrometers":29000,"marginsMicrometers":{"top":1,"bottom":2,"left":3,"right":4},"resolutionDpi":300,"rasterWidth":732,"rasterHeight":342,"displayRotation":90,"colorMode":"monochrome","cutPolicy":"cut","orientation":"landscape"}]}]`
		}
		if empty != "" {
			data = empty
		}
		io.WriteString(w, `{"$schema":"catalog-schema","data":`+data+`,"meta":{"requestId":"trace"}}`)
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

	for _, action := range [][]string{{"labels", "templates"}, {"printers", "profiles"}} {
		command := append(append([]string{}, action...), "--tenant", "home", "--inventory", "garage", "--json", "--no-input")
		var out, diagnostic bytes.Buffer
		if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 {
			t.Fatalf("catalog: %d %s", code, &diagnostic)
		}
		fields := []string{`"defaults":{"show_reference":true}`, `"options":["showReference"]`, `"font":"Noto"`, `"glyphCoverage":"Latin"`, `"minimumQRModulePixels":4`}
		if action[0] == "printers" {
			fields = []string{`"transport":"usb"`, `"physicallyVerified":true`, `"supportedPlatforms":["linux"]`, `"widthMicrometers":62000`, `"right":4`, `"presetId":"preset"`}
		}
		for _, field := range append(fields, `"requestId":"trace"`, `"$schema":"catalog-schema"`) {
			if !strings.Contains(out.String(), field) {
				t.Fatalf("missing %s: %s", field, &out)
			}
		}
		for _, override := range [][]string{{"--tenant", "other"}, {"--inventory", "other"}} {
			if code := Run(context.Background(), append(command, override...), getenv, &out, &diagnostic); code == 0 {
				t.Fatal("scope bypass")
			}
		}
		for _, value := range []string{"null", "[]"} {
			empty = value
			out.Reset()
			diagnostic.Reset()
			if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 || !strings.Contains(out.String(), `"data":`+value) {
				t.Fatalf("changed empty catalog: %d %s", code, &out)
			}
		}
		empty = ""
	}
}
