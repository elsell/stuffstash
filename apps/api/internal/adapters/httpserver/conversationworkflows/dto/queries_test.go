package dto

import (
	"reflect"
	"testing"

	"github.com/danielgtaylor/huma/v2"
)

func TestSelectionContractIncludesNullWithoutChangingSelectionObject(t *testing.T) {
	registry := huma.NewMapRegistry("#/components/schemas/", huma.DefaultSchemaNamer)
	body := registry.Schema(reflect.TypeOf(SelectionOutput{}.Body), false, "")
	data := body.Properties["data"]
	if data == nil || len(data.AnyOf) != 2 {
		t.Fatal("selection data must explicitly accept an object or null")
	}
	var object, null bool
	for _, branch := range data.AnyOf {
		if branch.Type == "null" {
			null = true
		}
		if branch.Ref != "" {
			schema := registry.SchemaFromRef(branch.Ref)
			object = schema.Type == "object" && !schema.Nullable && schema.Properties["workflowId"] != nil && schema.Properties["revisionId"] != nil
		}
	}
	if !object || !null {
		t.Fatal("selection contract lost its non-null object or null alternative")
	}
}
