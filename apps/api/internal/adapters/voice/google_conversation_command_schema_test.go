package voice

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGoogleCommandSchemaMergesRepeatedShapesWithoutExtraBranches(t *testing.T) {
	object := func(field any) map[string]any {
		return map[string]any{"type": "object", "properties": map[string]any{"value": field}, "additionalProperties": false}
	}
	text := map[string]any{"type": "string", "description": "First guidance"}
	annotated := map[string]any{"type": "string", "description": "Other guidance"}
	result := googleMergeObjectShapes([]map[string]any{object(text), object(annotated), object(text)})
	field := result["properties"].(map[string]any)["value"].(map[string]any)
	if field["anyOf"] != nil || field["type"] != "string" {
		t.Fatal("annotations created validation alternatives")
	}
	nullable := map[string]any{"anyOf": []any{text, map[string]any{"type": "null"}}}
	result = googleMergeObjectShapes([]map[string]any{object(text), object(nullable), object(text), object(nullable)})
	field = result["properties"].(map[string]any)["value"].(map[string]any)
	alternatives, _ := field["anyOf"].([]any)
	if len(alternatives) != 2 {
		t.Fatalf("expected two unique alternatives, got %v", alternatives)
	}
	for _, raw := range alternatives {
		if raw.(map[string]any)["anyOf"] != nil {
			t.Fatal("nested redundant alternatives")
		}
	}
}

func TestGoogleCommandSchemaProjectsArgumentsWithoutMutatingContract(t *testing.T) {
	input := json.RawMessage(`{"type":"object","properties":{"commands":{"type":"array","maxItems":10,"items":{"anyOf":[{"type":"object","properties":{"kind":{"type":"string","enum":["create"]},"arguments":{"type":"object","properties":{"title":{"type":"string"},"expiration":{"type":"object"}},"required":["title"],"additionalProperties":false}},"required":["kind","arguments"],"additionalProperties":false},{"type":"object","properties":{"kind":{"type":"string","enum":["update"]},"arguments":{"type":"object","properties":{"assetId":{"type":"string"},"expiration":{"anyOf":[{"type":"object"},{"type":"null"}]}},"required":["assetId","expiration"],"additionalProperties":false}},"required":["kind","arguments"],"additionalProperties":false}]}}}}`)
	result, err := googleConversationParameters(input)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	_ = json.Unmarshal(result, &root)
	commands := root["properties"].(map[string]any)["commands"].(map[string]any)
	items := commands["items"].(map[string]any)
	if items["anyOf"] != nil {
		t.Fatal("nested command union retained")
	}
	props := items["properties"].(map[string]any)
	kinds := props["kind"].(map[string]any)["enum"].([]any)
	if len(kinds) != 2 {
		t.Fatal("lost command kind")
	}
	args := props["arguments"].(map[string]any)
	fields := args["properties"].(map[string]any)
	if fields["title"] == nil || fields["assetId"] == nil || fields["expiration"] == nil {
		t.Fatal("lost argument")
	}
	if required, _ := args["required"].([]any); len(required) != 0 {
		t.Fatal("required an argument of another command")
	}
	if !strings.Contains(string(result), `"type":"null"`) || !strings.Contains(string(result), "create requires title") {
		t.Fatal("lost null or conditional guidance")
	}
	if !strings.Contains(string(input), `"anyOf"`) || strings.Contains(string(input), "create requires") {
		t.Fatal("mutated original schema")
	}
}
