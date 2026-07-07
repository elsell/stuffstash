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
	if err := rejectUnknownRealtimeVoiceArgs(args, "kind", "lifecycleState", "parentTitle", "locationTitle", "parentScope", "limit"); err != nil {
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
	parentScope, err := realtimeVoiceOptionalParentScope(args["parentScope"])
	if err != nil {
		return realtimeVoiceListArgs{}, err
	}
	if parentScope == realtimeVoiceParentScopeRoot && (parentTitle != "" || locationTitle != "") {
		return realtimeVoiceListArgs{}, ports.ErrInvalidProviderInput
	}
	return realtimeVoiceListArgs{Kind: kind, LifecycleState: lifecycleState, ParentTitle: parentTitle, LocationTitle: locationTitle, ParentScope: parentScope, Limit: limit}, nil
}

func parseRealtimeVoiceAssetAuditHistoryArgs(args map[string]any) (realtimeVoiceAssetAuditHistoryArgs, error) {
	if err := rejectUnknownRealtimeVoiceArgs(args, "assetId", "limit"); err != nil {
		return realtimeVoiceAssetAuditHistoryArgs{}, err
	}
	assetID := strings.TrimSpace(stringArg(args["assetId"]))
	if _, ok := asset.NewID(assetID); !ok {
		return realtimeVoiceAssetAuditHistoryArgs{}, ports.ErrInvalidProviderInput
	}
	limit, err := realtimeVoiceToolLimit(args["limit"])
	if err != nil {
		return realtimeVoiceAssetAuditHistoryArgs{}, err
	}
	return realtimeVoiceAssetAuditHistoryArgs{AssetID: assetID, Limit: limit}, nil
}

func parseRealtimeVoiceAssetDetailArgs(args map[string]any) (realtimeVoiceAssetDetailArgs, error) {
	if err := rejectUnknownRealtimeVoiceArgs(args, "assetId"); err != nil {
		return realtimeVoiceAssetDetailArgs{}, err
	}
	assetID := strings.TrimSpace(stringArg(args["assetId"]))
	if _, ok := asset.NewID(assetID); !ok {
		return realtimeVoiceAssetDetailArgs{}, ports.ErrInvalidProviderInput
	}
	return realtimeVoiceAssetDetailArgs{AssetID: assetID}, nil
}

func parseRealtimeVoiceCheckedOutAssetsArgs(args map[string]any) (realtimeVoiceCheckedOutAssetsArgs, error) {
	if err := rejectUnknownRealtimeVoiceArgs(args, "limit"); err != nil {
		return realtimeVoiceCheckedOutAssetsArgs{}, err
	}
	limit, err := realtimeVoiceToolLimit(args["limit"])
	if err != nil {
		return realtimeVoiceCheckedOutAssetsArgs{}, err
	}
	return realtimeVoiceCheckedOutAssetsArgs{Limit: limit}, nil
}

