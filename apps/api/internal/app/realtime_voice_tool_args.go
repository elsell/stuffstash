package app

import (
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func parseRealtimeVoiceSearchArgs(args map[string]any) (realtimeVoiceSearchArgs, error) {
	if err := rejectUnknownRealtimeVoiceArgs(args, "query", "limit"); err != nil {
		return realtimeVoiceSearchArgs{}, err
	}
	query := strings.TrimSpace(stringArg(args["query"]))
	if query == "" || len(query) > 120 {
		return realtimeVoiceSearchArgs{}, ports.ErrInvalidProviderInput
	}
	limit, err := realtimeVoiceToolLimit(args["limit"])
	if err != nil {
		return realtimeVoiceSearchArgs{}, err
	}
	return realtimeVoiceSearchArgs{Query: query, Limit: limit}, nil
}

func parseRealtimeVoiceListArgs(args map[string]any) (realtimeVoiceListArgs, error) {
	if err := rejectUnknownRealtimeVoiceArgs(args, "kind", "lifecycleState", "parentTitle", "locationTitle", "limit"); err != nil {
		return realtimeVoiceListArgs{}, err
	}
	kind, err := realtimeVoiceOptionalAssetKind(args["kind"])
	if err != nil {
		return realtimeVoiceListArgs{}, err
	}
	parentTitle, err := optionalRealtimeVoiceTitle(args["parentTitle"])
	if err != nil {
		return realtimeVoiceListArgs{}, err
	}
	locationTitle, err := optionalRealtimeVoiceTitle(args["locationTitle"])
	if err != nil {
		return realtimeVoiceListArgs{}, err
	}
	limit, err := realtimeVoiceToolLimit(args["limit"])
	if err != nil {
		return realtimeVoiceListArgs{}, err
	}
	lifecycleState, err := realtimeVoiceOptionalLifecycleState(args["lifecycleState"])
	if err != nil {
		return realtimeVoiceListArgs{}, err
	}
	return realtimeVoiceListArgs{Kind: kind, LifecycleState: lifecycleState, ParentTitle: parentTitle, LocationTitle: locationTitle, Limit: limit}, nil
}

func rejectUnknownRealtimeVoiceArgs(args map[string]any, allowed ...string) error {
	allowedSet := map[string]struct{}{}
	for _, key := range allowed {
		allowedSet[key] = struct{}{}
	}
	for key := range args {
		if _, ok := allowedSet[key]; !ok {
			return ports.ErrInvalidProviderInput
		}
	}
	return nil
}

func optionalRealtimeVoiceTitle(raw any) (string, error) {
	if raw == nil {
		return "", nil
	}
	value, ok := raw.(string)
	if !ok {
		return "", ports.ErrInvalidProviderInput
	}
	value = strings.TrimSpace(value)
	if len(value) > 160 {
		return "", ports.ErrInvalidProviderInput
	}
	return value, nil
}

type realtimeVoiceSearchArgs struct {
	Query string
	Limit int
}

type realtimeVoiceListArgs struct {
	Kind           asset.Kind
	LifecycleState string
	ParentTitle    string
	LocationTitle  string
	Limit          int
}

type realtimeVoiceAssetToolOutput struct {
	Tool    string                       `json:"tool"`
	Query   string                       `json:"query,omitempty"`
	Filters map[string]string            `json:"filters,omitempty"`
	Count   int                          `json:"count"`
	HasMore bool                         `json:"hasMore,omitempty"`
	Note    string                       `json:"note,omitempty"`
	Items   []realtimeVoiceAssetToolItem `json:"items"`
}

type realtimeVoiceAssetToolItem struct {
	AssetID         string   `json:"assetId,omitempty"`
	Title           string   `json:"title"`
	Kind            string   `json:"kind"`
	Description     string   `json:"description,omitempty"`
	InventoryName   string   `json:"inventoryName"`
	LifecycleState  string   `json:"lifecycleState"`
	ParentTitle     string   `json:"parentTitle,omitempty"`
	ParentKind      string   `json:"parentKind,omitempty"`
	LocationTitle   string   `json:"locationTitle,omitempty"`
	ContainmentPath []string `json:"containmentPath,omitempty"`
	MatchFields     []string `json:"matchFields,omitempty"`
}
