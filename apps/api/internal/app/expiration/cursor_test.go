package expiration

import (
	"encoding/base64"
	"testing"
)

func TestQueryContinuationBoundToFiltersAndRecipient(t *testing.T) {
	query := Query{Status: QueryUpcoming, TagKey: "medicine"}
	token := EncodeContinuation("asset-page", "recipient", query)
	cursor, err := DecodeContinuation(token, "recipient", query)
	if err != nil || cursor != "asset-page" {
		t.Fatal("cursor did not round trip")
	}
	if _, err := DecodeContinuation(token, "another", query); err == nil {
		t.Fatal("cross-recipient cursor accepted")
	}
	query.Status = QueryExpired
	if _, err := DecodeContinuation(token, "recipient", query); err == nil {
		t.Fatal("different query reused cursor")
	}
	if _, err := DecodeContinuation("malformed", "recipient", query); err == nil {
		t.Fatal("malformed cursor accepted")
	}
}

func TestQueryContinuationRejectsTrailingJSON(t *testing.T) {
	q := Query{Status: QueryAll}
	token := EncodeContinuation("page", "scope", q)
	raw, _ := base64.RawURLEncoding.DecodeString(token)
	raw = append(raw, []byte(` {}`)...)
	if _, err := DecodeContinuation(base64.RawURLEncoding.EncodeToString(raw), "scope", q); err == nil {
		t.Fatal("trailing JSON accepted")
	}
}
