package agentmodel

import "testing"

func TestVoiceVocabularyManifestAndTargetedDefinitionsValidateWithoutInternalIDs(t *testing.T) {
	t.Parallel()
	manifest := VoiceVocabularyManifest{
		CustomAssetTypes: []VoiceVocabularyAssetType{{Key: "medicine", DisplayName: "Medicine", Description: "Medication and supplements"}},
		CustomFields:     []VoiceVocabularyFieldSummary{{Key: "expiration-date", DisplayName: "Expiration Date", FieldType: "date", Applicability: "custom_asset_types"}},
		Tags:             []VoiceVocabularyTag{{Key: "camping", DisplayName: "Camping"}}, TagsTruncated: true,
	}
	if err := manifest.Validate(); err != nil {
		t.Fatalf("expected valid vocabulary manifest: %v", err)
	}
	request := VoiceVocabularyRequest{Kind: VoiceVocabularyKindCustomField, Key: "expiration-date"}
	definition := VoiceVocabularyDefinition{Kind: VoiceVocabularyKindCustomField, Key: "expiration-date", DisplayName: "Expiration Date", FieldType: "date", Applicability: "custom_asset_types", ApplicableCustomAssetTypeKeys: []string{"medicine"}}
	if request.Validate() != nil || definition.Validate() != nil {
		t.Fatalf("expected stable-key request and resolved definition: %+v %+v", request, definition)
	}
	invalid := []VoiceVocabularyManifest{
		{CustomAssetTypes: []VoiceVocabularyAssetType{{Key: "medicine", DisplayName: "Medicine"}, {Key: "medicine", DisplayName: "Duplicate"}}},
		{CustomFields: []VoiceVocabularyFieldSummary{{Key: "expires", DisplayName: "Expires", FieldType: "provider-type", Applicability: "all_assets"}}},
		{Tags: make([]VoiceVocabularyTag, MaxVoiceVocabularyTags+1)},
	}
	for _, value := range invalid {
		if value.Validate() == nil {
			t.Fatalf("expected invalid manifest: %+v", value)
		}
	}
}
