package voice

import (
	"encoding/json"
	"sort"
	"strings"
)

// Google receives a decoding shape; application parsers retain the stricter
// per-command contract. Only recognizable command envelopes are projected.
func googleConversationCommandUnion(schema map[string]any) {
	alternatives, ok := schema["anyOf"].([]any)
	if !ok || len(schema) != 1 || len(alternatives) < 2 {
		return
	}
	var objects, arguments []map[string]any
	var kinds []any
	var guidance []string
	for _, raw := range alternatives {
		object, ok := raw.(map[string]any)
		if !ok || !googleMergeableObject(object) {
			return
		}
		properties := object["properties"].(map[string]any)
		kind, ok := properties["kind"].(map[string]any)
		if !ok {
			return
		}
		values, ok := kind["enum"].([]any)
		if !ok || len(values) == 0 {
			return
		}
		args, ok := properties["arguments"].(map[string]any)
		if !ok || !googleMergeableObject(args) {
			return
		}
		objects = append(objects, object)
		arguments = append(arguments, args)
		for _, value := range values {
			name, ok := value.(string)
			if !ok || name == "" {
				return
			}
			kinds = append(kinds, name)
			guidance = append(guidance, name+" requires "+strings.Join(googleSchemaRequired(args), ", "))
		}
	}
	merged := googleMergeObjectShapes(objects)
	properties := merged["properties"].(map[string]any)
	properties["kind"] = map[string]any{"type": "string", "enum": kinds}
	args := googleMergeObjectShapes(arguments)
	args["description"] = strings.Join(guidance, "; ") + ". Use only fields appropriate to the selected kind; the application validates each command."
	properties["arguments"] = args
	delete(schema, "anyOf")
	for key, value := range merged {
		schema[key] = value
	}
}
func googleMergeableObject(schema map[string]any) bool {
	if schema["type"] != "object" || schema["additionalProperties"] != false {
		return false
	}
	if _, ok := schema["properties"].(map[string]any); !ok {
		return false
	}
	for key := range schema {
		switch key {
		case "type", "properties", "required", "additionalProperties", "description":
		default:
			return false
		}
	}
	return true
}
func googleSchemaRequired(schema map[string]any) []string {
	var result []string
	values, _ := schema["required"].([]any)
	for _, value := range values {
		if name, ok := value.(string); ok {
			result = append(result, name)
		}
	}
	sort.Strings(result)
	return result
}
func googleMergeObjectShapes(objects []map[string]any) map[string]any {
	properties := map[string]any{}
	requiredCounts := map[string]int{}
	for _, object := range objects {
		for _, name := range googleSchemaRequired(object) {
			requiredCounts[name]++
		}
		for name, value := range object["properties"].(map[string]any) {
			previous, found := properties[name]
			if !found {
				properties[name] = value
				continue
			}
			properties[name] = googleMergeSchemaAlternatives(previous, value)
		}
	}
	var required []string
	for name, count := range requiredCounts {
		if count == len(objects) {
			required = append(required, name)
		}
	}
	sort.Strings(required)
	result := map[string]any{"type": "object", "properties": properties, "additionalProperties": false}
	if len(required) > 0 {
		result["required"] = required
	}
	return result
}

func googleMergeSchemaAlternatives(values ...any) any {
	var alternatives []any
	seen := map[string]bool{}
	var collect func(any)
	collect = func(value any) {
		if node, ok := value.(map[string]any); ok {
			if branches, ok := node["anyOf"].([]any); ok && len(node) == 1 {
				for _, branch := range branches {
					collect(branch)
				}
				return
			}
		}
		// Descriptions guide the provider but do not distinguish valid values.
		// Remove only this node's annotation, preserving properties named description.
		identity := value
		if node, ok := value.(map[string]any); ok {
			copy := make(map[string]any, len(node))
			for key, field := range node {
				if key != "description" {
					copy[key] = field
				}
			}
			identity = copy
		}
		encoded, _ := json.Marshal(identity)
		if !seen[string(encoded)] {
			seen[string(encoded)] = true
			alternatives = append(alternatives, value)
		}
	}
	for _, value := range values {
		collect(value)
	}
	if len(alternatives) == 1 {
		return alternatives[0]
	}
	return map[string]any{"anyOf": alternatives}
}
