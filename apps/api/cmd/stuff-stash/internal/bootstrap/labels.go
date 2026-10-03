package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	"github.com/stuffstash/stuff-stash/internal/adapters/labelrenderer"
	labelapp "github.com/stuffstash/stuff-stash/internal/app/printing"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func buildLabels(cfg config.Config, repos repositories, authorizer ports.Authorizer, observer ports.Observer) (*labelapp.LabelService, error) {
	settings, err := cfg.Labels.Settings()
	if err != nil {
		return nil, err
	}
	renderer, err := labelrenderer.New(labelrenderer.Limits{MaxPixels: settings.MaxPixels, MaxURLBytes: settings.MaxURLBytes, MaxTitleRunes: settings.MaxTitleRunes, MaxReferenceRunes: settings.MaxReferenceRunes})
	if err != nil {
		return nil, err
	}
	return labelapp.NewLabelService(labelapp.LabelDependencies{Repository: repos.labels, Renders: repos.labelRenders, Assets: repos.assets, Inventories: repos.inventories, Tenants: repos.tenants, Authorizer: authorizer, Audit: repos.audit, IDs: idgen.NewULIDGenerator(), Clock: ports.SystemClock{}, Observer: observer, Renderer: renderer, Templates: renderer, BaseURL: settings.BaseURL, RenderTTL: settings.RenderTTL, MaxRenderBytes: settings.MaxRenderBytes}), nil
}
func RunLabelsCommand(ctx context.Context, cfg config.Config, args []string, output io.Writer) error {
	if len(args) != 1 || args[0] != "bootstrap-instance" {
		return errors.New("labels requires bootstrap-instance")
	}
	if strings.TrimSpace(cfg.DatabaseDSN) == "" {
		return errors.New("database dsn is required")
	}
	var repository ports.LabelRepository
	var closeStore func() error
	switch strings.ToLower(strings.TrimSpace(cfg.RepositoryMode)) {
	case "postgres":
		store, close, err := openPostgresStore(ctx, cfg.DatabaseDSN)
		if err != nil {
			return err
		}
		repository = store
		closeStore = close
	case "sqlite":
		store, close, err := openSQLiteStore(ctx, cfg.DatabaseDSN)
		if err != nil {
			return err
		}
		repository = store
		closeStore = close
	default:
		return errors.New("instance bootstrap requires persistent postgres or sqlite storage")
	}
	defer closeStore()
	service := labelapp.NewLabelService(labelapp.LabelDependencies{Repository: repository, IDs: idgen.NewULIDGenerator()})
	instance, err := service.BootstrapInstance(ctx)
	if err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(map[string]any{"protocolVersion": 1, "instanceId": instance})
}
