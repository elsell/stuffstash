package shared

import (
	"errors"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/app"
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

func TestNewErrorEnvelopeUsesSafePayloadTooLargeVocabulary(t *testing.T) {
	response := NewErrorEnvelope(http.StatusRequestEntityTooLarge, "request body is too large limit=21 bytes")
	envelope, ok := response.(*ErrorEnvelope)
	if !ok {
		t.Fatalf("expected ErrorEnvelope, got %T", response)
	}

	if envelope.BodyError.Code != "payload_too_large" {
		t.Fatalf("expected payload_too_large code, got %q", envelope.BodyError.Code)
	}
	if envelope.BodyError.Message != "Request body too large." {
		t.Fatalf("expected safe payload-too-large message, got %q", envelope.BodyError.Message)
	}
}

func TestToHumaErrorSurfacesSafeImportSourceChangedPrecondition(t *testing.T) {
	previous := huma.NewError
	huma.NewError = NewErrorEnvelope
	t.Cleanup(func() {
		huma.NewError = previous
	})

	response := ToHumaError(app.ImportSourceChangedAfterPreviewError{})
	envelope, ok := response.(*ErrorEnvelope)
	if !ok {
		t.Fatalf("expected ErrorEnvelope, got %T", response)
	}
	if envelope.GetStatus() != http.StatusPreconditionFailed {
		t.Fatalf("expected status %d, got %d", http.StatusPreconditionFailed, envelope.GetStatus())
	}
	if envelope.BodyError.Code != "precondition_failed" {
		t.Fatalf("expected precondition_failed code, got %q", envelope.BodyError.Code)
	}
	if envelope.BodyError.Message != "Import source changed after preview. Preview the source again before starting the import." {
		t.Fatalf("expected actionable source-changed message, got %q", envelope.BodyError.Message)
	}
}

func TestToHumaErrorMapsSpecificApplicationErrorVocabulary(t *testing.T) {
	previous := huma.NewError
	huma.NewError = NewErrorEnvelope
	t.Cleanup(func() {
		huma.NewError = previous
	})

	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{name: "validation", err: app.ErrValidation, status: http.StatusBadRequest, code: "invalid_request"},
		{name: "conflict", err: app.ErrConflict, status: http.StatusConflict, code: "conflict"},
		{name: "precondition", err: app.ErrPrecondition, status: http.StatusPreconditionFailed, code: "precondition_failed"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response := ToHumaError(tc.err)
			envelope, ok := response.(*ErrorEnvelope)
			if !ok {
				t.Fatalf("expected ErrorEnvelope, got %T", response)
			}
			if envelope.GetStatus() != tc.status {
				t.Fatalf("expected status %d, got %d", tc.status, envelope.GetStatus())
			}
			if envelope.BodyError.Code != tc.code {
				t.Fatalf("expected code %q, got %q", tc.code, envelope.BodyError.Code)
			}
		})
	}
}
