package voice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestGoogleGeminiLanguageInferenceUsesStructuredInvestigationContract(t *testing.T) {
	t.Parallel()

	var request map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(geminiTextResponse(`{
          "decision":"search",
          "intent":{"requestShape":"single_target","kind":"read","operation":"locate","subjectMention":"Sarah winter coat","newAssetKind":"","destinationPath":[],"destinationKinds":[],"details":""},
          "searchRequests":[{"referenceKey":"subject","readKind":"search_assets","mention":"Sarah winter coat","kindHint":"","visibleAssetId":"","searchProbes":["Sarah winter coat","Sarah winter clothes","winter clothing"],"lifecycleScope":"active"}],
          "vocabularyRequests":[{"kind":"custom_asset_type","key":"winter-clothing"}],
          "resolutions":[],
          "rationale":"Gather authorized candidates for the remembered title."
        }`))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID: "project", Location: "us-central1", Model: "gemini-test",
		BaseURL: server.URL, TokenSource: staticTokenSource{}, HTTPClient: server.Client(),
	})
	turn, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{
		Transcript: "Where are Sarah's winter coat?",
		Investigation: &agentmodel.InvestigationInput{
			Phase: agentmodel.InvestigationPhaseInitial, PromptVersion: "voice-investigation-v1",
			SchemaVersion: "voice-investigation-v1", Transcript: "Where are Sarah's winter coat?",
			MaxEvidenceRounds: agentmodel.MaxEvidenceRounds,
			Vocabulary:        agentmodel.VoiceVocabularyManifest{CustomAssetTypes: []agentmodel.VoiceVocabularyAssetType{{Key: "winter-clothing", DisplayName: "Winter Clothing"}}},
		},
	})
	if err != nil {
		t.Fatalf("investigation turn: %v", err)
	}
	if turn.Investigation == nil || turn.Investigation.Intent.Operation != agentmodel.OperationLocate {
		t.Fatalf("unexpected investigation turn: %+v", turn)
	}
	if got := turn.Investigation.SearchRequests[0].SearchProbes; len(got) != 3 || got[1] != "Sarah winter clothes" {
		t.Fatalf("expected diverse model-owned probes, got %+v", got)
	}
	if turn.Investigation.SearchRequests[0].LifecycleScope != agentmodel.LifecycleScopeActive || len(turn.Investigation.VocabularyRequests) != 1 {
		t.Fatalf("expected lifecycle-scoped read and targeted vocabulary request, got %+v", turn.Investigation)
	}
	if _, exists := request["tools"]; exists {
		t.Fatalf("investigation must not expose provider-callable tools: %+v", request)
	}
	if _, exists := request["toolConfig"]; exists {
		t.Fatalf("investigation must not expose provider tool choice: %+v", request)
	}
	config := objectAt(t, request, "generationConfig")
	if config["responseMimeType"] != "application/json" || config["responseJsonSchema"] == nil {
		t.Fatalf("expected JSON-schema constrained investigation output, got %+v", config)
	}
	contents, ok := request["contents"].([]any)
	if !ok || len(contents) != 1 || !strings.Contains(string(mustJSON(t, contents[0])), "search hypotheses") {
		t.Fatalf("expected bounded investigation prompt, got %+v", request["contents"])
	}
	requestText := string(mustJSON(t, request))
	if !strings.Contains(requestText, "destinationKinds") || !strings.Contains(requestText, "do not rely on a segment's array position") || !strings.Contains(requestText, "winter-clothing") || !strings.Contains(requestText, "lifecycleScope") {
		t.Fatalf("expected ordered destination-kind contract in prompt and schema, got %s", requestText)
	}
}

func TestGoogleGeminiLanguageInferenceRejectsMissingInvestigationWithoutCallingProvider(t *testing.T) {
	t.Parallel()

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		_ = json.NewEncoder(w).Encode(geminiTextResponse(`{"final":{"kind":"answer","spokenResponse":"legacy","displayResponse":"legacy"}}`))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID: "project", Location: "us-central1", Model: "gemini-test",
		BaseURL: server.URL, TokenSource: staticTokenSource{}, HTTPClient: server.Client(),
	})
	_, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{Transcript: "Where are my tools?"})
	if !errors.Is(err, ports.ErrInvalidProviderInput) {
		t.Fatalf("expected missing typed investigation to be rejected, got %v", err)
	}
	if calls != 0 {
		t.Fatalf("missing investigation must fail before provider I/O, got %d calls", calls)
	}
}

