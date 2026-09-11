package push

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type fcmOAuthTransport struct {
	calls int
	body  string
}

func (f *fcmOAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	f.calls++
	if r.URL.String() != "https://oauth2.googleapis.com/token" {
		panic("credentials sent to alternate endpoint")
	}
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(f.body))}, nil
}
func fcmServiceAccount(t *testing.T, endpoint string) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(map[string]string{"type": "service_account", "client_email": "sender@test-project.iam.gserviceaccount.com", "private_key": string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})), "token_uri": endpoint})
	return data
}
func TestFCMAuthCachesAndRejectsAlternateCredentialEndpoint(t *testing.T) {
	clock := &authClock{now: time.Now()}
	transport := &fcmOAuthTransport{body: `{"access_token":"test-access-token","token_type":"Bearer","expires_in":3600}`}
	sender, err := NewFCMAuth(fcmServiceAccount(t, "https://oauth2.googleapis.com/token"), clock, transport)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		token, err := sender.Token(context.Background())
		if err != nil || token != "test-access-token" {
			t.Fatalf("token unavailable: %v", err)
		}
	}
	if transport.calls != 1 {
		t.Fatal("token not cached")
	}
	clock.now = clock.now.Add(-time.Minute)
	if _, err := sender.Token(context.Background()); err != nil {
		t.Fatal(err)
	}
	if transport.calls != 2 {
		t.Fatal("backwards clock reused cache")
	}
	if _, err := NewFCMAuth(fcmServiceAccount(t, "https://outside.test/token"), clock, transport); err == nil {
		t.Fatal("credential endpoint override accepted")
	}
}

func TestFCMAuthRefreshAndWaiterCancellation(t *testing.T) {
	clock := &authClock{now: time.Now()}
	transport := blockingAPNSTransport{started: make(chan struct{})}
	auth, err := NewFCMAuth(fcmServiceAccount(t, fcmTokenEndpoint), clock, transport)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := auth.Token(ctx); done <- err }()
	<-transport.started
	waiter, stop := context.WithCancel(context.Background())
	stop()
	if _, err := auth.Token(waiter); err != context.Canceled {
		t.Fatalf("blocked cancelled waiter: %v", err)
	}
	cancel()
	if err := <-done; err != context.Canceled {
		t.Fatalf("refresh lost cancellation: %v", err)
	}
}
func TestFCMAuthRejectsBadProviderResponseWithoutLeaking(t *testing.T) {
	data := fcmServiceAccount(t, fcmTokenEndpoint)
	for _, body := range []string{`private-key-material`, `{"access_token":"private-key-material","token_type":"Bearer","expires_in":-1}`, `{"access_token":"","token_type":"Bearer","expires_in":3600}`} {
		auth, err := NewFCMAuth(data, &authClock{now: time.Now()}, &fcmOAuthTransport{body: body})
		if err != nil {
			t.Fatal(err)
		}
		token, err := auth.Token(context.Background())
		if err == nil || strings.Contains(err.Error(), "private-key-material") || token != "" {
			t.Fatal("bad credentials escaped")
		}
	}
}
