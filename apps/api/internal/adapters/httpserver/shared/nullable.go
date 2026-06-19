package shared

import (
	"bytes"
	"encoding/json"

	"github.com/danielgtaylor/huma/v2"
)

type NullableString struct {
	present bool
	null    bool
	value   string
}

func (s *NullableString) UnmarshalJSON(data []byte) error {
	s.present = true
	if bytes.Equal(data, []byte("null")) {
		s.null = true
		s.value = ""
		return nil
	}
	s.null = false
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	s.value = value
	return nil
}

func (s NullableString) Schema(huma.Registry) *huma.Schema {
	return &huma.Schema{Type: huma.TypeString, Nullable: true}
}

func (s NullableString) Present() bool {
	return s.present
}

func (s NullableString) Null() bool {
	return s.null
}

func (s NullableString) Value() string {
	return s.value
}
