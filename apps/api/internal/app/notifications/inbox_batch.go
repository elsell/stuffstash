package notifications

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
)

type UnreadCountPage struct {
	Count      int
	NextCursor string
}

type InboxReadPage struct{ NextCursor string }

func (s Service) CountUnreadPage(ctx context.Context, input ScopeInput, beforeID string) (UnreadCountPage, error) {
	page, err := s.listInbox(ctx, input, beforeID, maxInboxPageSize, true, false)
	if err != nil {
		return UnreadCountPage{}, err
	}
	return UnreadCountPage{Count: len(page.Items), NextCursor: page.NextCursor}, nil
}

func (s Service) MarkInboxPageRead(ctx context.Context, input ScopeInput, beforeID string) (InboxReadPage, error) {
	page, err := s.listInbox(ctx, input, beforeID, maxInboxPageSize, true, false)
	if err != nil {
		return InboxReadPage{}, err
	}
	for _, view := range page.Items {
		if err := s.MarkRead(ctx, input, view.Notification.ID); err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return InboxReadPage{}, err
		}
	}
	return InboxReadPage{NextCursor: page.NextCursor}, nil
}
