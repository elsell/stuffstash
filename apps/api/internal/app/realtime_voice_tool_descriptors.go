package app

import "github.com/stuffstash/stuff-stash/internal/ports"

func realtimeVoiceToolDescriptors() []ports.AgentToolDescriptor {
	return []ports.AgentToolDescriptor{
		{
			Name:        RealtimeVoiceToolSearchAuthorizedAssets,
			Label:       realtimeVoiceSearchAuthorizedAssetsPublicName,
			Description: "Search visible assets in the selected inventory by natural-language keywords. Use this for where-is, do-I-have, specific-item questions, and resolving resources for action plans. Arguments: query string, optional limit number. Results are JSON with asset metadata, opaque internal asset IDs for follow-up tool calls or action-plan arguments, and containment paths. Do not speak or display asset IDs to the user.",
			ReadOnly:    true,
			Parameters: ports.AgentToolParameters{
				Required: []string{"query"},
				Properties: map[string]ports.AgentToolParameter{
					"query": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Short natural-language keywords for the visible asset, container, or location the user asked about.",
					},
					"limit": {
						Type:        ports.AgentToolParameterTypeInteger,
						Description: "Maximum number of visible matching assets to return. Defaults to 10 and is capped at 20.",
					},
				},
			},
		},
		{
			Name:        RealtimeVoiceToolListAuthorizedAssets,
			Label:       realtimeVoiceListAuthorizedAssetsPublicName,
			Description: "List visible assets in the selected inventory. Use this for broad inventory questions like what items do I have, what is in a place, or what archived item should be restored. Arguments: optional kind item|container|location, optional lifecycleState active|archived|all, optional parentTitle string, optional locationTitle string, optional limit number. Results are JSON with asset metadata, internal asset IDs for action-plan arguments, and containment paths.",
			ReadOnly:    true,
			Parameters: ports.AgentToolParameters{
				Properties: map[string]ports.AgentToolParameter{
					"kind": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Optional asset kind filter.",
						Enum:        []string{"item", "container", "location"},
					},
					"lifecycleState": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Optional lifecycle filter. Defaults to active. Use archived when the user asks to restore an archived asset.",
						Enum:        []string{"active", "archived", "all"},
					},
					"parentTitle": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Optional direct parent title filter for questions about what is inside a specific container or location.",
					},
					"locationTitle": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Optional containing location title filter for questions about what is in a place.",
					},
					"limit": {
						Type:        ports.AgentToolParameterTypeInteger,
						Description: "Maximum number of visible assets to return. Defaults to 10 and is capped at 20.",
					},
				},
			},
		},
		{
			Name:        RealtimeVoiceToolProposeActionPlan,
			Label:       realtimeVoiceProposeActionPlanPublicName,
			Description: "Prepare a user-reviewable action plan for a requested inventory change. This does not execute the change. Use commandKind plus arguments for single-command create, move, update, archive, or restore plans. Prefer commands array for multi-step plans. Each command object has id, kind, summary, and arguments. For create or move into an existing location/container, resolve the parent with read tools and use its assetId as parentAssetId. For create or move into something created earlier in the same plan, use parentCommandId. If the user asks to move an existing asset to a missing but clearly named location or container, assume they want it created unless the words are ambiguous or likely mistranscribed. Propose creating the missing destination first and then moving the existing asset using parentCommandId. Do not ask a final yes/no clarification for clear missing destinations; this proposal is the review step.",
			ReadOnly:    true,
			Parameters: ports.AgentToolParameters{
				Required: []string{"intentSummary", "modelInterpretationSummary", "confirmationSummary"},
				Properties: map[string]ports.AgentToolParameter{
					"commandKind": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Action-plan command kind.",
						Enum:        []string{"create_asset", "create_location", "move_asset", "update_asset", "archive_asset", "restore_asset"},
					},
					"intentSummary": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Safe user-facing summary of what the user asked to change.",
					},
					"modelInterpretationSummary": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Safe user-facing summary of how the request was interpreted.",
					},
					"confirmationSummary": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Short confirmation question shown to the user.",
					},
					"commandSummary": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Compatibility field for a short safe summary of a single proposed command. Prefer commands[].summary.",
					},
					"arguments": {
						Type:        ports.AgentToolParameterTypeObject,
						Description: "Compatibility field for one safe structured command argument object. Prefer commands[].arguments.",
					},
					"commands": {
						Type:        ports.AgentToolParameterTypeArray,
						Description: "Preferred command array for multi-step plans. Each item must include id, kind, summary, and arguments. For create_asset use title/name, optional kind item|container|location, optional description, optional parentAssetId, or optional parentCommandId. For move_asset use assetId plus either parentAssetId, parentCommandId, or null parentAssetId for root. parentCommandId must point to an earlier create command id. Never use parentTitle or locationTitle here.",
						Items: &ports.AgentToolParameter{
							Type:     ports.AgentToolParameterTypeObject,
							Required: []string{"id", "kind", "summary", "arguments"},
							Properties: map[string]ports.AgentToolParameter{
								"id": {
									Type:        ports.AgentToolParameterTypeString,
									Description: "Stable command id used by later commands as parentCommandId. Use short ASCII letters, numbers, dashes, underscores, or dots.",
								},
								"kind": {
									Type:        ports.AgentToolParameterTypeString,
									Description: "Action-plan command kind for multi-step plans.",
									Enum:        []string{"create_asset", "create_location", "move_asset"},
								},
								"summary": {
									Type:        ports.AgentToolParameterTypeString,
									Description: "Short safe user-facing summary of this command.",
								},
								"arguments": {
									Type:        ports.AgentToolParameterTypeObject,
									Description: "Structured command arguments. For creates include title or name, kind item|container|location, optional description, and exactly one parentAssetId or parentCommandId when placing the new asset. For move_asset include assetId and either parentAssetId, parentCommandId, or parentAssetId null for root.",
									Properties: map[string]ports.AgentToolParameter{
										"assetId": {
											Type:        ports.AgentToolParameterTypeString,
											Description: "Existing asset id returned by a read tool for move_asset.",
										},
										"title": {
											Type:        ports.AgentToolParameterTypeString,
											Description: "New asset title.",
										},
										"name": {
											Type:        ports.AgentToolParameterTypeString,
											Description: "Compatibility alias for title.",
										},
										"kind": {
											Type:        ports.AgentToolParameterTypeString,
											Description: "New asset kind for create_asset.",
											Enum:        []string{"item", "container", "location"},
										},
										"description": {
											Type:        ports.AgentToolParameterTypeString,
											Description: "Optional user-provided description.",
										},
										"parentAssetId": {
											Type:        ports.AgentToolParameterTypeString,
											Description: "Existing parent asset id returned by a read tool.",
										},
										"parentCommandId": {
											Type:        ports.AgentToolParameterTypeString,
											Description: "Earlier create command id when placing this asset inside something created in the same plan.",
										},
									},
								},
							},
						},
					},
					"riskSummary": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Optional short safe risk summary.",
					},
				},
			},
		},
	}
}

func realtimeVoiceToolLabel(name string) string {
	switch name {
	case RealtimeVoiceToolProposeActionPlan:
		return realtimeVoiceProposeActionPlanPublicName
	case RealtimeVoiceToolListAuthorizedAssets:
		return realtimeVoiceListAuthorizedAssetsPublicName
	default:
		return realtimeVoiceSearchAuthorizedAssetsPublicName
	}
}
