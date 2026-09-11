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
	"errors"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"regexp"
	"sync"
	"time"
)

var errAPNSCredentials = errors.New("APNs credentials are invalid")
var errAPNSAuthentication = errors.New("APNs authentication is unavailable")
var appleIdentifier = regexp.MustCompile(`^[A-Z0-9]{10}$`)

const apnsTokenLifetime = 50 * time.Minute

// APNSAuth keeps the provider key and token inside the adapter.
type APNSAuth struct {
	key           *ecdsa.PrivateKey
	keyID, teamID string
	clock         ports.Clock
	mu            sync.Mutex
	token         string
	issued        time.Time
}

func NewAPNSAuth(keyID, teamID string, keyPEM []byte, clock ports.Clock) (*APNSAuth, error) {
	if clock == nil || !appleIdentifier.MatchString(keyID) || !appleIdentifier.MatchString(teamID) {
		return nil, errAPNSCredentials
	}
	block, rest := pem.Decode(keyPEM)
	if block == nil || block.Type != "PRIVATE KEY" || len(rest) != 0 {
		return nil, errAPNSCredentials
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, errAPNSCredentials
	}
	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok || key.Curve != elliptic.P256() {
		return nil, errAPNSCredentials
	}
	return &APNSAuth{key: key, keyID: keyID, teamID: teamID, clock: clock}, nil
}
func (a *APNSAuth) Token() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := a.clock.Now()
	if now.IsZero() {
		return "", errAPNSAuthentication
	}
	if a.token != "" && !now.Before(a.issued) && now.Sub(a.issued) < apnsTokenLifetime {
		return a.token, nil
	}
	header, _ := json.Marshal(struct {
		Algorithm string `json:"alg"`
		KeyID     string `json:"kid"`
	}{"ES256", a.keyID})
	claims, _ := json.Marshal(struct {
		Issuer string `json:"iss"`
		Issued int64  `json:"iat"`
	}{a.teamID, now.Unix()})
	encoded := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(claims)
	hash := sha256.Sum256([]byte(encoded))
	r, s, err := ecdsa.Sign(rand.Reader, a.key, hash[:])
	if err != nil {
		return "", errAPNSAuthentication
	}
	signature := make([]byte, 64)
	r.FillBytes(signature[:32])
	s.FillBytes(signature[32:])
	a.token = encoded + "." + base64.RawURLEncoding.EncodeToString(signature)
	a.issued = now
	return a.token, nil
}

func (*APNSAuth) String() string   { return "APNSAuth(redacted)" }
func (*APNSAuth) GoString() string { return "APNSAuth(redacted)" }
