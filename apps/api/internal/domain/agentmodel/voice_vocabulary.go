package agentmodel

import "strings"

const (
	MaxVoiceVocabularyAssetTypes   = 32
	MaxVoiceVocabularyCustomFields = 64
	MaxVoiceVocabularyTags         = 32
	MaxVoiceVocabularyRequests     = 12
	MaxVoiceVocabularyEnumOptions  = 40
)

type VoiceVocabularyKind string

const (
	VoiceVocabularyKindCustomAssetType VoiceVocabularyKind = "custom_asset_type"
	VoiceVocabularyKindCustomField     VoiceVocabularyKind = "custom_field"
	VoiceVocabularyKindTag             VoiceVocabularyKind = "tag"
)

func (kind VoiceVocabularyKind) Valid() bool {
	return kind == VoiceVocabularyKindCustomAssetType || kind == VoiceVocabularyKindCustomField || kind == VoiceVocabularyKindTag
}

type VoiceVocabularyAssetType struct {
	Key         string `json:"key"`
	DisplayName string `json:"displayName"`
	Description string `json:"description,omitempty"`
}

func (value VoiceVocabularyAssetType) Validate() error {
	if !validVoiceVocabularyKey(value.Key) || !bounded(value.DisplayName, 120, false) || !bounded(value.Description, 500, true) {
		return ErrInvalidVoiceInvestigation
	}
	return nil
}

type VoiceVocabularyFieldSummary struct {
	Key           string `json:"key"`
	DisplayName   string `json:"displayName"`
	FieldType     string `json:"fieldType"`
	Applicability string `json:"applicability"`
}

func (value VoiceVocabularyFieldSummary) Validate() error {
	if !validVoiceVocabularyKey(value.Key) || !bounded(value.DisplayName, 120, false) || !validVoiceVocabularyFieldType(value.FieldType) || !validVoiceVocabularyApplicability(value.Applicability) {
		return ErrInvalidVoiceInvestigation
	}
	return nil
}

type VoiceVocabularyTag struct {
	Key         string `json:"key"`
	DisplayName string `json:"displayName"`
}

func (value VoiceVocabularyTag) Validate() error {
	if !validVoiceVocabularyTagKey(value.Key) || !bounded(value.DisplayName, 80, false) {
		return ErrInvalidVoiceInvestigation
	}
	return nil
}

type VoiceVocabularyManifest struct {
	CustomAssetTypes          []VoiceVocabularyAssetType    `json:"customAssetTypes"`
	CustomFields              []VoiceVocabularyFieldSummary `json:"customFields"`
	Tags                      []VoiceVocabularyTag          `json:"tags"`
	CustomAssetTypesTruncated bool                          `json:"customAssetTypesTruncated"`
	CustomFieldsTruncated     bool                          `json:"customFieldsTruncated"`
	TagsTruncated             bool                          `json:"tagsTruncated"`
}

func (manifest VoiceVocabularyManifest) Validate() error {
	if len(manifest.CustomAssetTypes) > MaxVoiceVocabularyAssetTypes || len(manifest.CustomFields) > MaxVoiceVocabularyCustomFields || len(manifest.Tags) > MaxVoiceVocabularyTags ||
		!validUniqueVoiceVocabulary(manifest.CustomAssetTypes, func(value VoiceVocabularyAssetType) (string, error) { return value.Key, value.Validate() }) ||
		!validUniqueVoiceVocabulary(manifest.CustomFields, func(value VoiceVocabularyFieldSummary) (string, error) { return value.Key, value.Validate() }) ||
		!validUniqueVoiceVocabulary(manifest.Tags, func(value VoiceVocabularyTag) (string, error) { return value.Key, value.Validate() }) {
		return ErrInvalidVoiceInvestigation
	}
	return nil
}

type VoiceVocabularyRequest struct {
	Kind VoiceVocabularyKind `json:"kind"`
	Key  string              `json:"key"`
}

func (request VoiceVocabularyRequest) Validate() error {
	if !request.Kind.Valid() || (request.Kind == VoiceVocabularyKindTag && !validVoiceVocabularyTagKey(request.Key)) || (request.Kind != VoiceVocabularyKindTag && !validVoiceVocabularyKey(request.Key)) {
		return ErrInvalidVoiceInvestigation
	}
	return nil
}

