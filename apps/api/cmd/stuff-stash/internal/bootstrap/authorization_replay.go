package bootstrap

import (
	"context"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func replayLocalDevelopmentAuthorization(ctx context.Context, cfg config.Config, authorizer ports.Authorizer, repositories repositories) error {
	if strings.ToLower(strings.TrimSpace(cfg.AuthzMode)) != "memory" {
		return nil
	}
	if authorizer == nil || repositories.outbox == nil {
		return nil
	}

	events, err := repositories.outbox.ListAuthorizationOutboxReplayEvents(ctx)
	if err != nil {
		return err
	}
	for _, event := range events {
		if err := app.ApplyAuthorizationOutboxEvent(ctx, authorizer, event); err != nil {
			return err
		}
	}
	return nil
}
