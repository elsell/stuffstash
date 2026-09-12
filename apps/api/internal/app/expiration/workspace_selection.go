package expiration

import (
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"slices"
	"strings"
)

type WorkspaceCounts struct {
	Soon    int
	Expired int
	All     int
}
type WorkspacePage struct {
	Items   []WorkspaceItem
	Counts  WorkspaceCounts
	HasMore bool
}

// WorkspaceSelection counts a complete authorized stream while retaining only
// the requested page and one lookahead item, independent of inventory size.
type WorkspaceSelection struct {
	mode   WorkspaceMode
	limit  int
	after  string
	counts WorkspaceCounts
	items  []WorkspaceItem
}

func NewWorkspaceSelection(mode WorkspaceMode, limit int, after string) *WorkspaceSelection {
	return &WorkspaceSelection{mode: mode, limit: limit, after: after, items: []WorkspaceItem{}}
}
func (s *WorkspaceSelection) Add(item WorkspaceItem) {
	if item.Asset.Expiration.Value() == "" {
		return
	}
	s.counts.All++
	soon := item.Expiration.TrackingEnabled && item.Expiration.State == expirationdate.Upcoming
	expired := item.Expiration.TrackingEnabled && item.Expiration.State == expirationdate.Expired
	if soon {
		s.counts.Soon++
	}
	if expired {
		s.counts.Expired++
	}
	if s.mode == WorkspaceSoon && !soon || s.mode == WorkspaceExpired && !expired {
		return
	}
	position := WorkspacePosition(item, s.mode)
	if position <= s.after {
		return
	}
	index, _ := slices.BinarySearchFunc(s.items, position, func(item WorkspaceItem, key string) int { return strings.Compare(WorkspacePosition(item, s.mode), key) })
	if index > s.limit {
		return
	}
	s.items = slices.Insert(s.items, index, item)
	if len(s.items) > s.limit+1 {
		s.items = s.items[:s.limit+1]
	}
}
func (s *WorkspaceSelection) Result() WorkspacePage {
	hasMore := len(s.items) > s.limit
	items := s.items
	if hasMore {
		items = items[:s.limit]
	}
	return WorkspacePage{Items: slices.Clone(items), Counts: s.counts, HasMore: hasMore}
}

// A lexical key orders expired dates descending ahead of future dates ascending.
// Inverting calendar digits retains a stable, precision-independent position.
func WorkspacePosition(item WorkspaceItem, mode WorkspaceMode) string {
	date := item.Asset.Expiration.LastValidDate()
	prefix := "1:"
	if item.Expiration.State == expirationdate.Expired {
		prefix = "0:"
		date = strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return '9' - (r - '0')
			}
			return r
		}, date)
	}
	return prefix + date + ":" + item.Asset.ID.String()
}
