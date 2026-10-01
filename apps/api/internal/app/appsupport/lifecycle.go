package appsupport

import (
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"strings"
)

func LifecycleFilter(value string) (ports.AssetLifecycleFilter, error) {
	switch strings.TrimSpace(value) {
	case "":
		return ports.AssetLifecycleFilterActive, nil
	case string(ports.AssetLifecycleFilterActive):
		return ports.AssetLifecycleFilterActive, nil
	case string(ports.AssetLifecycleFilterArchived):
		return ports.AssetLifecycleFilterArchived, nil
	case string(ports.AssetLifecycleFilterAll):
		return ports.AssetLifecycleFilterAll, nil
	default:
		return "", apperrors.ErrInvalidInput
	}
}
