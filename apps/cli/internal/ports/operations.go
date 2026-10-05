package ports

import "context"

type OperationAction string

const (
	UndoOperation OperationAction = "undo"
	RedoOperation OperationAction = "redo"
)

type OperationsAPI interface {
	ApplyOperation(context.Context, Scope, string, OperationAction) (Result[Asset], error)
}
