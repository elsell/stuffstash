package agentmodel

import (
	"sort"
	"strings"

	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
)

type VoiceVocabularyCatalog struct {
	definitions map[string]domain.VoiceVocabularyDefinition
}

func (catalog VoiceVocabularyCatalog) Resolve(requests []domain.VoiceVocabularyRequest) ([]domain.VoiceVocabularyDefinition, int, error) {
	definitions := make([]domain.VoiceVocabularyDefinition, 0, len(requests))
	seen := map[string]struct{}{}
	unavailable := 0
	for _, request := range requests {
		if request.Validate() != nil {
			return nil, 0, domain.ErrInvalidVoiceVocabulary
		}
		key := voiceVocabularyCatalogKey(request.Kind, request.Key)
		if _, exists := seen[key]; exists {
			return nil, 0, domain.ErrInvalidVoiceVocabulary
		}
		seen[key] = struct{}{}
		definition, exists := catalog.definitions[key]
		if !exists {
			unavailable++
			continue
		}
		definitions = append(definitions, definition)
	}
	return definitions, unavailable, nil
}

func ProjectVoiceVocabulary(assetTypes []customfield.AssetType, fields []customfield.Definition, tags []assettag.Tag) (domain.VoiceVocabularyManifest, VoiceVocabularyCatalog, error) {
	manifest := domain.VoiceVocabularyManifest{}
	catalog := VoiceVocabularyCatalog{definitions: map[string]domain.VoiceVocabularyDefinition{}}

	manifest.CustomAssetTypesTruncated = len(assetTypes) > domain.MaxVoiceVocabularyAssetTypes
	if manifest.CustomAssetTypesTruncated {
		assetTypes = assetTypes[:domain.MaxVoiceVocabularyAssetTypes]
	}
	manifest.CustomFieldsTruncated = len(fields) > domain.MaxVoiceVocabularyCustomFields
	if manifest.CustomFieldsTruncated {
		fields = fields[:domain.MaxVoiceVocabularyCustomFields]
	}
	manifest.TagsTruncated = len(tags) > domain.MaxVoiceVocabularyTags
	if manifest.TagsTruncated {
		tags = tags[:domain.MaxVoiceVocabularyTags]
	}

	typeKeysByID := map[customfield.AssetTypeID]string{}
	for _, assetType := range assetTypes {
		key := assetType.Key.String()
		typeKeysByID[assetType.ID] = key
		manifest.CustomAssetTypes = append(manifest.CustomAssetTypes, domain.VoiceVocabularyAssetType{AssetTypeID: assetType.ID.String(), ExpirationEnabled: assetType.ExpirationEnabled, Key: key, DisplayName: assetType.DisplayName.String(), Description: assetType.Description.String()})
		catalog.definitions[voiceVocabularyCatalogKey(domain.VoiceVocabularyKindCustomAssetType, key)] = domain.VoiceVocabularyDefinition{
			AssetTypeID: assetType.ID.String(), ExpirationEnabled: assetType.ExpirationEnabled, Kind: domain.VoiceVocabularyKindCustomAssetType, Key: key, DisplayName: assetType.DisplayName.String(), Description: assetType.Description.String(),
		}
	}
	for _, field := range fields {
		key := field.Key.String()
		manifest.CustomFields = append(manifest.CustomFields, domain.VoiceVocabularyFieldSummary{Key: key, DisplayName: field.DisplayName.String(), FieldType: field.Type.String(), Applicability: field.Applicability.String()})
		definition := domain.VoiceVocabularyDefinition{Kind: domain.VoiceVocabularyKindCustomField, Key: key, DisplayName: field.DisplayName.String(), FieldType: field.Type.String(), Applicability: field.Applicability.String()}
		for index, option := range field.EnumOptions {
			if index == domain.MaxVoiceVocabularyEnumOptions {
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
		catalog.definitions[voiceVocabularyCatalogKey(domain.VoiceVocabularyKindCustomField, key)] = definition
	}
	for _, tag := range tags {
		key := tag.Key.String()
		manifest.Tags = append(manifest.Tags, domain.VoiceVocabularyTag{Key: key, DisplayName: tag.DisplayName.String()})
		catalog.definitions[voiceVocabularyCatalogKey(domain.VoiceVocabularyKindTag, key)] = domain.VoiceVocabularyDefinition{Kind: domain.VoiceVocabularyKindTag, Key: key, DisplayName: tag.DisplayName.String()}
	}
	if manifest.Validate() != nil {
		return domain.VoiceVocabularyManifest{}, VoiceVocabularyCatalog{}, domain.ErrInvalidVoiceVocabulary
	}
	return manifest, catalog, nil
}

func voiceVocabularyCatalogKey(kind domain.VoiceVocabularyKind, key string) string {
	return string(kind) + "\x00" + strings.TrimSpace(key)
}
