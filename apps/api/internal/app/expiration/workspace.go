package expiration

import (
	"github.com/stuffstash/stuff-stash/internal/ports"
	"slices"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
)

type WorkspaceMode string

const (
	WorkspaceSoon    WorkspaceMode = "soon"
	WorkspaceExpired WorkspaceMode = "expired"
	WorkspaceAll     WorkspaceMode = "all"
)

func (m WorkspaceMode) Valid() bool {
	return m == WorkspaceSoon || m == WorkspaceExpired || m == WorkspaceAll
}

type WorkspaceItem struct {
	CheckedOut  bool
	Asset       asset.Asset
	Expiration  Description
	TagIDs      []string
	AncestorIDs []string
}
type WorkspaceFilter struct {
	Kind          asset.Kind                     `json:"kind,omitempty"`
	CheckoutState ports.AssetCheckoutStateFilter `json:"checkoutState,omitempty"`
	Text          string                         `json:"text,omitempty"`
	TypeID        string                         `json:"typeId,omitempty"`
	TagIDs        []string                       `json:"tagIds,omitempty"`
	LocationID    string                         `json:"locationId,omitempty"`
	FromDate      string                         `json:"fromDate,omitempty"`
	ThroughDate   string                         `json:"throughDate,omitempty"`
}

func (f WorkspaceFilter) Validate() error {
	if f.Kind != "" && f.Kind != asset.KindItem && f.Kind != asset.KindContainer && f.Kind != asset.KindLocation {
		return ErrInvalidQuery
	}
	if f.CheckoutState != "" && f.CheckoutState != ports.AssetCheckoutStateFilterAny && f.CheckoutState != ports.AssetCheckoutStateFilterAvailable && f.CheckoutState != ports.AssetCheckoutStateFilterCheckedOut {
		return ErrInvalidQuery
	}
	if len(f.Text) > 120 || len(f.TypeID) > 128 || len(f.LocationID) > 128 || len(f.TagIDs) > 50 {
		return ErrInvalidQuery
	}
	for _, id := range append([]string{f.TypeID, f.LocationID}, f.TagIDs...) {
		if len(id) > 128 || strings.TrimSpace(id) != id || strings.ContainsAny(id, "\x00\r\n") {
			return ErrInvalidQuery
		}
	}
	for _, id := range f.TagIDs {
		if id == "" {
			return ErrInvalidQuery
		}
	}
	return (Query{Status: QueryAll, FromDate: f.FromDate, ThroughDate: f.ThroughDate}).Validate()
}
func (f WorkspaceFilter) Matches(item WorkspaceItem) bool {
	if f.Kind != "" && item.Asset.Kind != f.Kind {
		return false
	}
	if f.CheckoutState == ports.AssetCheckoutStateFilterAvailable && item.CheckedOut || f.CheckoutState == ports.AssetCheckoutStateFilterCheckedOut && !item.CheckedOut {
		return false
	}
	if item.Asset.Expiration.Value() == "" {
		return false
	}
	if f.TypeID != "" && string(item.Asset.CustomAssetTypeID) != f.TypeID {
		return false
	}
	for _, id := range f.TagIDs {
		if !slices.Contains(item.TagIDs, id) {
			return false
		}
	}
	if f.LocationID != "" && !slices.Contains(item.AncestorIDs, f.LocationID) {
		return false
	}
	text := strings.ToLower(strings.TrimSpace(f.Text))
	if text != "" && !strings.Contains(strings.ToLower(item.Asset.Title.String()+" "+item.Asset.Description.String()), text) {
		return false
	}
	last := item.Asset.Expiration.LastValidDate()
	return (f.FromDate == "" || last >= f.FromDate) && (f.ThroughDate == "" || last <= f.ThroughDate)
}
