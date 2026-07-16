package audit

import "time"

// AssetActivityCategory is a stable product-level grouping for asset history.
// It deliberately avoids making clients infer meaning from action strings.
type AssetActivityCategory string

const (
	AssetActivityCategoryChange AssetActivityCategory = "change"
	AssetActivityCategoryRead   AssetActivityCategory = "read"
)

type AssetActivityField string

const (
	AssetActivityFieldTitle          AssetActivityField = "title"
	AssetActivityFieldDescription    AssetActivityField = "description"
	AssetActivityFieldTags           AssetActivityField = "tags"
	AssetActivityFieldParent         AssetActivityField = "parent"
	AssetActivityFieldLifecycleState AssetActivityField = "lifecycle_state"
	AssetActivityFieldCheckoutState  AssetActivityField = "checkout_state"
)

type AssetActivityChange struct {
	Field         AssetActivityField
	PreviousValue string
	CurrentValue  string
}

type AssetActivityUndo struct {
	OperationID string
	Status      string
}

type AssetActivityEntry struct {
	ID                ID
	PrincipalID       PrincipalID
	Action            Action
	Category          AssetActivityCategory
	Source            Source
	OccurredAt        time.Time
	RequestID         string
	Changes           []AssetActivityChange
	Undo              *AssetActivityUndo
	TechnicalMetadata map[string]string
}

var assetActivityChangeActions = []Action{
	ActionAssetCreated,
	ActionAssetUpdated,
	ActionAssetMoved,
	ActionAssetArchived,
	ActionAssetRestored,
	ActionAssetDeleted,
	ActionAssetCheckedOut,
	ActionAssetReturned,
	ActionAssetReturnDetailsUpdated,
	ActionUndoableOperationUndone,
	ActionUndoableOperationRedone,
}

func AssetActivityChangeActions() []Action {
	actions := make([]Action, len(assetActivityChangeActions))
	copy(actions, assetActivityChangeActions)
	return actions
}

func (a Action) AssetActivityCategory() AssetActivityCategory {
	for _, changeAction := range assetActivityChangeActions {
		if a == changeAction {
			return AssetActivityCategoryChange
		}
	}
	return AssetActivityCategoryRead
}
