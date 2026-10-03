package pairingcrypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"testing"
)

func TestProofCannotChangePairingTokenOrKey(t *testing.T) {
	key, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	crypto := Secrets{}
	signature := ed25519.Sign(private, printing.PairingProofMessage("one", "token"))
	if !crypto.Verify(key, "one", "token", signature) {
		t.Fatal("legitimate proof denied")
	}
	if crypto.Verify(key, "two", "token", signature) || crypto.Verify(key, "one", "other", signature) || crypto.Verify(nil, "one", "token", signature) || crypto.Verify(key, "one", "token", nil) {
		t.Fatal("forged proof accepted")
	}
	token, err := crypto.NewToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(token) < 40 || !crypto.Matches(token, crypto.Digest(token)) || crypto.Matches("wrong", crypto.Digest(token)) {
		t.Fatal("token integrity failure")
	}
}