func TestGeminiInvestigationPromptPreservesValidatedVocabularyAsCompleteUntrustedJSON(t *testing.T) {
	t.Parallel()

	manifest := agentmodel.VoiceVocabularyManifest{}
	for index := 0; index < agentmodel.MaxVoiceVocabularyAssetTypes; index++ {
		manifest.CustomAssetTypes = append(manifest.CustomAssetTypes, agentmodel.VoiceVocabularyAssetType{
			Key: fmt.Sprintf("secret-documents-%d", index), DisplayName: fmt.Sprintf("Password records %d", index), Description: strings.Repeat("x", 500),
		})
	}
	for index := 0; index < agentmodel.MaxVoiceVocabularyCustomFields; index++ {
		manifest.CustomFields = append(manifest.CustomFields, agentmodel.VoiceVocabularyFieldSummary{
			Key: fmt.Sprintf("api-token-location-%d", index), DisplayName: fmt.Sprintf("Credential location %d", index), FieldType: "text", Applicability: "all_assets",
		})
	}
	for index := 0; index < agentmodel.MaxVoiceVocabularyTags; index++ {
		manifest.Tags = append(manifest.Tags, agentmodel.VoiceVocabularyTag{Key: fmt.Sprintf("token-%d", index), DisplayName: fmt.Sprintf("Token %d", index)})
	}
	input := agentmodel.InvestigationInput{
		Phase: agentmodel.InvestigationPhaseInitial, PromptVersion: "voice-investigation-v1", SchemaVersion: "voice-investigation-v1",
		Transcript: "Where are the secret documents?", MaxEvidenceRounds: agentmodel.MaxEvidenceRounds, Vocabulary: manifest,
	}
	if err := input.Validate(); err != nil {
		t.Fatalf("max-sized investigation fixture must be valid: %v", err)
	}
	prompt := geminiInvestigationPrompt(ports.LanguageInferenceInput{Investigation: &input})
	const begin = "<BEGIN_UNTRUSTED_INVESTIGATION_JSON>\n"
	const end = "\n<END_UNTRUSTED_INVESTIGATION_JSON>"
	start := strings.Index(prompt, begin)
	finish := strings.Index(prompt, end)
	if start < 0 || finish <= start {
		t.Fatalf("expected explicit untrusted JSON boundary, got %s", prompt)
	}
	payload := prompt[start+len(begin) : finish]
	if len(payload) <= 24000 {
		t.Fatalf("fixture must prove payloads larger than the former truncation limit, got %d bytes", len(payload))
	}
	var decoded agentmodel.InvestigationInput
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		t.Fatalf("typed prompt payload was not complete JSON: %v", err)
	}
	if decoded.Vocabulary.CustomAssetTypes[0].Key != "secret-documents-0" || decoded.Vocabulary.CustomFields[0].Key != "api-token-location-0" || decoded.Vocabulary.Tags[0].Key != "token-0" {
		t.Fatalf("typed vocabulary keys were rewritten: %+v", decoded.Vocabulary)
	}
	if strings.Contains(payload, "[redacted]") || decoded.Vocabulary.CustomAssetTypes[0].Description != strings.Repeat("x", 500) {
		t.Fatalf("validated structured input was sanitized or truncated")
	}
}

func TestGeminiInvestigationSchemaAvoidsStatusSpecificResolutionBranches(t *testing.T) {
	t.Parallel()

	schema := geminiInvestigationResponseSchema(agentmodel.InvestigationInput{Phase: agentmodel.InvestigationPhaseEvidenceAssessment})
	resolution := schema.Properties["resolutions"].Items
	if resolution == nil {
		t.Fatal("resolution item schema is missing")
	}
	if len(resolution.AnyOf) != 0 {
		t.Fatalf("provider schema must not multiply states with per-status branches: %+v", resolution.AnyOf)
	}
	statuses := resolution.Properties["status"].Enum
	if len(statuses) != 7 {
		t.Fatalf("expected one bounded status enum, got %+v", statuses)
	}
	if candidateIDs := resolution.Properties["candidateIds"]; candidateIDs.Type != "array" || candidateIDs.MinItems != 0 {
		t.Fatalf("candidate status cardinality belongs to project validation, got %+v", candidateIDs)
	}
	for _, property := range []string{"searchRequests", "resolutions", "vocabularyRequests"} {
		if bounded := schema.Properties[property]; bounded.MinItems != 0 {
			t.Fatalf("%s phase cardinality belongs to project validation, got %+v", property, bounded)
		}
	}
}

