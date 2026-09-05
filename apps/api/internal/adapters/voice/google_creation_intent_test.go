package voice

import (
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
)

func TestGeminiInvestigationPreservesAdditionalCreationIntent(t *testing.T) {
	raw := `{"decision":"search","intent":{"requestShape":"single_target","kind":"change","operation":"create","subjectMention":"Charger","newAssetKind":"item","creationMode":"additional","creationEvidence":"I bought another charger","destinationPath":[],"destinationKinds":[],"details":""},"searchRequests":[{"referenceKey":"subject","readKind":"search_assets","mention":"Charger","kindHint":"","visibleAssetId":"","searchProbes":["Charger"],"lifecycleScope":"active"}],"resolutions":[],"vocabularyRequests":[],"rationale":"Check existing items before proposing another."}`
	turn, err := parseGeminiInvestigationTurn(raw)
	if err != nil || turn.Investigation == nil || turn.Investigation.Intent.CreationMode != agentmodel.CreationModeAdditional || turn.Investigation.Intent.CreationEvidence != "I bought another charger" {
		t.Fatalf("provider discarded creation intent: %+v %v", turn, err)
	}
	if _, err = parseGeminiInvestigationTurn(strings.Replace(raw, `"creationEvidence":"I bought another charger"`, `"creationEvidence":""`, 1)); err == nil {
		t.Fatal("provider accepted unsupported additional intent")
	}
}
