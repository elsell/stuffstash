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

func TestProviderProfilesHouseholdReadContract(t *testing.T) {
	calls := 0
	denied := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if denied || r.Header.Get("Authorization") != "Bearer owner" || !strings.HasPrefix(r.URL.Path, "/tenants/home/provider-profiles") {
			w.WriteHeader(403)
			io.WriteString(w, "private-provider-secret")
			return
		}
		if r.Method != "GET" || r.URL.RawQuery != "" {
			t.Error("wrong request")
		}
		data := `{"id":"profile","tenantId":"home","displayName":"Local model","capability":"inference","providerKind":"openai-compatible","modelName":"model","endpointUrl":"https://model.example","credentialStatus":"configured","lifecycleState":"active","createdAt":"created","updatedAt":"updated","lastTestedAt":"tested","promptTemplate":"Identify items","runtimeOptions":{"limit":9007199254740993},"capabilityMetadata":{"tools":true},"credential":"private-provider-secret"}`
		if r.URL.Path == "/tenants/home/provider-profiles" {
			data = "[" + data + "]"
		} else if r.URL.Path != "/tenants/home/provider-profiles/profile" {
			t.Error("wrong profile route")
		}
		io.WriteString(w, `{"$schema":"profile-schema","data":`+data+`,"meta":{"requestId":"trace"}}`)
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

	for _, action := range [][]string{{"list"}, {"show", "profile"}} {
		command := append([]string{"provider-profiles"}, action...)
		command = append(command, "--tenant", "home", "--json", "--no-input")
		var out, diagnostic bytes.Buffer
		if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 {
			t.Fatalf("read: %d %s", code, &diagnostic)
		}
		for _, field := range []string{`"limit":9007199254740993`, `"tools":true`, `"promptTemplate":"Identify items"`, `"lastTestedAt":"tested"`, `"credentialStatus":"configured"`, `"requestId":"trace"`, `"$schema":"profile-schema"`} {
			if !strings.Contains(out.String(), field) {
				t.Fatalf("missing %s: %s", field, &out)
			}
		}
		if strings.Contains(out.String()+diagnostic.String(), "private-provider-secret") {
			t.Fatal("credential leaked")
		}
		human := []string{"provider-profiles"}
		human = append(human, action...)
		human = append(human, "--tenant", "home", "--no-input")
		out.Reset()
		diagnostic.Reset()
		if code := Run(context.Background(), human, getenv, &out, &diagnostic); code != 0 || !strings.Contains(out.String(), "Local model") || strings.Contains(out.String(), "private-provider-secret") {
			t.Fatalf("human output: %d %s %s", code, &out, &diagnostic)
		}
		denied = true
		out.Reset()
		diagnostic.Reset()
		if code := Run(context.Background(), command, getenv, &out, &diagnostic); code == 0 || strings.Contains(out.String()+diagnostic.String(), "private-provider-secret") {
			t.Fatal("denial unsafe")
		}
		denied = false
		before := calls
		if code := Run(context.Background(), append(command, "--tenant", "other"), getenv, &out, &diagnostic); code == 0 || calls != before+1 {
			t.Fatal("scope isolation")
		}
		before = calls
		if code := Run(context.Background(), append(command, "--cursor", "next"), getenv, &out, &diagnostic); code != 2 || calls != before {
			t.Fatal("unsupported cursor reached API")
		}
	}
}
