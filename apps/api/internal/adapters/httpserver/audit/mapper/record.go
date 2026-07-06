package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/audit/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
)

func RecordToResponse(record audit.Record, users map[identity.PrincipalID]identity.User) dto.RecordResponse {
	response := dto.RecordResponse{
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
	if user, ok := users[identity.PrincipalID(record.PrincipalID.String())]; ok {
		response.Principal = &dto.AuditPrincipalResponse{
			ID:    user.ID.String(),
			Email: user.Email.String(),
		}
	}
	return response
}

func RecordsToResponse(records []audit.Record, users map[identity.PrincipalID]identity.User) []dto.RecordResponse {
	data := make([]dto.RecordResponse, 0, len(records))
	for _, record := range records {
		data = append(data, RecordToResponse(record, users))
	}
	return data
}
