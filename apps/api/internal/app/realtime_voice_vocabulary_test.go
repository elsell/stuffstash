package app

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceVocabularyLoadsActiveScopedManifestAndResolvesRequestedDefinition(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := memory.NewStore()
	application := App{customAssetTypes: store, customFields: store, assetTags: store}
	tenantName, _ := tenant.NewName("Home")
	if err := store.SaveTenant(ctx, tenant.Tenant{ID: "tenant-home", Name: tenantName}); err != nil {
		t.Fatalf("save tenant: %v", err)
	}
	inventoryName, _ := inventory.NewName("Home")
	if err := store.SaveInventory(ctx, inventory.Inventory{ID: "inventory-home", TenantID: "tenant-home", Name: inventoryName}); err != nil {
		t.Fatalf("save inventory: %v", err)
	}
	if err := store.SaveInventory(ctx, inventory.Inventory{ID: "inventory-private", TenantID: "tenant-home", Name: inventoryName}); err != nil {
		t.Fatalf("save sibling inventory: %v", err)
	}
	otherTenantName, _ := tenant.NewName("Other")
	if err := store.SaveTenant(ctx, tenant.Tenant{ID: "tenant-other", Name: otherTenantName}); err != nil {
		t.Fatalf("save other tenant: %v", err)
	}
	if err := store.SaveInventory(ctx, inventory.Inventory{ID: "inventory-other", TenantID: "tenant-other", Name: inventoryName}); err != nil {
		t.Fatalf("save other inventory: %v", err)
	}

	medicine, ok := customfield.NewAssetType("type-medicine", "tenant-home", "inventory-home", customfield.ScopeInventory, "medicine", "Medicine", "Medication and supplements")
	if !ok {
		t.Fatal("create custom asset type fixture")
	}
	if err := store.SaveCustomAssetType(ctx, medicine, audit.Record{ID: "audit-type"}); err != nil {
		t.Fatalf("save custom asset type: %v", err)
	}
	expires, ok := customfield.NewDefinition("field-expires", "tenant-home", "inventory-home", customfield.ScopeInventory, "expiration-date", "Expiration Date", customfield.FieldTypeDate, nil, customfield.ApplicabilityCustomAssetTypes, []customfield.AssetTypeID{medicine.ID})
	if !ok {
		t.Fatal("create field definition fixture")
	}
	if err := store.SaveCustomFieldDefinition(ctx, expires, audit.Record{ID: "audit-field"}); err != nil {
		t.Fatalf("save custom field: %v", err)
	}
	tag, ok := assettag.NewTag("tag-camping", "tenant-home", "inventory-home", "camping", "Camping", "", time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC))
	if !ok {
		t.Fatal("create tag fixture")
	}
	if err := store.CreateAssetTag(ctx, tag, audit.Record{ID: "audit-tag"}); err != nil {
		t.Fatalf("save tag: %v", err)
	}
	seedHiddenVoiceVocabulary(t, ctx, store, "tenant-home", "inventory-private", "private")
	seedHiddenVoiceVocabulary(t, ctx, store, "tenant-other", "inventory-other", "other")

	manifest, catalog, err := application.loadRealtimeVoiceVocabulary(ctx, tenant.ID("tenant-home"), inventory.InventoryID("inventory-home"))
	if err != nil {
		t.Fatalf("load vocabulary: %v", err)
	}
	if len(manifest.CustomAssetTypes) != 1 || manifest.CustomAssetTypes[0].Key != "medicine" || len(manifest.CustomFields) != 1 || len(manifest.Tags) != 1 {
		t.Fatalf("unexpected manifest: %+v", manifest)
	}
	definitions, err := catalog.resolve([]agentmodel.VoiceVocabularyRequest{{Kind: agentmodel.VoiceVocabularyKindCustomField, Key: "expiration-date"}})
	if err != nil || len(definitions) != 1 {
		t.Fatalf("resolve field definition: %+v, %v", definitions, err)
	}
	if got := definitions[0].ApplicableCustomAssetTypeKeys; len(got) != 1 || got[0] != "medicine" {
		t.Fatalf("expected key-based applicability without IDs, got %+v", definitions[0])
	}
	if definitions[0].Key == expires.ID.String() || definitions[0].ApplicableCustomAssetTypeKeys[0] == medicine.ID.String() {
		t.Fatalf("model vocabulary must not expose internal IDs: %+v", definitions[0])
	}
	if _, err := catalog.resolve([]agentmodel.VoiceVocabularyRequest{{Kind: agentmodel.VoiceVocabularyKindCustomField, Key: "not-in-manifest"}}); err == nil {
		t.Fatal("expected an unscoped or invented vocabulary key to fail")
	}
	for _, hidden := range []string{"private-field", "other-field"} {
		if _, err := catalog.resolve([]agentmodel.VoiceVocabularyRequest{{Kind: agentmodel.VoiceVocabularyKindCustomField, Key: hidden}}); err == nil {
			t.Fatalf("expected hidden key %q to remain unavailable", hidden)
		}
	}
}

