package httpserver

import (
	"net/http/httptest"
	"testing"
)

type attachmentResponse struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenantId"`
	InventoryID string `json:"inventoryId"`
	AssetID     string `json:"assetId"`
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
	SHA256      string `json:"sha256"`
	CreatedAt   string `json:"createdAt"`
}

type attachmentBody struct {
	Data attachmentResponse `json:"data"`
	Meta responseMeta       `json:"meta"`
}

type attachmentListBody struct {
	Data []attachmentResponse `json:"data"`
	Meta responseMeta         `json:"meta"`
}

func decodeAttachment(t *testing.T, response *httptest.ResponseRecorder) attachmentBody {
	t.Helper()

	var body attachmentBody
	decodeBody(t, response, &body)
	return body
}

func decodeAttachmentList(t *testing.T, response *httptest.ResponseRecorder) attachmentListBody {
	t.Helper()

	var body attachmentListBody
	decodeBody(t, response, &body)
	return body
}
