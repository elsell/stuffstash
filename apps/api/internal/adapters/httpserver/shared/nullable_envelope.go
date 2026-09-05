package shared

import "github.com/danielgtaylor/huma/v2"

// NullableSuccessEnvelope represents a successful optional resource query.
// Huma does not infer nullability for pointers to object schemas.
type NullableSuccessEnvelope[T any] struct {
	Data *T   `json:"data"`
	Meta Meta `json:"meta"`
}

func (NullableSuccessEnvelope[T]) TransformSchema(_ huma.Registry, schema *huma.Schema) *huma.Schema {
	schema.Properties["data"] = &huma.Schema{AnyOf: []*huma.Schema{
		schema.Properties["data"], {Type: "null"},
	}}
	return schema
}