func parseRealtimeVoiceAssetCheckoutHistoryArgs(args map[string]any) (realtimeVoiceAssetCheckoutHistoryArgs, error) {
	if err := rejectUnknownRealtimeVoiceArgs(args, "assetId", "limit"); err != nil {
		return realtimeVoiceAssetCheckoutHistoryArgs{}, err
	}
	assetID := strings.TrimSpace(stringArg(args["assetId"]))
	if _, ok := asset.NewID(assetID); !ok {
		return realtimeVoiceAssetCheckoutHistoryArgs{}, ports.ErrInvalidProviderInput
	}
	limit, err := realtimeVoiceToolLimit(args["limit"])
	if err != nil {
		return realtimeVoiceAssetCheckoutHistoryArgs{}, err
	}
	return realtimeVoiceAssetCheckoutHistoryArgs{AssetID: assetID, Limit: limit}, nil
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

func realtimeVoiceOptionalParentScope(raw any) (string, error) {
	if raw == nil {
		return "", nil
	}
	value, ok := raw.(string)
	if !ok {
		return "", ports.ErrInvalidProviderInput
	}
	value = strings.TrimSpace(value)
	switch value {
	case "", realtimeVoiceParentScopeAny, realtimeVoiceParentScopeRoot:
		return value, nil
	default:
		return "", ports.ErrInvalidProviderInput
	}
}

const (
	realtimeVoiceParentScopeAny  = "any"
	realtimeVoiceParentScopeRoot = "root"
)

type realtimeVoiceSearchArgs struct {
	Query string
	Limit int
}

type realtimeVoiceListArgs struct {
	Kind           asset.Kind
	LifecycleState string
	ParentTitle    string
	LocationTitle  string
	ParentScope    string
	Limit          int
}

type realtimeVoiceAssetAuditHistoryArgs struct {
	AssetID string
	Limit   int
}

type realtimeVoiceAssetDetailArgs struct {
	AssetID string
}

type realtimeVoiceCheckedOutAssetsArgs struct {
	Limit int
}

type realtimeVoiceAssetCheckoutHistoryArgs struct {
	AssetID string
	Limit   int
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
	AssetID         string                             `json:"assetId,omitempty"`
	Title           string                             `json:"title"`
	Kind            string                             `json:"kind"`
	Description     string                             `json:"description,omitempty"`
	InventoryName   string                             `json:"inventoryName"`
	LifecycleState  string                             `json:"lifecycleState"`
	ParentTitle     string                             `json:"parentTitle,omitempty"`
	ParentKind      string                             `json:"parentKind,omitempty"`
	LocationTitle   string                             `json:"locationTitle,omitempty"`
	ContainmentPath []string                           `json:"containmentPath,omitempty"`
	MatchFields     []string                           `json:"matchFields,omitempty"`
	CurrentCheckout *realtimeVoiceCurrentCheckoutEntry `json:"currentCheckout,omitempty"`
	CheckoutState   *realtimeVoiceCheckoutState        `json:"checkoutState,omitempty"`
}

type realtimeVoiceCurrentCheckoutEntry struct {
	ID                      string `json:"id"`
	CheckedOutAt            string `json:"checkedOutAt"`
	CheckedOutByPrincipalID string `json:"checkedOutByPrincipalId"`
}

type realtimeVoiceCheckoutState struct {
	State        string `json:"state"`
	CheckedOut   bool   `json:"checkedOut"`
	CheckedOutAt string `json:"checkedOutAt,omitempty"`
}

type realtimeVoiceAssetAuditHistoryToolOutput struct {
	Tool    string                                `json:"tool"`
	Asset   realtimeVoiceAssetToolItem            `json:"asset"`
	Order   string                                `json:"order"`
	Count   int                                   `json:"count"`
	HasMore bool                                  `json:"hasMore,omitempty"`
	Note    string                                `json:"note,omitempty"`
	Entries []realtimeVoiceAssetAuditHistoryEntry `json:"entries"`
}

type realtimeVoiceAssetAuditHistoryEntry struct {
	Action              string `json:"action"`
	Source              string `json:"source"`
	OccurredAt          string `json:"occurredAt"`
	Actor               string `json:"actor,omitempty"`
	TargetType          string `json:"targetType"`
	AssetKind           string `json:"assetKind,omitempty"`
	PreviousParentTitle string `json:"previousParentTitle,omitempty"`
	NewParentTitle      string `json:"newParentTitle,omitempty"`
	PreviousState       string `json:"previousState,omitempty"`
	LifecycleState      string `json:"lifecycleState,omitempty"`
	Summary             string `json:"summary"`
}

type realtimeVoiceAssetCheckoutHistoryToolOutput struct {
	Tool    string                                   `json:"tool"`
	Asset   realtimeVoiceAssetToolItem               `json:"asset"`
	Order   string                                   `json:"order"`
	Count   int                                      `json:"count"`
	HasMore bool                                     `json:"hasMore,omitempty"`
	Note    string                                   `json:"note,omitempty"`
	Entries []realtimeVoiceAssetCheckoutHistoryEntry `json:"entries"`
}

type realtimeVoiceAssetCheckoutHistoryEntry struct {
	ID                      string `json:"id"`
	State                   string `json:"state"`
	CheckedOutAt            string `json:"checkedOutAt"`
	CheckedOutByPrincipalID string `json:"checkedOutByPrincipalId"`
	CheckoutDetails         string `json:"checkoutDetails,omitempty"`
	ReturnedAt              string `json:"returnedAt,omitempty"`
	ReturnedByPrincipalID   string `json:"returnedByPrincipalId,omitempty"`
	ReturnDetails           string `json:"returnDetails,omitempty"`
}
