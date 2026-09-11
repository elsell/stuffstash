package ports

import "github.com/stuffstash/stuff-stash/internal/domain/asset"

type NotificationAncestor struct {
	AssetID asset.ID
	Title   string
	Kind    asset.Kind
}
