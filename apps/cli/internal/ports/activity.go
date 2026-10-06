package ports

import (
	"context"
	"time"
)

type ActivityView string

const (
	ActivityChanges ActivityView = "changes"
	ActivityAll     ActivityView = "all"
)

type ActivityChange struct {
	Field         string  `json:"field"`
	PreviousValue *string `json:"previousValue,omitempty"`
	CurrentValue  *string `json:"currentValue,omitempty"`
}
type ActivityUndo struct {
	OperationID string `json:"operationId"`
	Status      string `json:"status"`
}
type Activity struct {
	ID                string            `json:"id"`
	Action            string            `json:"action"`
	Category          string            `json:"category"`
	Changes           []ActivityChange  `json:"changes"`
	OccurredAt        time.Time         `json:"occurredAt"`
	PrincipalID       string            `json:"principalId"`
	Principal         *AuditPrincipal   `json:"principal,omitempty"`
	RequestID         *string           `json:"requestId,omitempty"`
	Source            string            `json:"source"`
	TechnicalMetadata map[string]string `json:"technicalMetadata"`
	Undo              *ActivityUndo     `json:"undo,omitempty"`
}
type ActivityAPI interface {
	AssetActivity(context.Context, Scope, string, ActivityView, Page) (Result[[]Activity], error)
}