func TestGeminiInitialInvestigationSchemaRequiresEvidenceSearchWithoutPrematureResolution(t *testing.T) {
	t.Parallel()

	schema := geminiInvestigationResponseSchema(agentmodel.InvestigationInput{Phase: agentmodel.InvestigationPhaseInitial})
	if decisions := schema.Properties["decision"].Enum; len(decisions) != 1 || decisions[0] != "search" {
		t.Fatalf("initial turn must always gather bounded evidence, got %+v", decisions)
	}
	if resolutions := schema.Properties["resolutions"]; resolutions.MaxItems == nil || *resolutions.MaxItems != 0 {
		t.Fatalf("initial resolutions must be structurally empty, got %+v", resolutions)
	}
	request := schema.Properties["searchRequests"].Items
	if request == nil {
		t.Fatal("initial read request schema is missing")
	}
	if visibleIDs := request.Properties["visibleAssetId"].Enum; len(visibleIDs) != 1 || visibleIDs[0] != "" {
		t.Fatalf("initial reads cannot target an unseen ID, got %+v", visibleIDs)
	}
	intent := schema.Properties["intent"]
	if shapes := intent.Properties["requestShape"].Enum; len(shapes) != 3 || shapes[0] != "single_target" || shapes[1] != "collection_target" || shapes[2] != "compound" {
		t.Fatalf("intent schema must require typed request shape, got %+v", shapes)
	}
	if !slices.Contains(intent.Required, "requestShape") {
		t.Fatalf("requestShape must be required, got %+v", intent.Required)
	}
	if !strings.Contains(intent.Properties["requestShape"].Description, "compound") || !strings.Contains(intent.Properties["operation"].Description, "physical custody") {
		t.Fatalf("semantic discriminators need local schema guidance, got shape=%q operation=%q", intent.Properties["requestShape"].Description, intent.Properties["operation"].Description)
	}
}

func TestParseGeminiInvestigationTurnCanonicalizesUnsafeRequestShapesToUnsupported(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		shape string
	}{
		{name: "collection mutation", shape: "collection_target"},
		{name: "compound request", shape: "compound"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			raw := `{"decision":"search","intent":{"requestShape":"` + test.shape + `","kind":"change","operation":"move","subjectMention":"tools","newAssetKind":"","destinationPath":["garage"],"destinationKinds":["location"],"details":""},"searchRequests":[{"referenceKey":"subject","readKind":"search_assets","mention":"tools","kindHint":"item","visibleAssetId":"","searchProbes":["tools"],"lifecycleScope":"active"},{"referenceKey":"destination.0","readKind":"search_assets","mention":"garage","kindHint":"location","visibleAssetId":"","searchProbes":["garage"],"lifecycleScope":"active"}],"resolutions":[],"vocabularyRequests":[],"rationale":"Classify request."}`
			turn, err := parseGeminiInvestigationTurn(raw)
			if err != nil {
				t.Fatalf("parse request shape: %v", err)
			}
			intent := turn.Investigation.Intent
			if intent.Operation != agentmodel.OperationUnsupported || intent.Kind != agentmodel.IntentKindUnsupported || string(intent.RequestShape) != test.shape {
				t.Fatalf("unsafe shape was not canonicalized to unsupported: %+v", intent)
			}
			if len(intent.DestinationPath) != 0 || len(turn.Investigation.SearchRequests) != 1 || turn.Investigation.SearchRequests[0].ReferenceKey != agentmodel.SemanticReferenceSubject {
				t.Fatalf("unsupported canonicalization retained executable destinations: %+v", turn.Investigation)
			}
		})
	}
}