func seedHiddenVoiceVocabulary(t *testing.T, ctx context.Context, store *memory.Store, tenantID customfield.TenantID, inventoryID customfield.InventoryID, prefix string) {
	t.Helper()
	assetType, _ := customfield.NewAssetType(customfield.AssetTypeID(prefix+"-type-id"), tenantID, inventoryID, customfield.ScopeInventory, customfield.Key(prefix+"-type"), customfield.DisplayName("Hidden type"), "")
	if err := store.SaveCustomAssetType(ctx, assetType, audit.Record{ID: audit.ID(prefix + "-type-audit")}); err != nil {
		t.Fatalf("save hidden type: %v", err)
	}
	field, _ := customfield.NewDefinition(customfield.ID(prefix+"-field-id"), tenantID, inventoryID, customfield.ScopeInventory, customfield.Key(prefix+"-field"), customfield.DisplayName("Hidden field"), customfield.FieldTypeText, nil, customfield.ApplicabilityAllAssets, nil)
	if err := store.SaveCustomFieldDefinition(ctx, field, audit.Record{ID: audit.ID(prefix + "-field-audit")}); err != nil {
		t.Fatalf("save hidden field: %v", err)
	}
	tag, _ := assettag.NewTag(assettag.ID(prefix+"-tag-id"), assettag.TenantID(tenantID), assettag.InventoryID(inventoryID), assettag.Key(prefix+"-tag"), "Hidden tag", "", time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC))
	if err := store.CreateAssetTag(ctx, tag, audit.Record{ID: audit.ID(prefix + "-tag-audit")}); err != nil {
		t.Fatalf("save hidden tag: %v", err)
	}
}

func TestRealtimeVoiceInvestigationCarriesManifestThenOnlyRequestedDefinitions(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationLocate, SubjectMention: "camp medicine"}
	initial := agentmodel.InvestigationStep{
		Decision: agentmodel.InvestigationDecisionSearch, Intent: intent,
		SearchRequests:     []agentmodel.SearchRequest{{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "camp medicine", SearchProbes: []string{"camp medicine"}}},
		VocabularyRequests: []agentmodel.VoiceVocabularyRequest{{Kind: agentmodel.VoiceVocabularyKindCustomField, Key: "expiration-date"}},
	}
	final := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionStrong, CandidateIDs: []string{"medicine-1"}}}}
	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{{Investigation: &initial}, {Investigation: &final}}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.LanguageInference = language
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Where is the camp medicine?"}
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	application.customAssetTypes, application.customFields, application.assetTags = store, store, store

	medicine, _ := customfield.NewAssetType("type-medicine", "tenant-home", "inventory-home", customfield.ScopeInventory, "medicine", "Medicine", "Medication and supplements")
	if err := store.SaveCustomAssetType(context.Background(), medicine, audit.Record{ID: "vocab-loop-type"}); err != nil {
		t.Fatalf("save custom type: %v", err)
	}
	expires, _ := customfield.NewDefinition("field-expires", "tenant-home", "inventory-home", customfield.ScopeInventory, "expiration-date", "Expiration Date", customfield.FieldTypeDate, nil, customfield.ApplicabilityCustomAssetTypes, []customfield.AssetTypeID{medicine.ID})
	if err := store.SaveCustomFieldDefinition(context.Background(), expires, audit.Record{ID: "vocab-loop-field"}); err != nil {
		t.Fatalf("save field: %v", err)
	}
	seedRealtimeVoiceLoopAsset(t, store, realtimeVoiceInvestigationAsset("medicine-1", "Camp medicine", asset.KindItem, ""), "vocab-loop-asset")
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error { return nil }); err != nil {
		t.Fatalf("run loop: %v", err)
	}
	if len(language.seenInvestigations) != 2 || language.seenInvestigations[0].Vocabulary.CustomAssetTypes[0].Key != "medicine" {
		t.Fatalf("expected initial scoped manifest: %+v", language.seenInvestigations)
	}
	assessment := language.seenInvestigations[1]
	if len(assessment.VocabularyDefinitions) != 1 || assessment.VocabularyDefinitions[0].Key != "expiration-date" || len(assessment.VocabularyDefinitions[0].ApplicableCustomAssetTypeKeys) != 1 {
		t.Fatalf("expected only requested key-resolved definition: %+v", assessment)
	}
}
