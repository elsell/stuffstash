package shared

import (
	"errors"
	"net/http"
	"testing"
)

func TestNewErrorEnvelopeSuppressesInternalErrorDetails(t *testing.T) {
	response := NewErrorEnvelope(http.StatusInternalServerError, "boom", errors.New("database password leaked"))
	envelope, ok := response.(*ErrorEnvelope)
	if !ok {
		t.Fatalf("expected ErrorEnvelope, got %T", response)
	}

	if envelope.BodyError.Message != "Internal server error." {
		t.Fatalf("expected safe internal message, got %q", envelope.BodyError.Message)
	}
	if len(envelope.BodyError.Details) != 0 {
		t.Fatalf("expected no internal error details, got %+v", envelope.BodyError.Details)
	}
}

func TestNewErrorEnvelopeKeepsClientErrorDetails(t *testing.T) {
	response := NewErrorEnvelope(http.StatusBadRequest, "Invalid request.", errors.New("name is required"))
	envelope, ok := response.(*ErrorEnvelope)
	if !ok {
		t.Fatalf("expected ErrorEnvelope, got %T", response)
	}

	if len(envelope.BodyError.Details) != 1 || envelope.BodyError.Details[0].Message != "name is required" {
		t.Fatalf("expected client error detail, got %+v", envelope.BodyError.Details)
	}
}