func TestParseGeminiInvestigationTurnPreservesOrderedDestinationKinds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		path  string
		kinds string
		want  []agentmodel.DestinationKind
	}{
		{name: "single toolbox is a container", path: `["Toolbox"]`, kinds: `["container"]`, want: []agentmodel.DestinationKind{agentmodel.DestinationKindContainer}},
		{name: "single room is a location", path: `["Craft room"]`, kinds: `["location"]`, want: []agentmodel.DestinationKind{agentmodel.DestinationKindLocation}},
		{name: "nested path keeps semantic roles", path: `["Garage","Blue cabinet","Upper shelf"]`, kinds: `["location","container","container"]`, want: []agentmodel.DestinationKind{agentmodel.DestinationKindLocation, agentmodel.DestinationKindContainer, agentmodel.DestinationKindContainer}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			raw := `{"decision":"search","intent":{"requestShape":"single_target","kind":"change","operation":"move","subjectMention":"drill","newAssetKind":"","destinationPath":` + test.path + `,"destinationKinds":` + test.kinds + `,"details":""},"searchRequests":[{"referenceKey":"subject","readKind":"search_assets","mention":"drill","kindHint":"item","visibleAssetId":"","searchProbes":["drill"]}],"resolutions":[],"rationale":"Gather evidence."}`
			turn, err := parseGeminiInvestigationTurn(raw)
			if err != nil {
				t.Fatalf("parse investigation turn: %v", err)
			}
			got := turn.Investigation.Intent.DestinationKinds
			if len(got) != len(test.want) {
				t.Fatalf("unexpected destination kinds: got %+v want %+v", got, test.want)
			}
			for index := range got {
				if got[index] != test.want[index] {
					t.Fatalf("unexpected destination kind at %d: got %q want %q", index, got[index], test.want[index])
				}
			}
		})
	}
}

func TestParseGeminiInvestigationTurnClearsCreateOnlyKindFromOtherOperations(t *testing.T) {
	t.Parallel()

	raw := `{"decision":"search","intent":{"requestShape":"single_target","kind":"change","operation":"move","subjectMention":"drill","newAssetKind":"item","destinationPath":["garage"],"destinationKinds":["location"],"details":""},"searchRequests":[{"referenceKey":"subject","readKind":"search_assets","mention":"drill","kindHint":"item","visibleAssetId":"","searchProbes":["drill"],"lifecycleScope":"active"}],"resolutions":[],"vocabularyRequests":[],"rationale":"Gather evidence."}`
	turn, err := parseGeminiInvestigationTurn(raw)
	if err != nil {
		t.Fatalf("parse investigation turn: %v", err)
	}
	if turn.Investigation.Intent.NewAssetKind != "" {
		t.Fatalf("non-create intent retained create-only kind: %+v", turn.Investigation.Intent)
	}
}

func TestParseGeminiInvestigationTurnDropsDecisionIrrelevantCollections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		raw   string
		check func(*testing.T, agentmodel.InvestigationStep)
	}{
		{
			name: "search drops premature resolutions",
			raw:  `{"decision":"search","intent":{"requestShape":"single_target","kind":"read","operation":"locate","subjectMention":"drill","newAssetKind":"","destinationPath":[],"destinationKinds":[],"details":""},"searchRequests":[{"referenceKey":"subject","readKind":"search_assets","mention":"drill","kindHint":"item","visibleAssetId":"","searchProbes":["drill"],"lifecycleScope":"active"}],"resolutions":[{"referenceKey":"subject","status":"absent","candidateIds":[],"evidence":"premature"}],"vocabularyRequests":[],"rationale":"Gather evidence."}`,
			check: func(t *testing.T, step agentmodel.InvestigationStep) {
				if len(step.Resolutions) != 0 || len(step.SearchRequests) != 1 {
					t.Fatalf("search retained decision-irrelevant resolutions: %+v", step)
				}
			},
		},
		{
			name: "finish drops repeated reads and vocabulary requests",
			raw:  `{"decision":"finish","intent":{"requestShape":"single_target","kind":"read","operation":"locate","subjectMention":"drill","newAssetKind":"","destinationPath":[],"destinationKinds":[],"details":""},"searchRequests":[{"referenceKey":"subject","readKind":"search_assets","mention":"drill","kindHint":"item","visibleAssetId":"","searchProbes":["drill"],"lifecycleScope":"active"}],"resolutions":[{"referenceKey":"subject","status":"strong","candidateIds":["drill-1"],"evidence":"authorized candidate"}],"vocabularyRequests":[{"kind":"tag","key":"tools"}],"rationale":"Finish from evidence."}`,
			check: func(t *testing.T, step agentmodel.InvestigationStep) {
				if len(step.SearchRequests) != 0 || len(step.VocabularyRequests) != 0 || len(step.Resolutions) != 1 {
					t.Fatalf("finish retained decision-irrelevant requests: %+v", step)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			turn, err := parseGeminiInvestigationTurn(test.raw)
			if err != nil {
				t.Fatalf("parse investigation turn: %v", err)
			}
			test.check(t, *turn.Investigation)
		})
	}
}

