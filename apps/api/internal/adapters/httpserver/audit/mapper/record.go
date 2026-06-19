package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/audit/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
)

func RecordToResponse(record audit.Record) dto.RecordResponse {
	return dto.RecordResponse{
		ID:          record.ID.String(),
		TenantID:    record.TenantID.String(),
		InventoryID: record.InventoryID.String(),
		PrincipalID: record.PrincipalID.String(),
		Action:      record.Action.String(),
		Source:      record.Source.String(),
		TargetType:  record.TargetType.String(),
		TargetID:    record.TargetID,
		OccurredAt:  record.OccurredAt,
		RequestID:   record.RequestID,
		Metadata:    record.MetadataValues(),
	}
}

func RecordsToResponse(records []audit.Record) []dto.RecordResponse {
	data := make([]dto.RecordResponse, 0, len(records))
	for _, record := range records {
		data = append(data, RecordToResponse(record))
	}
	return data
}
