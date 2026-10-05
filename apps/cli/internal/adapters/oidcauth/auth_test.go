package oidcauth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type testClock struct{}

func (testClock) Now() time.Time { return time.Now() }

type captureOutput struct{ messages []string }

func (o *captureOutput) Notice(s string) error { o.messages = append(o.messages, s); return nil }
func (*captureOutput) Result(any) error        { return nil }
func (*captureOutput) Error(string, string)    {}
func TestDeviceLoginValidatesIdentityAndNeverLeaksSecrets(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	for _, scenario := range []string{"success", "wrong-audience", "missing-id-token", "denied", "expired", "bad-signature", "missing-subject"} {
		t.Run(scenario, func(t *testing.T) {
			var server *httptest.Server
			subject := "user"
			if scenario == "missing-subject" {
				subject = ""
			}
			server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/.well-known/openid-configuration":
					json.NewEncoder(w).Encode(map[string]any{"issuer": server.URL, "authorization_endpoint": server.URL + "/auth", "token_endpoint": server.URL + "/token", "device_authorization_endpoint": server.URL + "/device", "jwks_uri": server.URL + "/keys", "id_token_signing_alg_values_supported": []string{"RS256"}})
				case "/keys":
					json.NewEncoder(w).Encode(map[string]any{"keys": []any{map[string]any{"kty": "RSA", "kid": "test", "alg": "RS256", "use": "sig", "n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()), "e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes())}}})
				case "/device":
					json.NewEncoder(w).Encode(map[string]any{"device_code": "secret-device-code", "user_code": "ABCD", "verification_uri": server.URL + "/approve", "expires_in": 20, "interval": 1})
				case "/token":
					if scenario == "denied" {
						w.WriteHeader(400)
						json.NewEncoder(w).Encode(map[string]string{"error": "access_denied"})
						return
					}
					audience := "cli"
					if scenario == "wrong-audience" {
						audience = "someone-else"
					}
					expiry := time.Now().Add(time.Hour)
					if scenario == "expired" {
						expiry = time.Now().Add(-time.Hour)
					}
					claims, _ := json.Marshal(map[string]any{"iss": server.URL, "sub": subject, "aud": audience, "exp": expiry.Unix(), "iat": time.Now().Unix()})
					encoded := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","kid":"test"}`)) + "." + base64.RawURLEncoding.EncodeToString(claims)
					digest := sha256.Sum256([]byte(encoded))
					signature, _ := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
					if scenario == "bad-signature" {
						signature[0] ^= 1
					}
					id := encoded + "." + base64.RawURLEncoding.EncodeToString(signature)
					response := map[string]any{"access_token": "secret-access-token", "token_type": "Bearer", "expires_in": 3600, "refresh_token": "secret-refresh-token"}
					if scenario != "missing-id-token" {
						response["id_token"] = id
					}
					json.NewEncoder(w).Encode(response)
				default:
					w.WriteHeader(404)
				}
			}))
			defer server.Close()
			output := &captureOutput{}
			adapter := Adapter{HTTP: server.Client(), Clock: testClock{}, Output: output}
			result, err := adapter.Login(context.Background(), "https://api.example", ports.AuthConfig{Issuer: server.URL, ClientID: "cli", Scopes: []string{"openid"}, LoginMethods: []string{"device_code"}}, true)
			if scenario == "success" {
				if err != nil || result.IDToken == "" || result.RefreshToken == "" || result.Subject != "user" || result.Issuer != server.URL {
					t.Fatalf("login failed: %v", err)
				}
				refreshed, refreshErr := adapter.Refresh(context.Background(), result)
				if refreshErr != nil || refreshed.Subject != result.Subject {
					t.Fatalf("valid refresh: %+v %v", refreshed, refreshErr)
				}
				legacy := result
				legacy.Subject = ""
				upgraded, upgradeErr := adapter.Refresh(context.Background(), legacy)
				if upgradeErr != nil || upgraded.Subject != result.Subject {
					t.Fatalf("legacy refresh did not bind verified identity: %v", upgradeErr)
				}
				subject = "different-user"
				if _, refreshErr := adapter.Refresh(context.Background(), upgraded); refreshErr == nil {
					t.Fatal("refresh accepted a different account")
				}
			} else if err == nil {
				t.Fatal("accepted invalid authentication")
			}
			combined := strings.Join(output.messages, " ")
			if err != nil {
				combined += err.Error()
			}
			for _, secret := range []string{"secret-device-code", "secret-access-token", "secret-refresh-token"} {
				if strings.Contains(combined, secret) {
					t.Fatal("secret leaked")
				}
			}
		})
	}
}
