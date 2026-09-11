package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/notifications/dto"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func DeviceToResponse(value ports.NotificationDevice) dto.DeviceResponse {
	return dto.DeviceResponse{ID: value.ID, InstallationID: value.InstallationID, Transport: string(value.Transport), Revision: value.Revision, Active: value.Active}
}
