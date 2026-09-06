package voice

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

// Google expands bounded arrays of alternatives into excessive decoding states.
// Keep the command shapes, but leave this count to the application validator.
// Work on a decoded copy: the shared catalog and other providers retain bounds.
func googleConversationEnvelopeParameters(parameters json.RawMessage) (json.RawMessage, error) {
	var schema map[string]any
	if json.Unmarshal(parameters, &schema) != nil || schema == nil {
		return nil, ports.ErrInvalidProviderInput
	}
	googleConversationSchemaArrayLimits(schema)
	return json.Marshal(schema)
}

func googleConversationSchemaArrayLimits(schema map[string]any) {
	if schema["type"] == "array" {
		items, _ := schema["items"].(map[string]any)
		alternatives, _ := items["anyOf"].([]any)
		if maximum, exists := schema["maxItems"]; exists && len(alternatives) > 0 {
			description, _ := schema["description"].(string)
			schema["description"] = strings.TrimSpace(description + fmt.Sprintf(" At most %v entries.", maximum))
			delete(schema, "maxItems")
		}
	}
	// Traverse schema locations only, never names or data inside examples/enums.
	for _, keyword := range []string{"properties", "$defs", "definitions"} {
		children, _ := schema[keyword].(map[string]any)
		for _, child := range children {
			if node, ok := child.(map[string]any); ok {
				googleConversationSchemaArrayLimits(node)
			}
		}
	}
	for _, keyword := range []string{"items", "additionalProperties"} {
		if node, ok := schema[keyword].(map[string]any); ok {
			googleConversationSchemaArrayLimits(node)
		}
	}
	for _, keyword := range []string{"anyOf", "oneOf", "allOf"} {
		children, _ := schema[keyword].([]any)
		for _, child := range children {
			if node, ok := child.(map[string]any); ok {
				googleConversationSchemaArrayLimits(node)
			}
		}
	}
}
