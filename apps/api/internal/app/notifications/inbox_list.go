package notifications

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
)

const inboxScanLimit = 200
const maxInboxPageSize = 100

type InboxPage struct {
	Items      []NotificationView
	NextCursor string
}

func (s Service) ListInbox(ctx context.Context, input ScopeInput, beforeID string, limit int, unreadOnly bool) (InboxPage, error) {
	return s.listInbox(ctx, input, beforeID, limit, unreadOnly, true)
}

func (s Service) listInbox(ctx context.Context, input ScopeInput, beforeID string, limit int, unreadOnly, includePlacement bool) (InboxPage, error) {
	if err := s.access(ctx, input); err != nil {
		return InboxPage{}, err
	}
	if limit < 1 || limit > maxInboxPageSize || s.deps.Inbox == nil || s.deps.Assets == nil || s.deps.Types == nil || s.deps.Clock == nil {
		return InboxPage{}, apperrors.ErrInvalidInput
	}
	entries, err := s.deps.Inbox.ListNotifications(ctx, input.Scope(), beforeID, inboxScanLimit+1)
	if err != nil {
		return InboxPage{}, err
	}
	page := InboxPage{Items: make([]NotificationView, 0, limit)}
	placement := placementCache{}
	for index, entry := range entries {
		if index == inboxScanLimit {
			break
		}
		if index+1 < len(entries) {
			page.NextCursor = entry.ID
		} else {
			page.NextCursor = ""
		}
		if unreadOnly && entry.ReadAt != nil {
			continue
		}
		view, err := s.currentNotification(ctx, input, entry.ID)
		if errors.Is(err, apperrors.ErrNotFound) {
			continue
		}
		if err != nil {
			return InboxPage{}, err
		}
		if unreadOnly && view.Notification.ReadAt != nil {
			continue
		}
		if includePlacement {
			if err := s.enrichPlacement(ctx, input, &view, placement); err != nil {
				return InboxPage{}, err
			}
		}
		page.Items = append(page.Items, view)
		if len(page.Items) == limit {
			break
		}
	}
	auditInput := s.auditInput(input, audit.ActionNotificationListed, input.InventoryID.String())
	auditInput.TargetType = audit.TargetNotification
	if err := appsupport.SaveReadAuditRecord(ctx, s.deps.Audit, s.deps.IDs, s.deps.Clock, auditInput); err != nil {
		return InboxPage{}, err
	}
	return page, nil
}