func TestParseGeminiInvestigationTurnCanonicalizesOperationAndReadDiscriminators(t *testing.T) {
	t.Parallel()

	t.Run("list inventory drops structurally irrelevant probes", func(t *testing.T) {
		t.Parallel()
		raw := `{"decision":"search","intent":{"requestShape":"collection_target","kind":"read","operation":"list_inventory","subjectMention":"stuff","newAssetKind":"","destinationPath":[],"destinationKinds":[],"details":""},"searchRequests":[{"referenceKey":"subject","readKind":"list_inventory","mention":"stuff","kindHint":"item","visibleAssetId":"","searchProbes":["stuff","item"],"lifecycleScope":"active"}],"resolutions":[],"vocabularyRequests":[],"rationale":"List inventory."}`
		turn, err := parseGeminiInvestigationTurn(raw)
		if err != nil {
			t.Fatalf("parse list inventory: %v", err)
		}
		if got := turn.Investigation.SearchRequests[0].SearchProbes; len(got) != 0 {
			t.Fatalf("list inventory retained irrelevant probes: %+v", got)
		}
	})

	t.Run("non containment operation keeps subject and drops contextual destination", func(t *testing.T) {
		t.Parallel()
		raw := `{"decision":"search","intent":{"requestShape":"single_target","kind":"change","operation":"checkout","subjectMention":"garden shears","newAssetKind":"item","destinationPath":["yard"],"destinationKinds":["location"],"details":"using them in the yard"},"searchRequests":[{"referenceKey":"subject","readKind":"search_assets","mention":"garden shears","kindHint":"item","visibleAssetId":"","searchProbes":["garden shears"],"lifecycleScope":"active"},{"referenceKey":"destination.0","readKind":"search_assets","mention":"yard","kindHint":"location","visibleAssetId":"","searchProbes":["yard"],"lifecycleScope":"active"}],"resolutions":[],"vocabularyRequests":[],"rationale":"Find checkout subject."}`
		turn, err := parseGeminiInvestigationTurn(raw)
		if err != nil {
			t.Fatalf("parse checkout: %v", err)
		}
		step := turn.Investigation
		if len(step.Intent.DestinationPath) != 0 || len(step.SearchRequests) != 1 || step.SearchRequests[0].ReferenceKey != agentmodel.SemanticReferenceSubject {
			t.Fatalf("checkout retained contextual destination: %+v", step)
		}
	})

	t.Run("sole subject-like destination reference is relabeled", func(t *testing.T) {
		t.Parallel()
		raw := `{"decision":"search","intent":{"requestShape":"single_target","kind":"read","operation":"list_contents","subjectMention":"toolbox","newAssetKind":"","destinationPath":["toolbox"],"destinationKinds":["container"],"details":""},"searchRequests":[{"referenceKey":"destination.0","readKind":"list_inventory","mention":"toolbox","kindHint":"container","visibleAssetId":"","searchProbes":["toolbox"],"lifecycleScope":"active"}],"resolutions":[],"vocabularyRequests":[],"rationale":"Find container."}`
		turn, err := parseGeminiInvestigationTurn(raw)
		if err != nil {
			t.Fatalf("parse contents subject: %v", err)
		}
		step := turn.Investigation
		if len(step.Intent.DestinationPath) != 0 || len(step.SearchRequests) != 1 || step.SearchRequests[0].ReferenceKey != agentmodel.SemanticReferenceSubject || len(step.SearchRequests[0].SearchProbes) != 0 {
			t.Fatalf("contents subject was not canonicalized: %+v", step)
		}
	})
}

