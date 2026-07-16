package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/audit/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
)

func AssetActivityToResponse(entry audit.AssetActivityEntry, users map[identity.PrincipalID]identity.User) dto.AssetActivityResponse {
	response := dto.AssetActivityResponse{
		ID: entry.ID.String(), PrincipalID: entry.PrincipalID.String(), Action: dto.AssetActivityAction(entry.Action.String()), Category: dto.AssetActivityCategory(entry.Category),
		Source: dto.AssetActivitySource(entry.Source.String()), OccurredAt: entry.OccurredAt, RequestID: entry.RequestID,
		Changes: make([]dto.AssetActivityChangeResponse, 0, len(entry.Changes)), TechnicalMetadata: entry.TechnicalMetadata,
	}
	for _, change := range entry.Changes {
		response.Changes = append(response.Changes, dto.AssetActivityChangeResponse{Field: dto.AssetActivityField(change.Field), PreviousValue: change.PreviousValue, CurrentValue: change.CurrentValue})
	}
	if entry.Undo != nil {
		response.Undo = &dto.AssetActivityUndoResponse{OperationID: entry.Undo.OperationID, Status: dto.AssetActivityUndoStatus(entry.Undo.Status)}
	}
	if user, ok := users[identity.PrincipalID(entry.PrincipalID.String())]; ok {
		response.Principal = &dto.AuditPrincipalResponse{ID: user.ID.String(), Email: user.Email.String()}
	}
	return response
}

func AssetActivitiesToResponse(entries []audit.AssetActivityEntry, users map[identity.PrincipalID]identity.User) []dto.AssetActivityResponse {
	data := make([]dto.AssetActivityResponse, 0, len(entries))
	for _, entry := range entries {
		data = append(data, AssetActivityToResponse(entry, users))
	}
	return data
}
