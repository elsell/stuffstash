package printing

import (
	"regexp"
	"time"
)

type InstanceID string
type LabelID string
type RenderID string

var opaqueID = regexp.MustCompile(`^[0-9A-HJKMNP-TV-Z]{26}$`)

func ValidOpaqueID(id string) bool { return opaqueID.MatchString(id) }

type Label struct {
	InstanceID  InstanceID
	ID          LabelID
	TenantID    string
	InventoryID string
	AssetID     string
	Tombstoned  bool
	CreatedAt   time.Time
}

type LabelRender struct {
	ID                   RenderID
	TenantID             string
	InventoryID          string
	AssetID              string
	LabelID              LabelID
	SelectionFingerprint string
	MediaFingerprint     string
	ContentType          string
	SHA256               string
	Content              []byte
	WidthPixels          int
	HeightPixels         int
	DisplayRotation      int
	CreatedAt            time.Time
	ExpiresAt            time.Time
}

func (r LabelRender) Clone() LabelRender { r.Content = append([]byte(nil), r.Content...); return r }

type LabelView struct {
	Label          Label
	URL            string
	LifecycleState string
}
