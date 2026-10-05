package ports

import (
	"context"
	"time"
)

type PrintConnector struct {
	ID                   string                `json:"id"`
	Name                 string                `json:"name"`
	Generation           int64                 `json:"generation"`
	AuthorizationPending bool                  `json:"authorizationPending"`
	Availability         string                `json:"availability"`
	State                string                `json:"state"`
	LastSeenAt           *time.Time            `json:"lastSeenAt,omitempty"`
	ReportReceivedAt     *time.Time            `json:"reportReceivedAt,omitempty"`
	PrinterIDs           []string              `json:"printerIds"`
	Report               *PrintConnectorReport `json:"report,omitempty"`
}
type PrintConnectorReport struct {
	Architecture string                  `json:"architecture"`
	Platform     string                  `json:"platform"`
	Commit       string                  `json:"commit"`
	Version      string                  `json:"version"`
	Adapters     []PrintConnectorAdapter `json:"adapters"`
}
type PrintConnectorAdapter struct {
	ID                 string                `json:"id"`
	CompletionEvidence string                `json:"completionEvidence"`
	ContractVersions   []int32               `json:"contractVersions"`
	Formats            []string              `json:"formats"`
	Wake               bool                  `json:"wake"`
	Media              []PrintConnectorMedia `json:"media"`
}
type PrintConnectorMedia struct {
	ID      string `json:"id"`
	Version int32  `json:"version"`
}
type ConnectorInspectionAPI interface {
	PrintConnectors(context.Context, Scope, Page) (Result[[]PrintConnector], error)
	PrintConnector(context.Context, Scope, string) (Result[PrintConnector], error)
}
