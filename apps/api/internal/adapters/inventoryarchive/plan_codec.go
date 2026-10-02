package inventoryarchive

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"io"
)

// PlanCodec encodes private, short-lived server state, not the portable format.
// In particular, remapped destination scope and blob keys must survive approval.
type PlanCodec struct{}

func (PlanCodec) EncodePlan(ctx context.Context, plan ports.ArchiveRestorePlan, maxBytes int) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, ErrLimit
	}
	var result bytes.Buffer
	err := json.NewEncoder(&boundedWriter{ctx: ctx, w: &result, remaining: int64(maxBytes)}).Encode(plan)
	if err != nil {
		return nil, err
	}
	return result.Bytes(), nil
}
func (PlanCodec) DecodePlan(ctx context.Context, data []byte, maxBytes int) (ports.ArchiveRestorePlan, error) {
	var result ports.ArchiveRestorePlan
	if maxBytes <= 0 || len(data) > maxBytes {
		return result, ErrLimit
	}
	if err := uniqueJSON(ctx, data); err != nil {
		return result, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		return result, ErrInvalid
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return result, ErrInvalid
	}
	return result, nil
}
