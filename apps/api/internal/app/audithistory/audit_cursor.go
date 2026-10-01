package audithistory

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type auditRecordCursorPayload struct {
	Version    int    `json:"v"`
	Collection string `json:"collection"`
	Scope      string `json:"scope"`
	LastID     string `json:"lastId"`
	OccurredAt string `json:"occurredAt"`
}

func encodeAuditRecordCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, record audit.Record) *string {
	payload, err := json.Marshal(auditRecordCursorPayload{
		Version:    appsupport.PaginationCursorVersion,
		Collection: "audit_records",
		Scope:      auditRecordCursorScope(tenantID, inventoryID),
		LastID:     record.ID.String(),
		OccurredAt: record.OccurredAt.UTC().Format(time.RFC3339Nano),
	})
	if err != nil {
		return nil
	}
	cursor := base64.RawURLEncoding.EncodeToString(payload)
	return &cursor
}

func decodeAuditRecordCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, cursor string) (time.Time, audit.ID, error) {
	cursor = strings.TrimSpace(cursor)
	if cursor == "" {
		return time.Time{}, audit.ID(""), nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, audit.ID(""), err
	}
	var payload auditRecordCursorPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return time.Time{}, audit.ID(""), err
	}
	if payload.Version != appsupport.PaginationCursorVersion || payload.Collection != "audit_records" || payload.Scope != auditRecordCursorScope(tenantID, inventoryID) || strings.TrimSpace(payload.LastID) == "" || strings.TrimSpace(payload.OccurredAt) == "" {
		return time.Time{}, audit.ID(""), apperrors.ErrInvalidInput
	}
	occurredAt, err := time.Parse(time.RFC3339Nano, payload.OccurredAt)
	if err != nil {
		return time.Time{}, audit.ID(""), err
	}
	id, ok := audit.NewID(payload.LastID)
	if !ok {
		return time.Time{}, audit.ID(""), apperrors.ErrInvalidInput
	}
	return occurredAt, id, nil
}

func auditRecordCursorScope(tenantID tenant.ID, inventoryID inventory.InventoryID) string {
	if inventoryID.String() == "" {
		return tenantID.String()
	}
	return tenantID.String() + ":" + inventoryID.String()
}
