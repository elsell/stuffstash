package pairingkeys

import (
	"crypto/ed25519"
	"crypto/rand"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type Keys struct{}
type key struct {
	public  ed25519.PublicKey
	private ed25519.PrivateKey
}

func (Keys) NewKey() (ports.PairingKey, error) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return key{public, private}, nil
}
func (k key) PublicKey() []byte { return append([]byte(nil), k.public...) }
func (k key) Sign(id, poll string) []byte {
	return ed25519.Sign(k.private, []byte("stuffstash-print-pairing-v1\n"+id+"\n"+poll))
}
