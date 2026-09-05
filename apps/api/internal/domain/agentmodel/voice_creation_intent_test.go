package agentmodel

import "testing"

func TestAdditionalCreationRequiresEvidenceAndCreateOperation(t *testing.T) {
	intent := Intent{RequestShape: RequestShapeSingleTarget, Kind: IntentKindChange, Operation: OperationCreate, SubjectMention: "Charger", NewAssetKind: "item", CreationMode: CreationModeAdditional}
	if intent.Validate() == nil {
		t.Fatal("additional creation accepted without evidence")
	}
	intent.CreationEvidence = "I bought another charger"
	if err := intent.Validate(); err != nil {
		t.Fatal(err)
	}
	intent.Operation = OperationLocate
	intent.Kind = IntentKindRead
	intent.NewAssetKind = ""
	if intent.Validate() == nil {
		t.Fatal("read carried creation intent")
	}
	intent.Operation = OperationCreate
	intent.Kind = IntentKindChange
	intent.NewAssetKind = "item"
	intent.CreationMode = CreationModeRecord
	if intent.Validate() == nil {
		t.Fatal("record mode retained additional evidence")
	}
	intent.CreationEvidence = ""
	if err := intent.Validate(); err != nil {
		t.Fatal(err)
	}
}
