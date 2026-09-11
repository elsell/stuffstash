package notifications

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const deviceGenerationPageSize = 100

func (s Service) deliveriesForNotification(ctx context.Context, value ports.NotificationRecord, enabled bool) ([]ports.NotificationDelivery, error) {
	deliveries := []ports.NotificationDelivery{}
	if !enabled {
		return deliveries, nil
	}
	if s.deps.Devices == nil {
		return nil, apperrors.ErrInvalidInput
	}
	after := ""
	for {
		devices, err := s.deps.Devices.ListNotificationDevices(ctx, value.Scope, after, deviceGenerationPageSize)
		if err != nil {
			return nil, err
		}
		if len(devices) > deviceGenerationPageSize {
			return nil, apperrors.ErrInvalidInput
		}
		for _, device := range devices {
			if device.ID <= after || device.Scope != value.Scope || !device.Active || device.Revision < 1 {
				return nil, apperrors.ErrInvalidInput
			}
			state, err := notification.NewDelivery(value.CreatedAt)
			if err != nil {
				return nil, err
			}
			deliveries = append(deliveries, ports.NotificationDelivery{ID: s.deps.IDs.NewID(), Scope: value.Scope, NotificationID: value.ID, DeviceID: device.ID, DeviceRevision: device.Revision, CreatedAt: value.CreatedAt, State: state})
			after = device.ID
		}
		if len(devices) < deviceGenerationPageSize {
			return deliveries, nil
		}
	}
}
