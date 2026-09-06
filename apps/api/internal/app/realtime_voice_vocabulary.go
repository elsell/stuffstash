package app

import (
	"sort"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
)

type realtimeVoiceVocabularyCatalog struct {
	definitions map[string]agentmodel.VoiceVocabularyDefinition
}

func (catalog realtimeVoiceVocabularyCatalog) resolve(requests []agentmodel.VoiceVocabularyRequest) ([]agentmodel.VoiceVocabularyDefinition, error) {
	definitions := make([]agentmodel.VoiceVocabularyDefinition, 0, len(requests))
	seen := map[string]struct{}{}
	for _, request := range requests {
		if request.Validate() != nil {
			return nil, agentmodel.ErrInvalidVoiceVocabulary
		}
		key := realtimeVoiceVocabularyCatalogKey(request.Kind, request.Key)
		if _, exists := seen[key]; exists {
			return nil, agentmodel.ErrInvalidVoiceVocabulary
		}
		definition, exists := catalog.definitions[key]
		if !exists {
			return nil, agentmodel.ErrInvalidVoiceVocabulary
		}
		seen[key] = struct{}{}
		definitions = append(definitions, definition)
	}
	return definitions, nil
}

func projectRealtimeVoiceVocabulary(assetTypes []customfield.AssetType, fields []customfield.Definition, tags []assettag.Tag) (agentmodel.VoiceVocabularyManifest, realtimeVoiceVocabularyCatalog, error) {
	manifest := agentmodel.VoiceVocabularyManifest{}
	catalog := realtimeVoiceVocabularyCatalog{definitions: map[string]agentmodel.VoiceVocabularyDefinition{}}

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
		return agentmodel.VoiceVocabularyManifest{}, realtimeVoiceVocabularyCatalog{}, agentmodel.ErrInvalidVoiceVocabulary
	}
	return manifest, catalog, nil
}

func realtimeVoiceVocabularyCatalogKey(kind agentmodel.VoiceVocabularyKind, key string) string {
	return string(kind) + "\x00" + strings.TrimSpace(key)
}
