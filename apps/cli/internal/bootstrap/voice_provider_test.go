package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/contextfile"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
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

func TestVoiceProviderReadContract(t *testing.T) {
	profile := `{"capability":"inference","credentialPurpose":"model-access","credentialStatus":"configured","displayName":"Local model","id":"provider","lastTestedAt":"tested","lifecycleState":"active","modelName":"model","providerKind":"openai-compatible"}`
	configured := `{"tenantId":"home","readiness":"ready","updatedAt":"updated","profileIds":{"languageInference":"provider","speechToText":"stt","textToSpeech":"tts"},"slots":[{"capability":"inference","label":"Language model","readiness":"ready","recommendedAction":"none","selectedProfileId":"provider","selectionSource":"explicit","selectedProfile":` + profile + `,"duplicateProfiles":[` + profile + `],"issues":["review\u001b[31m"]}]}`
	variants := []string{configured, `{"tenantId":"home","readiness":"not-ready","profileIds":{},"slots":[{"capability":"stt","label":"Speech to text","readiness":"missing","recommendedAction":"select","selectionSource":"none","duplicateProfiles":null,"issues":null}]}`, `{"tenantId":"home","readiness":"not-ready","profileIds":{},"slots":null}`, `{"tenantId":"home","readiness":"not-ready","profileIds":{},"slots":[]}`, "null"}
	variants = append(variants, strings.ReplaceAll(strings.ReplaceAll(variants[1], `"duplicateProfiles":null`, `"duplicateProfiles":[]`), `"issues":null`, `"issues":[]`))
	status, calls := 200, 0
	data := configured
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if status != 200 {
			w.WriteHeader(status)
			io.WriteString(w, "private-provider-secret")
			return
		}
		if r.URL.Path != "/tenants/home/voice-provider-configuration" {
			w.WriteHeader(403)
			io.WriteString(w, "private-provider-secret")
			return
		}
		if r.Method != "GET" || r.URL.RawQuery != "" || r.Header.Get("Authorization") != "Bearer owner" || r.Header.Get("X-Request-ID") != "voice-read" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL)
		}
		safeData := strings.ReplaceAll(data, `"credentialStatus":"configured"`, `"credentialStatus":"configured","credential":"private-provider-secret"`)
		io.WriteString(w, `{"$schema":"voice-schema","data":`+safeData+`,"meta":{"requestId":"trace","tenantId":"home"}}`)
	}))
	defer server.Close()
	dir := t.TempDir()
	os.Chmod(dir, 0700)
	store := credentials.File{Path: filepath.Join(dir, "session.json")}
	session := ports.Session{Server: server.URL, Issuer: "https://identity.example", Subject: "owner", IDToken: "owner", ExpiresAt: time.Now().Add(time.Hour)}
	if err := store.Save(context.Background(), session); err != nil {
		t.Fatal(err)
	}
	getenv := func(k string) string {
		switch k {
		case "STUFF_STASH_CLI_SERVER":
			return server.URL
		case "STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP":
			return "true"
		case "STUFF_STASH_CLI_CREDENTIAL_FILE":
			return store.Path
		case "STUFF_STASH_CLI_CONFIG_FILE":
			return filepath.Join(dir, "contexts.json")
		}
		return ""
	}
	args := []string{"voice-provider", "show", "--tenant", "home", "--no-input", "--request-id", "voice-read"}
	var out, diag bytes.Buffer
	for _, variant := range variants {
		data = variant
		out.Reset()
		diag.Reset()
		if code := Run(context.Background(), append(args, "--json"), getenv, &out, &diag); code != 0 {
			t.Fatalf("read %d %s", code, &diag)
		}
		var envelope struct {
			Schema string          `json:"$schema"`
			Data   json.RawMessage `json:"data"`
			Meta   ports.Metadata  `json:"meta"`
		}
		if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		decode := func(raw []byte) any {
			var v any
			if err := json.Unmarshal(raw, &v); err != nil {
				t.Fatal(err)
			}
			return v
		}
		if !reflect.DeepEqual(decode(envelope.Data), decode([]byte(variant))) || envelope.Schema != "voice-schema" || envelope.Meta.RequestID == nil || *envelope.Meta.RequestID != "trace" {
			t.Fatalf("lost contract %s", &out)
		}
		if strings.Contains(out.String()+diag.String(), "private-provider-secret") {
			t.Fatal("secret leaked")
		}
		out.Reset()
		diag.Reset()
		if code := Run(context.Background(), args, getenv, &out, &diag); code != 0 || out.Len() == 0 || strings.ContainsAny(out.String(), "\x1b") || strings.Contains(out.String()+diag.String(), "private-provider-secret") {
			t.Fatalf("human output %d %s %s", code, &out, &diag)
		}
		if variant == configured {
			for _, field := range []string{"home", "Local model", "model-access", "openai-compatible", "Language model", "review", "tts", "updated"} {
				if !strings.Contains(out.String(), field) {
					t.Fatalf("missing human %s: %s", field, &out)
				}
			}
		}
	}
	for _, denial := range []int{401, 403, 404} {
		status = denial
		out.Reset()
		diag.Reset()
		if code := Run(context.Background(), append(args, "--json"), getenv, &out, &diag); code == 0 || strings.Contains(out.String()+diag.String(), "private-provider-secret") {
			t.Fatal("unsafe denial")
		}
	}
	status = 200
	before := calls
	if code := Run(context.Background(), append(args, "--tenant", "other"), getenv, &out, &diag); code == 0 || calls != before+1 {
		t.Fatal("cross household succeeded")
	}
	for _, flag := range []string{"--limit", "--cursor", "--title", "--input"} {
		before = calls
		if code := Run(context.Background(), append(args, flag, "bad"), getenv, &out, &diag); code != 2 || calls != before {
			t.Fatalf("unsupported %s reached API", flag)
		}
	}
	missing := []string{"voice-provider", "show", "--no-input", "--request-id", "voice-read"}
	before = calls
	if code := Run(context.Background(), missing, getenv, &out, &diag); code != 2 || calls != before {
		t.Fatal("missing household reached API")
	}
	manager := contexts.Manager{Store: contextfile.Store{Path: filepath.Join(dir, "contexts.json")}}
	if err := manager.Remember(context.Background(), contexts.Entry{Name: "home", Server: server.URL, Principal: contexts.Principal(session), Tenant: "home"}); err != nil {
		t.Fatal(err)
	}
	data = "null"
	out.Reset()
	diag.Reset()
	if code := Run(context.Background(), missing, getenv, &out, &diag); code != 0 || !strings.Contains(out.String(), "No voice provider configuration") {
		t.Fatalf("saved context %d %s %s", code, &out, &diag)
	}
}
