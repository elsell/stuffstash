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

func TestInvestigationStepBoundsAndDeduplicatesVocabularyRequests(t *testing.T) {
	t.Parallel()
	intent := Intent{Kind: IntentKindRead, Operation: OperationLocate, SubjectMention: "camp medicine"}
	read := SearchRequest{ReferenceKey: SemanticReferenceSubject, ReadKind: InvestigationReadSearchAssets, Mention: "camp medicine", SearchProbes: []string{"camp medicine"}}
	step := InvestigationStep{Decision: InvestigationDecisionSearch, Intent: intent, SearchRequests: []SearchRequest{read}, VocabularyRequests: []VoiceVocabularyRequest{{Kind: VoiceVocabularyKindCustomAssetType, Key: "medicine"}, {Kind: VoiceVocabularyKindCustomField, Key: "expiration-date"}}}
	if err := step.Validate(); err != nil {
		t.Fatalf("expected valid requests: %v", err)
	}
	step.VocabularyRequests = []VoiceVocabularyRequest{{Kind: VoiceVocabularyKindTag, Key: "camping"}, {Kind: VoiceVocabularyKindTag, Key: "camping"}}
	if step.Validate() == nil {
		t.Fatal("expected duplicate request to fail")
	}
	step = InvestigationStep{Decision: InvestigationDecisionFinish, Intent: intent, Resolutions: []Resolution{{ReferenceKey: SemanticReferenceSubject, Status: ResolutionAbsent}}, VocabularyRequests: []VoiceVocabularyRequest{{Kind: VoiceVocabularyKindTag, Key: "camping"}}}
	if step.Validate() == nil {
		t.Fatal("expected finish decision with vocabulary requests to fail")
	}
}

func TestSearchRequestLifecycleScopeDefaultsActiveAndRejectsUnknown(t *testing.T) {
	t.Parallel()
	request := SearchRequest{ReferenceKey: SemanticReferenceSubject, ReadKind: InvestigationReadSearchAssets, Mention: "archived drill", SearchProbes: []string{"archived drill"}}
	if request.Validate() != nil || request.LifecycleScope.Effective() != LifecycleScopeActive {
		t.Fatalf("expected omitted lifecycle to default active: %+v", request)
	}
	request.LifecycleScope = LifecycleScopeArchived
	if request.Validate() != nil || request.LifecycleScope.Effective() != LifecycleScopeArchived {
		t.Fatalf("expected archived lifecycle scope: %+v", request)
	}
	request.LifecycleScope = "deleted"
	if request.Validate() == nil {
		t.Fatal("expected unknown lifecycle scope to fail")
	}
}
