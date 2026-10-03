package pairingcrypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type Secrets struct{}

func (Secrets) NewToken() (string, error) {
	var value [32]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value[:]), nil
}
func (Secrets) NewUserCode() (string, error) {
	var value [5]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(value[:]), nil
}
func (Secrets) Digest(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}
func (s Secrets) Matches(value, digest string) bool {
	return subtle.ConstantTimeCompare([]byte(s.Digest(value)), []byte(digest)) == 1
}
func (Secrets) Verify(key []byte, id printing.PairingID, token string, signature []byte) bool {
	return len(key) == ed25519.PublicKeySize && ed25519.Verify(ed25519.PublicKey(key), printing.PairingProofMessage(id, token), signature)
}

var _ ports.PairingSecrets = Secrets{}