func TestGeminiInvestigationPromptDefinesGeneralContainmentAndCustodySemantics(t *testing.T) {
	t.Parallel()

	prompt := geminiInvestigationPrompt(ports.LanguageInferenceInput{Investigation: &agentmodel.InvestigationInput{
		Phase: agentmodel.InvestigationPhaseInitial, PromptVersion: "voice-investigation-v1", SchemaVersion: "voice-investigation-v1",
		Transcript: "generated request", MaxEvidenceRounds: agentmodel.MaxEvidenceRounds,
	}})
	for _, rule := range []string{"newly obtained subject cannot be moved", "Only create and move", "Usage, borrower, purpose", "imperative return", "physical-custody meaning", "programming or API request", "return operation", "past-tense location question", "placement verb alone", "Y before X", "not every named noun is a destination", "Spatial landmark relations", "[workshop, crate under the bench]", "must not search again merely to confirm absence"} {
		if !strings.Contains(prompt, rule) {
			t.Fatalf("prompt is missing general rule %q: %s", rule, prompt)
		}
	}
}

func TestParseGeminiInvestigationTurnDerivesLifecycleTransitionDiscoveryScope(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{"archive", "restore"} {
		raw := `{"decision":"search","intent":{"requestShape":"single_target","kind":"change","operation":"` + operation + `","subjectMention":"drill","newAssetKind":"","destinationPath":[],"destinationKinds":[],"details":""},"searchRequests":[{"referenceKey":"subject","readKind":"search_assets","mention":"drill","kindHint":"item","visibleAssetId":"","searchProbes":["drill"],"lifecycleScope":"active"}],"resolutions":[],"vocabularyRequests":[],"rationale":"Search lifecycle subject."}`
		turn, err := parseGeminiInvestigationTurn(raw)
		if err != nil {
			t.Fatalf("parse %s scope: %v", operation, err)
		}
		if got := turn.Investigation.SearchRequests[0].LifecycleScope; got != agentmodel.LifecycleScopeAll {
			t.Fatalf("%s subject scope = %q, want all", operation, got)
		}
	}
}

func TestParseGeminiInvestigationTurnDoesNotRewriteLifecycleScopeOnIDReads(t *testing.T) {
	t.Parallel()

	raw := `{"decision":"search_again","intent":{"requestShape":"single_target","kind":"change","operation":"archive","subjectMention":"drill","newAssetKind":"","destinationPath":[],"destinationKinds":[],"details":""},"searchRequests":[{"referenceKey":"subject","readKind":"asset_detail","mention":"drill","kindHint":"item","visibleAssetId":"drill-1","searchProbes":[],"lifecycleScope":"active"}],"resolutions":[],"vocabularyRequests":[],"rationale":"Read visible detail."}`
	turn, err := parseGeminiInvestigationTurn(raw)
	if err != nil {
		t.Fatalf("parse id read scope: %v", err)
	}
	if got := turn.Investigation.SearchRequests[0].LifecycleScope; got != agentmodel.LifecycleScopeActive {
		t.Fatalf("id read scope = %q, want provider-selected active scope", got)
	}
}

func TestParseGeminiInvestigationTurnRemovesOnlyExactStructuralDuplicates(t *testing.T) {
	t.Parallel()

	raw := `{"decision":"finish","intent":{"requestShape":"single_target","kind":"read","operation":"locate","subjectMention":"drill","newAssetKind":"","destinationPath":[],"destinationKinds":[],"details":""},"searchRequests":[],"resolutions":[{"referenceKey":"subject","status":"ambiguous","candidateIds":["drill-1","drill-1","drill-2"],"evidence":"two visible candidates"}],"vocabularyRequests":[],"rationale":"Clarify the candidates."}`
	turn, err := parseGeminiInvestigationTurn(raw)
	if err != nil {
		t.Fatalf("parse duplicate candidate IDs: %v", err)
	}
	if got := turn.Investigation.Resolutions[0].CandidateIDs; len(got) != 2 || got[0] != "drill-1" || got[1] != "drill-2" {
		t.Fatalf("expected stable exact deduplication, got %+v", got)
	}

	raw = `{"decision":"search","intent":{"requestShape":"single_target","kind":"read","operation":"locate","subjectMention":"drill","newAssetKind":"","destinationPath":[],"destinationKinds":[],"details":""},"searchRequests":[{"referenceKey":"subject","readKind":"search_assets","mention":"drill","kindHint":"item","visibleAssetId":"","searchProbes":["Drill"," drill ","cordless-drill","cordless drill"],"lifecycleScope":"active"},{"referenceKey":"subject","readKind":"search_assets","mention":"drill","kindHint":"item","visibleAssetId":"","searchProbes":["Drill"," drill ","cordless-drill","cordless drill"],"lifecycleScope":"active"}],"resolutions":[],"vocabularyRequests":[],"rationale":"Search."}`
	turn, err = parseGeminiInvestigationTurn(raw)
	if err != nil {
		t.Fatalf("parse duplicate reads and probes: %v", err)
	}
	if len(turn.Investigation.SearchRequests) != 1 || len(turn.Investigation.SearchRequests[0].SearchProbes) != 2 {
		t.Fatalf("expected exact duplicate reads and normalized probes to be removed, got %+v", turn.Investigation.SearchRequests)
	}
}

