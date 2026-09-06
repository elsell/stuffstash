package app

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func realtimeConversationVocabularyTool() ports.ConversationToolDefinition {
	return ports.ConversationToolDefinition{Name: RealtimeVoiceToolGetInventoryVocabulary, Description: "Discover the actual tags, item types and fields used to organize this inventory. Helpful when the user's concept differs from stored names or searches find nothing: the manifest lets you choose related local terms instead of guessing labels. No arguments returns a bounded manifest with truncation flags. Optional definitions requests use kind and stable key from the manifest to read field types, enum options and applicability. Unavailable keys are counted without discarding the manifest; they may be absent or outside its bounded coverage.", Parameters: json.RawMessage(`{"type":"object","properties":{"definitions":{"type":"array","maxItems":12,"items":{"type":"object","properties":{"kind":{"type":"string","enum":["custom_asset_type","custom_field","tag"]},"key":{"type":"string","maxLength":80}},"required":["kind","key"],"additionalProperties":false}}},"additionalProperties":false}`)}
}

func (a App) executeRealtimeVoiceVocabularyTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall) (ports.AgentToolResult, error) {
	var args struct {
		Definitions []agentmodel.VoiceVocabularyRequest `json:"definitions"`
	}
	raw, err := json.Marshal(call.Arguments)
	if err != nil {
		return ports.AgentToolResult{}, ports.ErrInvalidProviderInput
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&args) != nil || len(args.Definitions) > agentmodel.MaxVoiceVocabularyRequests {
		return ports.AgentToolResult{}, ports.ErrInvalidProviderInput
	}
	for _, definition := range args.Definitions {
		if definition.Validate() != nil {
			return ports.AgentToolResult{}, ports.ErrInvalidProviderInput
		}
	}
	types, err := a.ListInventoryCustomAssetTypes(ctx, ListCustomAssetTypesInput{Principal: session.Principal, TenantID: session.TenantID, InventoryID: session.InventoryID, Source: audit.SourceConversation, LifecycleState: "active", Limit: agentmodel.MaxVoiceVocabularyAssetTypes})
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	fields, err := a.ListInventoryCustomFieldDefinitions(ctx, ListCustomFieldDefinitionsInput{Principal: session.Principal, TenantID: session.TenantID, InventoryID: session.InventoryID, Source: audit.SourceConversation, LifecycleState: "active", Limit: agentmodel.MaxVoiceVocabularyCustomFields})
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	tags, err := a.ListAssetTags(ctx, ListAssetTagsInput{Principal: session.Principal, TenantID: session.TenantID, InventoryID: session.InventoryID, Source: audit.SourceConversation, Limit: agentmodel.MaxVoiceVocabularyTags})
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	manifest, catalog, err := projectRealtimeVoiceVocabulary(types.Items, fields.Items, tags.Items)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	manifest.CustomAssetTypesTruncated = manifest.CustomAssetTypesTruncated || types.HasMore
	manifest.CustomFieldsTruncated = manifest.CustomFieldsTruncated || fields.HasMore
	manifest.TagsTruncated = manifest.TagsTruncated || tags.HasMore
	definitions, unavailable, err := catalog.resolve(args.Definitions)
	if err != nil {
		return ports.AgentToolResult{}, ports.ErrInvalidProviderInput
	}
	if err := a.ensureRealtimeVoiceAccess(ctx, session.Principal, session.TenantID, session.InventoryID); err != nil {
		return ports.AgentToolResult{}, err
	}
	content, err := json.Marshal(struct {
		Manifest                   agentmodel.VoiceVocabularyManifest     `json:"manifest"`
		Definitions                []agentmodel.VoiceVocabularyDefinition `json:"definitions,omitempty"`
		UnavailableDefinitionCount int                                    `json:"unavailableDefinitionCount,omitempty"`
	}{Manifest: manifest, Definitions: definitions, UnavailableDefinitionCount: unavailable})
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	return ports.AgentToolResult{CallID: call.ID, Name: call.Name, Call: call, Content: string(content)}, nil
}
