package credentials

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"
	"unicode"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

const AES256GCMAlgorithm = ports.ProviderCredentialAlgorithmAES256GCM

type AESGCMSealerConfig struct {
	KeyID  string
	Key    []byte
	Random io.Reader
}

type AESGCMSealer struct {
	keyID  string
	aead   cipher.AEAD
	random io.Reader
}

func NewAESGCMSealer(cfg AESGCMSealerConfig) (AESGCMSealer, error) {
	keyID := strings.TrimSpace(cfg.KeyID)
	if keyID == "" || len(cfg.Key) != 32 {
		return AESGCMSealer{}, ports.ErrInvalidProviderCredential
	}
	block, err := aes.NewCipher(append([]byte{}, cfg.Key...))
	if err != nil {
		return AESGCMSealer{}, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return AESGCMSealer{}, err
	}
	randomSource := cfg.Random
	if randomSource == nil {
		randomSource = rand.Reader
	}
	return AESGCMSealer{keyID: keyID, aead: aead, random: randomSource}, nil
}

func NewAESGCMSealerFromBase64(keyID string, encodedKey string) (AESGCMSealer, error) {
	key, err := decodeBase64Key(encodedKey)
	if err != nil {
		return AESGCMSealer{}, ports.ErrInvalidProviderCredential
	}
	return NewAESGCMSealer(AESGCMSealerConfig{KeyID: keyID, Key: key})
}

func (s AESGCMSealer) SealProviderCredential(_ context.Context, scope ports.ProviderCredentialScope, raw []byte) (ports.SealedProviderCredential, error) {
	if err := validateScopeAndRaw(scope, raw); err != nil {
		return ports.SealedProviderCredential{}, err
	}
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := io.ReadFull(s.random, nonce); err != nil {
		return ports.SealedProviderCredential{}, err
	}
	ciphertext := s.aead.Seal(nil, nonce, raw, credentialAssociatedData(scope))
	return ports.SealedProviderCredential{
		KeyID:      s.keyID,
		Algorithm:  AES256GCMAlgorithm,
		Nonce:      nonce,
		Ciphertext: ciphertext,
	}, nil
}

func (s AESGCMSealer) UnsealProviderCredential(_ context.Context, scope ports.ProviderCredentialScope, sealed ports.SealedProviderCredential) ([]byte, error) {
	if err := validateScope(scope); err != nil {
		return nil, err
	}
	if sealed.KeyID != s.keyID || sealed.Algorithm != AES256GCMAlgorithm || len(sealed.Nonce) != s.aead.NonceSize() || len(sealed.Ciphertext) == 0 {
		return nil, ports.ErrInvalidProviderCredential
	}
	raw, err := s.aead.Open(nil, sealed.Nonce, sealed.Ciphertext, credentialAssociatedData(scope))
	if err != nil {
		return nil, ports.ErrInvalidProviderCredential
	}
	return raw, nil
}

func decodeBase64Key(encoded string) ([]byte, error) {
	trimmed := strings.TrimSpace(encoded)
	if trimmed == "" {
		return nil, ports.ErrInvalidProviderCredential
	}
	if key, err := base64.StdEncoding.DecodeString(trimmed); err == nil {
		return key, nil
	}
	key, err := base64.RawStdEncoding.DecodeString(trimmed)
	if err != nil {
		return nil, ports.ErrInvalidProviderCredential
	}
	return key, nil
}

func validateScopeAndRaw(scope ports.ProviderCredentialScope, raw []byte) error {
	if len(raw) == 0 {
		return ports.ErrInvalidProviderCredential
	}
	return validateScope(scope)
}

func validateScope(scope ports.ProviderCredentialScope) error {
	if strings.TrimSpace(scope.TenantID.String()) == "" ||
		strings.TrimSpace(scope.ProviderProfileID) == "" ||
		strings.TrimSpace(string(scope.Capability)) == "" ||
		strings.TrimSpace(string(scope.ProviderKind)) == "" ||
		strings.TrimSpace(string(scope.Purpose)) == "" {
		return ports.ErrInvalidProviderCredential
	}
	if containsControl(scope.TenantID.String()) ||
		containsControl(scope.ProviderProfileID) ||
		containsControl(string(scope.Capability)) ||
		containsControl(string(scope.ProviderKind)) ||
		containsControl(string(scope.Purpose)) {
		return ports.ErrInvalidProviderCredential
	}
	return nil
}

func credentialAssociatedData(scope ports.ProviderCredentialScope) []byte {
	payload, _ := json.Marshal(struct {
		TenantID          string `json:"tenantId"`
		ProviderProfileID string `json:"providerProfileId"`
		Capability        string `json:"capability"`
		ProviderKind      string `json:"providerKind"`
		Purpose           string `json:"purpose"`
	}{
		TenantID:          scope.TenantID.String(),
		ProviderProfileID: scope.ProviderProfileID,
		Capability:        string(scope.Capability),
		ProviderKind:      string(scope.ProviderKind),
		Purpose:           string(scope.Purpose),
	})
	return payload
}

func containsControl(value string) bool {
	for _, char := range value {
		if unicode.IsControl(char) {
			return true
		}
	}
	return false
}
