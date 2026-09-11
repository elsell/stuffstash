package push

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"
)

type authClock struct{ now time.Time }

func (c *authClock) Now() time.Time { return c.now }
func TestAPNSAuthSignsAndRefreshesProviderTokens(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	clock := &authClock{now: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)}
	signer, err := NewAPNSAuth("KEY1234567", "TEAM123456", pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), clock)
	if err != nil {
		t.Fatal(err)
	}
	token, err := signer.Token()
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatal("invalid jwt")
	}
	headerBytes, _ := base64.RawURLEncoding.DecodeString(parts[0])
	var header map[string]string
	_ = json.Unmarshal(headerBytes, &header)
	if header["alg"] != "ES256" || header["kid"] != "KEY1234567" {
		t.Fatal("invalid jwt header")
	}
	payloadBytes, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var claims struct {
		Issuer string `json:"iss"`
		Issued int64  `json:"iat"`
	}
	_ = json.Unmarshal(payloadBytes, &claims)
	if claims.Issuer != "TEAM123456" || claims.Issued != clock.now.Unix() {
		t.Fatal("invalid jwt claims")
	}
	signature, _ := base64.RawURLEncoding.DecodeString(parts[2])
	hash := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if len(signature) != 64 || !ecdsa.Verify(&key.PublicKey, hash[:], new(big.Int).SetBytes(signature[:32]), new(big.Int).SetBytes(signature[32:])) {
		t.Fatal("invalid jwt signature")
	}
	cached, _ := signer.Token()
	if cached != token {
		t.Fatal("provider token not cached")
	}

	concurrent := make(chan string, 16)
	for range 16 {
		go func() {
			value, err := signer.Token()
			if err != nil {
				concurrent <- ""
				return
			}
			concurrent <- value
		}()
	}
	for range 16 {
		if value := <-concurrent; value != token {
			t.Fatal("concurrent request did not reuse token")
		}
	}
	clock.now = clock.now.Add(50 * time.Minute)
	refreshed, _ := signer.Token()
	if refreshed == token {
		t.Fatal("provider token not refreshed")
	}
	clock.now = clock.now.Add(-time.Hour)
	backwards, _ := signer.Token()
	if backwards == refreshed {
		t.Fatal("future token reused")
	}
}
func TestAPNSAuthRejectsMalformedCredentials(t *testing.T) {
	clock := &authClock{now: time.Now()}
	for _, value := range []struct {
		key, team string
		body      []byte
	}{{"invalid", "TEAM123456", []byte("private secret")}, {"KEY1234567", "TEAM123456", []byte("private secret")}} {
		if _, err := NewAPNSAuth(value.key, value.team, value.body, clock); err == nil || strings.Contains(err.Error(), "private secret") {
			t.Fatal("unsafe credential validation")
		}
	}
}