func TestGeminiInvestigationSchemaKeepsUpperBoundsInProjectValidation(t *testing.T) {
	t.Parallel()

	schema := geminiInvestigationResponseSchema(agentmodel.InvestigationInput{Phase: agentmodel.InvestigationPhaseEvidenceAssessment})
	intent := schema.Properties["intent"]
	searchRequests := schema.Properties["searchRequests"]
	for name, value := range map[string]*int{
		"destinationPath":    intent.Properties["destinationPath"].MaxItems,
		"destinationKinds":   intent.Properties["destinationKinds"].MaxItems,
		"searchRequests":     searchRequests.MaxItems,
		"searchProbes":       searchRequests.Items.Properties["searchProbes"].MaxItems,
		"resolutions":        schema.Properties["resolutions"].MaxItems,
		"candidateIds":       schema.Properties["resolutions"].Items.Properties["candidateIds"].MaxItems,
		"vocabularyRequests": schema.Properties["vocabularyRequests"].MaxItems,
	} {
		if value != nil {
			t.Fatalf("%s maxItems must remain in project validation for Gemini state limits, got %v", name, *value)
		}
	}
	initial := geminiInvestigationResponseSchema(agentmodel.InvestigationInput{Phase: agentmodel.InvestigationPhaseInitial})
	if got := initial.Properties["searchRequests"].MinItems; got != 1 {
		t.Fatalf("initial searchRequests minItems = %d, want 1", got)
	}
}

func TestParseGeminiInvestigationTurnDerivesRedundantKindFromOperation(t *testing.T) {
	t.Parallel()

	raw := `{"decision":"search","intent":{"requestShape":"single_target","kind":"change","operation":"asset_history","subjectMention":"drill","newAssetKind":"","destinationPath":[],"destinationKinds":[],"details":""},"searchRequests":[{"referenceKey":"subject","readKind":"search_assets","mention":"drill","kindHint":"item","visibleAssetId":"","searchProbes":["drill"],"lifecycleScope":"active"}],"resolutions":[],"vocabularyRequests":[],"rationale":"Search history subject."}`
	turn, err := parseGeminiInvestigationTurn(raw)
	if err != nil {
		t.Fatalf("parse operation with redundant kind mismatch: %v", err)
	}
	if turn.Investigation.Intent.Kind != agentmodel.IntentKindRead || turn.Investigation.Intent.Operation != agentmodel.OperationAssetHistory {
		t.Fatalf("expected operation-owned read kind, got %+v", turn.Investigation.Intent)
	}
}

func TestGoogleGeminiLanguageInferenceRejectsInvalidInvestigationPayload(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(geminiTextResponse(`{
          "decision":"finish",
          "intent":{"requestShape":"single_target","kind":"change","operation":"move","subjectMention":"drill","newAssetKind":"","destinationPath":["garage"],"destinationKinds":["location"],"details":""},
          "searchRequests":[],
          "resolutions":[{"referenceKey":"subject","status":"strong","candidateIds":["invented-id"],"evidence":"guess"}],
          "commands":[{"kind":"move_asset"}],
          "rationale":""
        }`))
	}))
	t.Cleanup(server.Close)
	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID: "project", Location: "us-central1", Model: "gemini-test",
		BaseURL: server.URL, TokenSource: staticTokenSource{}, HTTPClient: server.Client(),
	})
	_, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{
		Transcript: "Move the drill to the garage",
		Investigation: &agentmodel.InvestigationInput{
			Phase: agentmodel.InvestigationPhaseInitial, PromptVersion: "voice-investigation-v1",
			SchemaVersion: "voice-investigation-v1", Transcript: "Move the drill to the garage",
			MaxEvidenceRounds: agentmodel.MaxEvidenceRounds,
		},
	})
	if err == nil {
		t.Fatal("expected provider-authored commands and invalid initial finish to be rejected")
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal value: %v", err)
	}
	return payload
}
