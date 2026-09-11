package expiration

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
)

type continuation struct {
	Cursor      string `json:"cursor"`
	Fingerprint string `json:"fingerprint"`
}

func queryFingerprint(scope string, query Query) string {
	encoded, _ := json.Marshal(struct {
		Scope string
		Query Query
	}{scope, query})
	hash := sha256.Sum256(encoded)
	return hex.EncodeToString(hash[:])
}
func EncodeContinuation(cursor, scope string, query Query) string {
	if cursor == "" {
		return ""
	}
	encoded, _ := json.Marshal(continuation{Cursor: cursor, Fingerprint: queryFingerprint(scope, query)})
	return base64.RawURLEncoding.EncodeToString(encoded)
}
func DecodeContinuation(token, scope string, query Query) (string, error) {
	if token == "" {
		return "", nil
	}
	if len(token) > 8192 {
		return "", ErrInvalidQuery
	}
	encoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return "", ErrInvalidQuery
	}
	var value continuation
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&value) != nil || value.Cursor == "" || value.Fingerprint != queryFingerprint(scope, query) {
		return "", ErrInvalidQuery
	}
	if decoder.Decode(new(any)) != io.EOF {
		return "", ErrInvalidQuery
	}
	return value.Cursor, nil
}
