package dto

import (
	"bytes"
	"encoding/json"
	"github.com/danielgtaylor/huma/v2"
)

type Expiration struct {
	Date      string `json:"date" minLength:"7" maxLength:"10" doc:"Calendar date as YYYY-MM-DD or YYYY-MM"`
	Precision string `json:"precision" enum:"day,month"`
}

type ExpirationUpdate struct {
	present bool
	value   *Expiration
}

func (v *ExpirationUpdate) UnmarshalJSON(data []byte) error {
	v.present = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		v.value = nil
		return nil
	}
	return json.Unmarshal(data, &v.value)
}
func (v ExpirationUpdate) Present() bool      { return v.present }
func (v ExpirationUpdate) Value() *Expiration { return v.value }
func (v ExpirationUpdate) Schema(huma.Registry) *huma.Schema {
	return &huma.Schema{Type: huma.TypeObject, Nullable: true, AdditionalProperties: false, Required: []string{"date", "precision"}, Properties: map[string]*huma.Schema{
		"date": {Type: huma.TypeString}, "precision": {Type: huma.TypeString, Enum: []any{"day", "month"}},
	}}
}

// Huma requires explicit nullability for optional object response values.
func (AssetResponse) TransformSchema(_ huma.Registry, schema *huma.Schema) *huma.Schema {
	schema.Properties["expiration"] = &huma.Schema{AnyOf: []*huma.Schema{schema.Properties["expiration"], {Type: "null"}}}
	return schema
}

type ExpirationContext struct {
	State           string `json:"state" enum:"current,upcoming,expired"`
	TrackingEnabled bool   `json:"trackingEnabled"`
	AdvanceDays     int    `json:"advanceDays" minimum:"0" maximum:"3650"`
	Timezone        string `json:"timezone"`
}
