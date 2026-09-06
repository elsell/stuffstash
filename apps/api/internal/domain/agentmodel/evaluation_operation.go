package agentmodel

// Operation identifies an actual inventory command in evaluation expectations
// and outcomes. It does not classify a user's request.
type Operation string

const (
	OperationCreate   Operation = "create"
	OperationMove     Operation = "move"
	OperationArchive  Operation = "archive"
	OperationRestore  Operation = "restore"
	OperationCheckout Operation = "checkout"
	OperationReturn   Operation = "return"
)

func (operation Operation) changesInventory() bool {
	switch operation {
	case OperationCreate, OperationMove, OperationArchive, OperationRestore, OperationCheckout, OperationReturn:
		return true
	default:
		return false
	}
}
