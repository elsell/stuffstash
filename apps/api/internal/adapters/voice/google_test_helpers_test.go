package voice

import (
	"testing"

	"golang.org/x/oauth2"
)

func roleAt(t *testing.T, content any) string {
	t.Helper()
	item := objectFromAny(t, content)
	role, ok := item["role"].(string)
	if !ok {
		t.Fatalf("content role missing or wrong type: %+v", item)
	}
	return role
}

func partObjectAt(t *testing.T, content any, partIndex int, key string) map[string]any {
	t.Helper()
	item := objectFromAny(t, content)
	parts, ok := item["parts"].([]any)
	if !ok {
		t.Fatalf("content parts missing or wrong type: %+v", item)
	}
	part := objectFromAny(t, parts[partIndex])
	return objectAt(t, part, key)
}

func objectAt(t *testing.T, item map[string]any, key string) map[string]any {
	t.Helper()
	return objectFromAny(t, item[key])
}

func objectFromAny(t *testing.T, value any) map[string]any {
	t.Helper()
	item, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("value is not an object: %+v", value)
	}
	return item
}

func requestHasFunctionDeclaration(request map[string]any, name string) bool {
	tools, ok := request["tools"].([]any)
	if !ok {
		return false
	}
	for _, rawTool := range tools {
		tool, ok := rawTool.(map[string]any)
		if !ok {
			continue
		}
		declarations, ok := tool["functionDeclarations"].([]any)
		if !ok {
			continue
		}
		for _, rawDeclaration := range declarations {
			declaration, ok := rawDeclaration.(map[string]any)
			if ok && declaration["name"] == name {
				return true
			}
		}
	}
	return false
}

func generationConfigHasFinalResponseSchema(config map[string]any) bool {
	schema, ok := config["responseJsonSchema"].(map[string]any)
	if !ok {
		schema, ok = config["responseSchema"].(map[string]any)
	}
	if !ok || schema["type"] != "object" {
		return false
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return false
	}
	final, ok := properties["final"].(map[string]any)
	if !ok || final["type"] != "object" {
		return false
	}
	finalProperties, ok := final["properties"].(map[string]any)
	if !ok {
		return false
	}
	kind, ok := finalProperties["kind"].(map[string]any)
	if !ok {
		return false
	}
	enum, ok := kind["enum"].([]any)
	if !ok || len(enum) == 0 {
		return false
	}
	_, hasSpoken := finalProperties["spokenResponse"].(map[string]any)
	_, hasDisplay := finalProperties["displayResponse"].(map[string]any)
	return hasSpoken && hasDisplay
}

func generationConfigHasActionPlanSchema(config map[string]any) bool {
	schema, ok := config["responseSchema"].(map[string]any)
	if !ok {
		schema, ok = config["responseJsonSchema"].(map[string]any)
	}
	if !ok || schema["type"] != "object" {
		return false
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return false
	}
	actionPlan, ok := properties["actionPlan"].(map[string]any)
	if !ok || actionPlan["type"] != "object" {
		return false
	}
	actionPlanProperties, ok := actionPlan["properties"].(map[string]any)
	if !ok {
		return false
	}
	commands, ok := actionPlanProperties["commands"].(map[string]any)
	if !ok || commands["type"] != "array" {
		return false
	}
	items, ok := commands["items"].(map[string]any)
	if !ok {
		return false
	}
	branches, ok := items["anyOf"].([]any)
	if !ok || len(branches) < 3 {
		return false
	}
	createBranch, ok := branches[0].(map[string]any)
	if !ok {
		return false
	}
	createProperties, ok := createBranch["properties"].(map[string]any)
	if !ok {
		return false
	}
	createKind, ok := createProperties["kind"].(map[string]any)
	if !ok {
		return false
	}
	createEnum, _ := createKind["enum"].([]any)
	createArguments, ok := createProperties["arguments"].(map[string]any)
	if !ok {
		return false
	}
	createArgumentProperties, ok := createArguments["properties"].(map[string]any)
	if !ok {
		return false
	}
	moveBranch, ok := branches[2].(map[string]any)
	if !ok {
		return false
	}
	moveProperties, ok := moveBranch["properties"].(map[string]any)
	if !ok {
		return false
	}
	moveKind, ok := moveProperties["kind"].(map[string]any)
	if !ok {
		return false
	}
	moveEnum, _ := moveKind["enum"].([]any)
	moveArguments, ok := moveProperties["arguments"].(map[string]any)
	if !ok {
		return false
	}
	moveArgumentProperties, ok := moveArguments["properties"].(map[string]any)
	if !ok {
		return false
	}
	_, hasIntent := actionPlanProperties["intentSummary"].(map[string]any)
	_, hasInterpretation := actionPlanProperties["modelInterpretationSummary"].(map[string]any)
	_, hasConfirmation := actionPlanProperties["confirmationSummary"].(map[string]any)
	_, hasTitle := createArgumentProperties["title"].(map[string]any)
	_, hasCreateParentCommandID := createArgumentProperties["parentCommandId"].(map[string]any)
	_, hasAssetID := moveArgumentProperties["assetId"].(map[string]any)
	_, hasMoveParentCommandID := moveArgumentProperties["parentCommandId"].(map[string]any)
	return hasIntent && hasInterpretation && hasConfirmation &&
		len(createEnum) == 1 && createEnum[0] == "create_asset" &&
		len(moveEnum) == 1 && moveEnum[0] == "move_asset" &&
		hasTitle && hasCreateParentCommandID && hasAssetID && hasMoveParentCommandID
}

func geminiFunctionCallResponse(name string, args map[string]any) map[string]any {
	return map[string]any{
		"candidates": []map[string]any{{
			"content": map[string]any{
				"parts": []map[string]any{{
					"functionCall": map[string]any{
						"name": name,
						"args": args,
					},
				}},
			},
		}},
	}
}

type staticTokenSource struct{}

func (staticTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: "test-token", TokenType: "Bearer"}, nil
}
