package httpserver

import (
	"net/http/httptest"
	"testing"
)

type auditRecordResponse struct {
	ID          string             `json:"id"`
	TenantID    string             `json:"tenantId"`
	InventoryID string             `json:"inventoryId,omitempty"`
	PrincipalID string             `json:"principalId"`
	Principal   *principalResponse `json:"principal,omitempty"`
	Action      string             `json:"action"`
	Source      string             `json:"source"`
	TargetType  string             `json:"targetType"`
	TargetID    string             `json:"targetId"`
	OccurredAt  string             `json:"occurredAt"`
	RequestID   string             `json:"requestId,omitempty"`
	Metadata    map[string]string  `json:"metadata"`
}

type principalResponse struct {
	ID    string `json:"id"`
	Email string `json:"email,omitempty"`
}

type auditRecordListBody struct {
	Data []auditRecordResponse `json:"data"`
	Meta responseMeta          `json:"meta"`
}

type assetActivityChangeResponse struct {
	Field         string `json:"field"`
	PreviousValue string `json:"previousValue,omitempty"`
	CurrentValue  string `json:"currentValue,omitempty"`
}

type assetActivityUndoResponse struct {
	OperationID string `json:"operationId"`
	Status      string `json:"status"`
}

type assetActivityResponse struct {
	ID                string                        `json:"id"`
	PrincipalID       string                        `json:"principalId"`
	Principal         *principalResponse            `json:"principal,omitempty"`
	Action            string                        `json:"action"`
	Category          string                        `json:"category"`
	Source            string                        `json:"source"`
	OccurredAt        string                        `json:"occurredAt"`
	RequestID         string                        `json:"requestId,omitempty"`
	Changes           []assetActivityChangeResponse `json:"changes"`
	Undo              *assetActivityUndoResponse    `json:"undo,omitempty"`
	TechnicalMetadata map[string]string             `json:"technicalMetadata"`
}

type assetActivityListBody struct {
	Data []assetActivityResponse `json:"data"`
	Meta responseMeta            `json:"meta"`
}

func decodeAssetActivityList(t *testing.T, response *httptest.ResponseRecorder) assetActivityListBody {
	t.Helper()
	var body assetActivityListBody
	decodeBody(t, response, &body)
	return body
}

func decodeAuditRecordList(t *testing.T, response *httptest.ResponseRecorder) auditRecordListBody {
	t.Helper()

	var body auditRecordListBody
	decodeBody(t, response, &body)
	return body
}

func auditRecordsContainTarget(records []auditRecordResponse, targetID string) bool {
	for _, record := range records {
		if record.TargetID == targetID {
			return true
		}
	}
	return false
}

func auditRecordsContainAction(records []auditRecordResponse, action string) bool {
	for _, record := range records {
		if record.Action == action {
			return true
		}
	}
	return false
}
