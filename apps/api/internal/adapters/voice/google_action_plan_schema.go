package voice

func geminiActionPlanResponseSchema() *geminiSchema {
	return &geminiSchema{
		Type: "object",
		Properties: map[string]geminiSchema{
			"actionPlan": {
				Type: "object",
				Properties: map[string]geminiSchema{
					"intentSummary": {
						Type:        "string",
						Description: "Safe user-facing summary of what the user asked to change.",
					},
					"modelInterpretationSummary": {
						Type:        "string",
						Description: "Safe user-facing summary of how the request was interpreted.",
					},
					"confirmationSummary": {
						Type:        "string",
						Description: "Short confirmation question shown to the user.",
					},
					"commands": {
						Type:        "array",
						Description: "Ordered commands to review and execute only after user approval.",
						Items:       geminiActionPlanCommandSchema(),
					},
					"riskSummary": {
						Type:        "string",
						Description: "Optional short safe risk summary.",
					},
				},
				Required: []string{"intentSummary", "modelInterpretationSummary", "confirmationSummary", "commands"},
			},
		},
		Required: []string{"actionPlan"},
	}
}

func geminiActionPlanCommandSchema() *geminiSchema {
	return &geminiSchema{
		Type: "object",
		Properties: map[string]geminiSchema{
			"id": {
				Type:        "string",
				Description: "Stable command id used by later commands as parentCommandId.",
			},
			"kind": {
				Type:        "string",
				Description: "Action-plan command kind.",
				Enum:        []string{"create_asset", "create_location", "move_asset", "archive_asset", "restore_asset"},
			},
			"summary": {
				Type:        "string",
				Description: "Short safe user-facing summary of this command.",
			},
			"arguments": {
				Type:        "object",
				Description: "Structured command arguments. Valid combinations are enforced by the Stuff Stash action-plan validator.",
				Properties: map[string]geminiSchema{
					"assetId": {
						Type:        "string",
						Description: "Existing asset id returned by a read tool for move, archive, or restore commands.",
					},
					"title": {
						Type:        "string",
						Description: "New or updated asset title.",
					},
					"name": {
						Type:        "string",
						Description: "Compatibility alias for title.",
					},
					"kind": {
						Type:        "string",
						Description: "New asset kind for create_asset.",
						Enum:        []string{"item", "container", "location"},
					},
					"description": {
						Type:        "string",
						Description: "Optional user-provided description.",
					},
					"parentAssetId": {
						Type:        "string",
						Description: "Existing parent asset id returned by a read tool.",
					},
					"parentCommandId": {
						Type:        "string",
						Description: "Earlier create command id when placing this asset inside something created in the same plan.",
					},
				},
			},
		},
		Required: []string{"id", "kind", "summary", "arguments"},
	}
}