type VoiceVocabularyDefinition struct {
	Kind                          VoiceVocabularyKind `json:"kind"`
	Key                           string              `json:"key"`
	DisplayName                   string              `json:"displayName"`
	Description                   string              `json:"description,omitempty"`
	FieldType                     string              `json:"fieldType,omitempty"`
	Applicability                 string              `json:"applicability,omitempty"`
	EnumOptions                   []string            `json:"enumOptions,omitempty"`
	EnumOptionsTruncated          bool                `json:"enumOptionsTruncated,omitempty"`
	ApplicableCustomAssetTypeKeys []string            `json:"applicableCustomAssetTypeKeys,omitempty"`
	ApplicabilityTargetsTruncated bool                `json:"applicabilityTargetsTruncated,omitempty"`
}

func (definition VoiceVocabularyDefinition) Validate() error {
	if !definition.Kind.Valid() || (definition.Kind == VoiceVocabularyKindTag && !validVoiceVocabularyTagKey(definition.Key)) || (definition.Kind != VoiceVocabularyKindTag && !validVoiceVocabularyKey(definition.Key)) || !bounded(definition.DisplayName, 120, false) || !bounded(definition.Description, 500, true) ||
		len(definition.EnumOptions) > MaxVoiceVocabularyEnumOptions || len(definition.ApplicableCustomAssetTypeKeys) > MaxVoiceVocabularyAssetTypes ||
		!validUniqueVocabularyKeys(definition.EnumOptions) || !validUniqueVocabularyKeys(definition.ApplicableCustomAssetTypeKeys) {
		return ErrInvalidVoiceInvestigation
	}
	switch definition.Kind {
	case VoiceVocabularyKindCustomAssetType, VoiceVocabularyKindTag:
		if definition.FieldType != "" || definition.Applicability != "" || len(definition.EnumOptions) != 0 || definition.EnumOptionsTruncated || len(definition.ApplicableCustomAssetTypeKeys) != 0 || definition.ApplicabilityTargetsTruncated {
			return ErrInvalidVoiceInvestigation
		}
	case VoiceVocabularyKindCustomField:
		if !validVoiceVocabularyFieldType(definition.FieldType) || !validVoiceVocabularyApplicability(definition.Applicability) ||
			(definition.FieldType == "enum" && len(definition.EnumOptions) == 0) || (definition.FieldType != "enum" && len(definition.EnumOptions) != 0) ||
			(definition.Applicability == "all_assets" && (len(definition.ApplicableCustomAssetTypeKeys) != 0 || definition.ApplicabilityTargetsTruncated)) ||
			(definition.Applicability == "custom_asset_types" && len(definition.ApplicableCustomAssetTypeKeys) == 0 && !definition.ApplicabilityTargetsTruncated) {
			return ErrInvalidVoiceInvestigation
		}
	}
	return nil
}

func validVoiceVocabularyFieldType(value string) bool {
	switch value {
	case "text", "number", "boolean", "date", "url", "enum":
		return true
	default:
		return false
	}
}

func validVoiceVocabularyApplicability(value string) bool {
	return value == "all_assets" || value == "custom_asset_types"
}

func validVoiceVocabularyKey(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) == 0 || len(value) > 80 || value[0] < 'a' || value[0] > 'z' || strings.HasSuffix(value, "-") {
		return false
	}
	for _, r := range value[1:] {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			continue
		}
		return false
	}
	return true
}

func validVoiceVocabularyTagKey(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) == 0 || len(value) > 80 || strings.HasSuffix(value, "-") {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			continue
		}
		return false
	}
	return true
}

func validUniqueVocabularyKeys(values []string) bool {
	seen := map[string]struct{}{}
	for _, value := range values {
		if !validVoiceVocabularyKey(value) {
			return false
		}
		if _, exists := seen[value]; exists {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}

func validUniqueVoiceVocabulary[T any](values []T, validate func(T) (string, error)) bool {
	seen := map[string]struct{}{}
	for _, value := range values {
		key, err := validate(value)
		if err != nil {
			return false
		}
		if _, exists := seen[key]; exists {
			return false
		}
		seen[key] = struct{}{}
	}
	return true
}
