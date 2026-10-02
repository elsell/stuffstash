package inventoryarchive

import (
	"bytes"
	"encoding/json"
	"io"
)

// Missing byte lengths must not be confused with an explicitly empty file.
func (m *Media) UnmarshalJSON(data []byte) error {
	var record struct {
		SHA256    *string `json:"sha256"`
		SizeBytes *int64  `json:"sizeBytes"`
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if d.Decode(&record) != nil || d.Decode(new(any)) != io.EOF || record.SHA256 == nil || record.SizeBytes == nil {
		return ErrInvalid
	}
	m.SHA256 = *record.SHA256
	m.SizeBytes = *record.SizeBytes
	if !validMedia(*m) {
		return ErrInvalid
	}
	return nil
}
