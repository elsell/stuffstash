package app

import (
	"context"
	"sort"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type realtimeVoiceVocabularyCatalog struct {
	definitions map[string]agentmodel.VoiceVocabularyDefinition
}

func (catalog realtimeVoiceVocabularyCatalog) resolve(requests []agentmodel.VoiceVocabularyRequest) ([]agentmodel.VoiceVocabularyDefinition, error) {
	definitions := make([]agentmodel.VoiceVocabularyDefinition, 0, len(requests))
	seen := map[string]struct{}{}
	for _, request := range requests {
		if request.Validate() != nil {
			return nil, agentmodel.ErrInvalidVoiceInvestigation
		}
		key := realtimeVoiceVocabularyCatalogKey(request.Kind, request.Key)
		if _, exists := seen[key]; exists {
			return nil, agentmodel.ErrInvalidVoiceInvestigation
		}
		definition, exists := catalog.definitions[key]
		if !exists {
			return nil, agentmodel.ErrInvalidVoiceInvestigation
		}
		seen[key] = struct{}{}
		definitions = append(definitions, definition)
	}
	return definitions, nil
}

func (a App) loadRealtimeVoiceVocabulary(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (agentmodel.VoiceVocabularyManifest, realtimeVoiceVocabularyCatalog, error) {
	manifest := agentmodel.VoiceVocabularyManifest{}
	catalog := realtimeVoiceVocabularyCatalog{definitions: map[string]agentmodel.VoiceVocabularyDefinition{}}

	assetTypes, err := a.loadRealtimeVoiceVocabularyAssetTypes(ctx, tenantID, inventoryID)
	if err != nil {
		return manifest, catalog, err
	}
	fields, err := a.loadRealtimeVoiceVocabularyFields(ctx, tenantID, inventoryID)
	if err != nil {
		return manifest, catalog, err
	}
	tags, err := a.loadRealtimeVoiceVocabularyTags(ctx, tenantID, inventoryID)
	if err != nil {
		return manifest, catalog, err
	}

	manifest.CustomAssetTypesTruncated = len(assetTypes) > agentmodel.MaxVoiceVocabularyAssetTypes
	if manifest.CustomAssetTypesTruncated {
		assetTypes = assetTypes[:agentmodel.MaxVoiceVocabularyAssetTypes]
	}
	manifest.CustomFieldsTruncated = len(fields) > agentmodel.MaxVoiceVocabularyCustomFields
	if manifest.CustomFieldsTruncated {
		fields = fields[:agentmodel.MaxVoiceVocabularyCustomFields]
	}
	manifest.TagsTruncated = len(tags) > agentmodel.MaxVoiceVocabularyTags
	if manifest.TagsTruncated {
		tags = tags[:agentmodel.MaxVoiceVocabularyTags]
	}

	typeKeysByID := map[customfield.AssetTypeID]string{}
	for _, assetType := range assetTypes {
		key := assetType.Key.String()
		typeKeysByID[assetType.ID] = key
		manifest.CustomAssetTypes = append(manifest.CustomAssetTypes, agentmodel.VoiceVocabularyAssetType{Key: key, DisplayName: assetType.DisplayName.String(), Description: assetType.Description.String()})
		catalog.definitions[realtimeVoiceVocabularyCatalogKey(agentmodel.VoiceVocabularyKindCustomAssetType, key)] = agentmodel.VoiceVocabularyDefinition{
			Kind: agentmodel.VoiceVocabularyKindCustomAssetType, Key: key, DisplayName: assetType.DisplayName.String(), Description: assetType.Description.String(),
		}
	}
	for _, field := range fields {
		key := field.Key.String()
		manifest.CustomFields = append(manifest.CustomFields, agentmodel.VoiceVocabularyFieldSummary{Key: key, DisplayName: field.DisplayName.String(), FieldType: field.Type.String(), Applicability: field.Applicability.String()})
		definition := agentmodel.VoiceVocabularyDefinition{Kind: agentmodel.VoiceVocabularyKindCustomField, Key: key, DisplayName: field.DisplayName.String(), FieldType: field.Type.String(), Applicability: field.Applicability.String()}
		for index, option := range field.EnumOptions {
			if index == agentmodel.MaxVoiceVocabularyEnumOptions {
				definition.EnumOptionsTruncated = true
				break
			}
			definition.EnumOptions = append(definition.EnumOptions, option.String())
		}
		for _, targetID := range field.CustomAssetTypeIDs {
			if targetKey, exists := typeKeysByID[targetID]; exists {
				definition.ApplicableCustomAssetTypeKeys = append(definition.ApplicableCustomAssetTypeKeys, targetKey)
			} else {
				definition.ApplicabilityTargetsTruncated = true
			}
		}
		sort.Strings(definition.ApplicableCustomAssetTypeKeys)
		catalog.definitions[realtimeVoiceVocabularyCatalogKey(agentmodel.VoiceVocabularyKindCustomField, key)] = definition
	}
	for _, tag := range tags {
		key := tag.Key.String()
		manifest.Tags = append(manifest.Tags, agentmodel.VoiceVocabularyTag{Key: key, DisplayName: tag.DisplayName.String()})
		catalog.definitions[realtimeVoiceVocabularyCatalogKey(agentmodel.VoiceVocabularyKindTag, key)] = agentmodel.VoiceVocabularyDefinition{Kind: agentmodel.VoiceVocabularyKindTag, Key: key, DisplayName: tag.DisplayName.String()}
	}
	if manifest.Validate() != nil {
		return agentmodel.VoiceVocabularyManifest{}, realtimeVoiceVocabularyCatalog{}, agentmodel.ErrInvalidVoiceInvestigation
	}
	return manifest, catalog, nil
}

func (a App) loadRealtimeVoiceVocabularyAssetTypes(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) ([]customfield.AssetType, error) {
	if a.customAssetTypes == nil {
		return nil, nil
	}
	values, err := a.customAssetTypes.ListInventoryCustomAssetTypes(ctx, tenantID, inventoryID, ports.CustomAssetTypePageRequest{Limit: agentmodel.MaxVoiceVocabularyAssetTypes + 1, Lifecycle: ports.CustomizationLifecycleActive})
	if err != nil {
		return nil, err
	}
	active := values[:0]
	for _, value := range values {
		if value.IsActive() {
			active = append(active, value)
		}
	}
	sort.Slice(active, func(i, j int) bool { return active[i].Key.String() < active[j].Key.String() })
	return active, nil
}

func (a App) loadRealtimeVoiceVocabularyFields(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) ([]customfield.Definition, error) {
	if a.customFields == nil {
		return nil, nil
	}
	values, err := a.customFields.ListInventoryCustomFieldDefinitions(ctx, tenantID, inventoryID, ports.CustomFieldDefinitionPageRequest{Limit: agentmodel.MaxVoiceVocabularyCustomFields + 1, Lifecycle: ports.CustomizationLifecycleActive})
	if err != nil {
		return nil, err
	}
	active := values[:0]
	for _, value := range values {
		if value.IsActive() {
			active = append(active, value)
		}
	}
	sort.Slice(active, func(i, j int) bool { return active[i].Key.String() < active[j].Key.String() })
	return active, nil
}

func (a App) loadRealtimeVoiceVocabularyTags(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) ([]assettag.Tag, error) {
	if a.assetTags == nil {
		return nil, nil
	}
	values, err := a.assetTags.ListAssetTags(ctx, tenantID, inventoryID, ports.AssetTagPageRequest{Limit: agentmodel.MaxVoiceVocabularyTags + 1})
	if err != nil {
		return nil, err
	}
	active := values[:0]
	for _, value := range values {
		if value.LifecycleState == assettag.LifecycleStateActive {
			active = append(active, value)
		}
	}
	sort.Slice(active, func(i, j int) bool {
		return strings.ToLower(active[i].DisplayName.String()) < strings.ToLower(active[j].DisplayName.String())
	})
	return active, nil
}

func realtimeVoiceVocabularyCatalogKey(kind agentmodel.VoiceVocabularyKind, key string) string {
	return string(kind) + "\x00" + strings.TrimSpace(key)
}

func mergeRealtimeVoiceVocabularyResolution(catalog realtimeVoiceVocabularyCatalog, requests []agentmodel.VoiceVocabularyRequest, definitions []agentmodel.VoiceVocabularyDefinition, additions []agentmodel.VoiceVocabularyRequest) ([]agentmodel.VoiceVocabularyRequest, []agentmodel.VoiceVocabularyDefinition, error) {
	seen := map[string]struct{}{}
	for _, request := range requests {
		seen[realtimeVoiceVocabularyCatalogKey(request.Kind, request.Key)] = struct{}{}
	}
	for _, request := range additions {
		if _, exists := seen[realtimeVoiceVocabularyCatalogKey(request.Kind, request.Key)]; exists {
			return nil, nil, agentmodel.ErrInvalidVoiceInvestigation
		}
	}
	resolved, err := catalog.resolve(additions)
	if err != nil || len(requests)+len(additions) > agentmodel.MaxVoiceVocabularyRequests {
		return nil, nil, agentmodel.ErrInvalidVoiceInvestigation
	}
	return append(requests, additions...), append(definitions, resolved...), nil
}
