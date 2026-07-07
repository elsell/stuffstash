package actionplan

type State string

const (
	StateProposed  State = "proposed"
	StateApproved  State = "approved"
	StateCancelled State = "cancelled"
	StateExecuted  State = "executed"
	StateFailed    State = "failed"
)

func (s State) Valid() bool {
	switch s {
	case StateProposed, StateApproved, StateCancelled, StateExecuted, StateFailed:
		return true
	default:
		return false
	}
}

func (s State) Terminal() bool {
	switch s {
	case StateCancelled, StateExecuted, StateFailed:
		return true
	default:
		return false
	}
}

type CommandKind string

const (
	CommandKindCreateAsset    CommandKind = "create_asset"
	CommandKindCreateLocation CommandKind = "create_location"
	CommandKindMoveAsset      CommandKind = "move_asset"
	CommandKindUpdateAsset    CommandKind = "update_asset"
	CommandKindArchiveAsset   CommandKind = "archive_asset"
	CommandKindRestoreAsset   CommandKind = "restore_asset"
	CommandKindCheckoutAsset  CommandKind = "checkout_asset"
	CommandKindReturnAsset    CommandKind = "return_asset"
)

func (k CommandKind) Valid() bool {
	switch k {
	case CommandKindCreateAsset, CommandKindCreateLocation, CommandKindMoveAsset, CommandKindUpdateAsset, CommandKindArchiveAsset, CommandKindRestoreAsset, CommandKindCheckoutAsset, CommandKindReturnAsset:
		return true
	default:
		return false
	}
}
