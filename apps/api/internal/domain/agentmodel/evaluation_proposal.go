package agentmodel

import "strings"

type EvaluationProposal struct {
	Operation     Operation
	TargetID      string
	DestinationID string
	NewTitle      string
	NewKind       EvaluationFixtureKind
	Details       string
}

func validEvaluationProposal(value EvaluationProposal, assets map[string]EvaluationFixtureAsset) bool {
	if !workflowTextWithin(value.Details, MaxInvestigationDetailRunes, true) || strings.TrimSpace(value.Details) != value.Details {
		return false
	}
	if value.Details != "" && value.Operation != OperationCheckout && value.Operation != OperationReturn {
		return false
	}
	if !value.Operation.changesInventory() {
		return false
	}
	if value.DestinationID != "" {
		destination, found := assets[value.DestinationID]
		if !found || destination.Kind == EvaluationFixtureItem {
			return false
		}
		if value.Operation != OperationCreate && value.Operation != OperationMove {
			return false
		}
	}
	if value.Operation == OperationCreate {
		if value.TargetID != "" || !workflowTextWithin(value.NewTitle, 160, false) || strings.TrimSpace(value.NewTitle) != value.NewTitle {
			return false
		}
		switch value.NewKind {
		case EvaluationFixtureItem, EvaluationFixtureContainer, EvaluationFixtureLocation:
			return true
		default:
			return false
		}
	}
	if _, found := assets[value.TargetID]; !found || value.NewTitle != "" || value.NewKind != "" {
		return false
	}
	if value.Operation == OperationMove {
		for parent := value.DestinationID; parent != ""; parent = assets[parent].ParentID {
			if parent == value.TargetID {
				return false
			}
		}
	}
	return true
}
