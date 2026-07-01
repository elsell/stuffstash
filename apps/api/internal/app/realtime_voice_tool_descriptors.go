package app

import "github.com/stuffstash/stuff-stash/internal/ports"

func realtimeVoiceToolDescriptors() []ports.AgentToolDescriptor {
	return []ports.AgentToolDescriptor{
		{
			Name:             RealtimeVoiceToolSearchAuthorizedAssets,
			Label:            realtimeVoiceSearchAuthorizedAssetsPublicName,
			Description:      "Search visible assets in the selected inventory by natural-language keywords. Use this for where-is, do-I-have, specific-item questions, and resolving resources for action plans. Arguments: query string, optional limit number. Results are JSON with asset metadata, opaque internal asset IDs for follow-up tool calls or action-plan arguments, and containment paths. Do not speak or display asset IDs to the user.",
			ReadOnly:         true,
			ProviderCallable: true,
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
			Name:             RealtimeVoiceToolListAuthorizedAssets,
			Label:            realtimeVoiceListAuthorizedAssetsPublicName,
			Description:      "List visible assets in the selected inventory. Use this for broad inventory questions like what items do I have, what is in a place, or what archived item should be restored. Arguments: optional kind item|container|location, optional lifecycleState active|archived|all, optional parentTitle string, optional locationTitle string, optional limit number. Results are JSON with asset metadata, internal asset IDs for action-plan arguments, and containment paths.",
			ReadOnly:         true,
			ProviderCallable: true,
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
			Name:             RealtimeVoiceToolListAssetAuditHistory,
			Label:            realtimeVoiceListAssetAuditHistoryPublicName,
			Description:      "List safe audit history for one visible asset already returned by an authorized read tool in this voice session. Use this for history, movement, when-created, when-moved, when-updated, archive/restore, and who-changed questions. Arguments: assetId from a prior read tool result, optional limit number. Results are JSON with safe audit entries, timestamps, action names, source, asset kind, current title, movement parent titles when available, lifecycle state changes when available, and concise summaries. Do not guess history from current location alone. Do not speak or display asset IDs to the user.",
			ReadOnly:         true,
			ProviderCallable: true,
			Parameters: ports.AgentToolParameters{
				Required: []string{"assetId"},
				Properties: map[string]ports.AgentToolParameter{
					"assetId": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Opaque asset ID copied exactly from an earlier authorized read tool result in this session.",
					},
					"limit": {
						Type:        ports.AgentToolParameterTypeInteger,
						Description: "Maximum number of safe audit entries to return. Defaults to 10 and is capped at 20.",
					},
				},
			},
		},
		{
			Name:             RealtimeVoiceToolProposeActionPlan,
			Label:            realtimeVoiceProposeActionPlanPublicName,
			Description:      "Prepare a user-reviewable action plan for a requested inventory change. This does not execute the change. Use commandKind plus arguments for single-command create, move, update, archive, or restore plans. Prefer commands array for multi-step plans. Each command object has id, kind, summary, and arguments. Use create_asset for new items and containers. Use create_location only for a true location. Never put assetId in create_asset arguments; assetId is only for move_asset, archive_asset, or restore_asset of an existing asset returned by a read tool. For create or move into an existing location/container, resolve the parent with read tools and use its assetId as parentAssetId. For create or move into something created earlier in the same plan, use parentCommandId. If the user asks to add a new item into a missing container under an existing location, create the missing container first with parentAssetId set to the existing location, then create the item with parentCommandId set to the container command id. If the user asks to move an existing asset to a missing but clearly named location or container, assume they want it created unless the words are ambiguous or likely mistranscribed. Propose creating the missing destination first and then moving the existing asset using parentCommandId. Do not ask a final yes/no clarification for clear missing destinations; this proposal is the review step.",
			ProviderCallable: true,
			RequiresApproval: true,
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
						Description: "Preferred command array for multi-step plans. Each item must include id, kind, summary, and arguments. For create_asset use title/name, kind item|container|location, optional description, and optional parentAssetId or parentCommandId. Do not include assetId in create_asset. For create_location use title/name and no kind, or kind location. For move_asset use assetId plus either parentAssetId, parentCommandId, or null parentAssetId for root. For archive_asset or restore_asset use assetId from a read tool. parentCommandId must point to an earlier create command id. Never use parentTitle or locationTitle here.",
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
									Enum:        []string{"create_asset", "create_location", "move_asset", "update_asset", "archive_asset", "restore_asset"},
								},
								"summary": {
									Type:        ports.AgentToolParameterTypeString,
									Description: "Short safe user-facing summary of this command.",
								},
								"arguments": {
									Type:        ports.AgentToolParameterTypeObject,
									Description: "Structured command arguments. For create_asset include title or name, kind item|container|location, optional description, and exactly one parentAssetId or parentCommandId when placing the new asset. Never include assetId in create_asset. For create_location include title or name and optional parent references only when creating a location-like child; do not use kind container with create_location. For move_asset include assetId and either parentAssetId, parentCommandId, or parentAssetId null for root. For archive_asset or restore_asset include assetId.",
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

func realtimeVoiceReadToolDescriptors() []ports.AgentToolDescriptor {
	tools := realtimeVoiceToolDescriptors()
	readTools := make([]ports.AgentToolDescriptor, 0, len(tools))
	for _, tool := range tools {
		if tool.Name == RealtimeVoiceToolProposeActionPlan {
			continue
		}
		readTools = append(readTools, tool)
	}
	return readTools
}

func realtimeVoiceInitialReadToolDescriptors() []ports.AgentToolDescriptor {
	return []ports.AgentToolDescriptor{
		realtimeVoiceSearchAuthorizedAssetsToolDescriptor(),
		realtimeVoiceListAuthorizedAssetsToolDescriptor(),
	}
}

func realtimeVoiceSearchAuthorizedAssetsToolDescriptor() ports.AgentToolDescriptor {
	for _, tool := range realtimeVoiceToolDescriptors() {
		if tool.Name == RealtimeVoiceToolSearchAuthorizedAssets {
			return tool
		}
	}
	return ports.AgentToolDescriptor{}
}

func realtimeVoiceListAuthorizedAssetsToolDescriptor() ports.AgentToolDescriptor {
	for _, tool := range realtimeVoiceToolDescriptors() {
		if tool.Name == RealtimeVoiceToolListAuthorizedAssets {
			return tool
		}
	}
	return ports.AgentToolDescriptor{}
}

func realtimeVoiceToolLabel(name string) string {
	switch name {
	case RealtimeVoiceToolProposeActionPlan:
		return realtimeVoiceProposeActionPlanPublicName
	case RealtimeVoiceToolListAuthorizedAssets:
		return realtimeVoiceListAuthorizedAssetsPublicName
	case RealtimeVoiceToolListAssetAuditHistory:
		return realtimeVoiceListAssetAuditHistoryPublicName
	default:
		return realtimeVoiceSearchAuthorizedAssetsPublicName
	}
}
