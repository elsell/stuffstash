package httpserver

import (
	"net/http/httptest"
	"testing"
)

type auditRecordResponse struct {
	ID          string            `json:"id"`
	TenantID    string            `json:"tenantId"`
	InventoryID string            `json:"inventoryId,omitempty"`
	PrincipalID string            `json:"principalId"`
	Action      string            `json:"action"`
	Source      string            `json:"source"`
	TargetType  string            `json:"targetType"`
	TargetID    string            `json:"targetId"`
	OccurredAt  string            `json:"occurredAt"`
	RequestID   string            `json:"requestId,omitempty"`
	Metadata    map[string]string `json:"metadata"`
}

type auditRecordListBody struct {
	Data []auditRecordResponse `json:"data"`
	Meta responseMeta          `json:"meta"`
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
