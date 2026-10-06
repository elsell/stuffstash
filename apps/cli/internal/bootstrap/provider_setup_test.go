package bootstrap

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProviderSetupExactInputAndSafeDenials(t *testing.T) {
	for _, action := range []string{"create", "update", "credential"} {
		t.Run(action, func(t *testing.T) {
			body := `{"capability":"language_inference","providerKind":"local_http","displayName":"Local","enable":false,"endpointUrl":null,"runtimeOptions":{"limit":9007199254740993},"capabilityMetadata":{"tools":false}}`
			method, path := "POST", "/tenants/home/provider-profiles"
			if action == "update" {
				body = `{"displayName":"New","promptTemplate":null,"runtimeOptions":{"limit":9007199254740993}}`
				method = "PATCH"
				path += "/profile"
			}
			if action == "credential" {
				body = `{"purpose":"api_key","credential":"secret-setup-value"}`
				method = "PUT"
				path += "/profile/credential"
			}
			calls, status := 0, 200
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != path || r.Method != method || r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					return
				}
				got, _ := io.ReadAll(r.Body)
				if string(got) != body {
					t.Errorf("changed input")
				}
				if status != 200 {
					w.WriteHeader(status)
					io.WriteString(w, "secret-setup-value")
					return
				}
				io.WriteString(w, `{"data":{"id":"profile","tenantId":"home","credentialStatus":"configured","credential":"secret-setup-value"}}`)
			}))
			defer server.Close()
			input := filepath.Join(t.TempDir(), "request.json")
			os.WriteFile(input, []byte(body), 0600)
			args := []string{"provider-profiles", action}
			if action != "create" {
				args = append(args, "profile")
			}
			args = append(args, "--tenant", "home", "--input", input, "--json", "--no-input")
			env := binaryEnvironment(t, server.URL)
			var out, diagnostic bytes.Buffer
			if code := Run(context.Background(), args, env, &out, &diagnostic); code == 0 || calls != 0 {
				t.Fatal("unconfirmed mutation")
			}
			args = append(args, "--yes")
			out.Reset()
			diagnostic.Reset()
			if code := Run(context.Background(), args, env, &out, &diagnostic); code != 0 {
				t.Fatalf("%d %s", code, &diagnostic)
			}
			if strings.Contains(out.String()+diagnostic.String(), "secret-setup-value") {
				t.Fatal("credential leaked")
			}
			for _, denied := range []int{401, 403, 409} {
				status = denied
				out.Reset()
				diagnostic.Reset()
				before := calls
				if code := Run(context.Background(), args, env, &out, &diagnostic); code == 0 || calls != before+1 || strings.Contains(diagnostic.String(), "secret-setup-value") {
					t.Fatal("unsafe rejection")
				}
			}
		})
	}
}
