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
		AnyOf: []geminiSchema{
			geminiCreateAssetCommandSchema(),
			geminiCreateLocationCommandSchema(),
			geminiMoveAssetCommandSchema(),
			geminiAssetIDOnlyCommandSchema("archive_asset", "Archive an existing asset returned by a read tool."),
			geminiAssetIDOnlyCommandSchema("restore_asset", "Restore an existing asset returned by a read tool."),
		},
	}
}

func geminiCreateAssetCommandSchema() geminiSchema {
	return geminiActionPlanCommandBranch("create_asset", "Create a new item, container, or location-like asset. Never include assetId in arguments.", geminiSchema{
		Type:        "object",
		Description: "Arguments for create_asset. Use this for new items and containers. Set parentAssetId to an existing visible parent id, or parentCommandId to an earlier create command id. Set the unused parent reference to an empty string. Never include assetId.",
		Properties: map[string]geminiSchema{
			"title": {
				Type:        "string",
				Description: "Title of the new asset.",
			},
			"kind": {
				Type:        "string",
				Description: "New asset kind.",
				Enum:        []string{"item", "container", "location"},
			},
			"description": {
				Type:        "string",
				Description: "Optional user-provided description.",
			},
			"parentAssetId": {
				Type:        "string",
				Description: "Existing parent asset id copied exactly from a read-tool result, or empty string when parentCommandId is used.",
			},
			"parentCommandId": {
				Type:        "string",
				Description: "Earlier create command id when the parent is created in this same plan, or empty string when parentAssetId is used.",
			},
		},
		Required: []string{"title", "kind", "parentAssetId", "parentCommandId"},
	})
}

func geminiCreateLocationCommandSchema() geminiSchema {
	return geminiActionPlanCommandBranch("create_location", "Create a new room or place. Use create_asset with kind container for boxes, cabinets, shelves, drawers, bins, counters, and surfaces.", geminiSchema{
		Type:        "object",
		Description: "Arguments for create_location. Use only for true rooms or places like Kitchen, Office, Garage, or Living room.",
		Properties: map[string]geminiSchema{
			"title": {
				Type:        "string",
				Description: "Title of the new location. Prefer this over name.",
			},
			"name": {
				Type:        "string",
				Description: "Compatibility alias for title when title is unavailable.",
			},
			"description": {
				Type:        "string",
				Description: "Optional user-provided description.",
			},
			"parentAssetId": {
				Type:        "string",
				Description: "Existing parent asset id copied exactly from a read-tool result, only for nested place-like locations.",
			},
			"parentCommandId": {
				Type:        "string",
				Description: "Earlier create command id when the parent is created in this same plan.",
			},
		},
		AnyOf: []geminiSchema{
			{Type: "object", Required: []string{"title"}},
			{Type: "object", Required: []string{"name"}},
		},
	})
}

func geminiMoveAssetCommandSchema() geminiSchema {
	return geminiActionPlanCommandBranch("move_asset", "Move an existing asset returned by a read tool. Never use this for an item created earlier in the same plan.", geminiSchema{
		Type:        "object",
		Description: "Arguments for move_asset. assetId must be an existing asset id copied exactly from a read-tool result. Set parentAssetId to an existing destination, parentCommandId to an earlier create command id, or parentAssetId null and parentCommandId empty string for root.",
		Properties: map[string]geminiSchema{
			"assetId": {
				Type:        "string",
				Description: "Existing source asset id copied exactly from a read-tool result.",
			},
			"parentAssetId": {
				Description: "Existing destination parent asset id copied from a read-tool result, or null to move to root.",
				AnyOf: []geminiSchema{
					{Type: "string"},
					{Type: "null"},
				},
			},
			"parentCommandId": {
				Type:        "string",
				Description: "Earlier create command id when moving into a destination created in this same plan, or empty string when parentAssetId is used.",
			},
		},
		Required: []string{"assetId", "parentAssetId", "parentCommandId"},
	})
}

func geminiAssetIDOnlyCommandSchema(kind string, description string) geminiSchema {
	return geminiActionPlanCommandBranch(kind, description, geminiSchema{
		Type:        "object",
		Description: "Arguments for " + kind + ".",
		Properties: map[string]geminiSchema{
			"assetId": {
				Type:        "string",
				Description: "Existing asset id copied exactly from a read-tool result.",
			},
		},
		Required: []string{"assetId"},
	})
}

func geminiActionPlanCommandBranch(kind string, description string, arguments geminiSchema) geminiSchema {
	return geminiSchema{
		Type:        "object",
		Description: description,
		Properties: map[string]geminiSchema{
			"id": {
				Type:        "string",
				Description: "Stable command id used by later commands as parentCommandId.",
			},
			"kind": {
				Type:        "string",
				Description: "Action-plan command kind.",
				Enum:        []string{kind},
			},
			"summary": {
				Type:        "string",
				Description: "Short safe user-facing summary of this command.",
			},
			"arguments": arguments,
		},
		Required: []string{"id", "kind", "summary", "arguments"},
	}
}
