package tools

import (
	"math"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func ParseRealtimeVoiceSearchArgs(args map[string]any) (RealtimeVoiceSearchArgs, error) {
	if err := RejectUnknownRealtimeVoiceArgs(args, "query", "lifecycleState", "limit"); err != nil {
		return RealtimeVoiceSearchArgs{}, err
	}
	query := strings.TrimSpace(stringArg(args["query"]))
	if query == "" || len(query) > 120 {
		return RealtimeVoiceSearchArgs{}, ports.ErrInvalidProviderInput
	}
	limit, err := RealtimeVoiceToolLimit(args["limit"])
	if err != nil {
		return RealtimeVoiceSearchArgs{}, err
	}
	lifecycleState, err := RealtimeVoiceOptionalLifecycleState(args["lifecycleState"])
	if err != nil {
		return RealtimeVoiceSearchArgs{}, err
	}
	return RealtimeVoiceSearchArgs{Query: query, LifecycleState: lifecycleState, Limit: limit}, nil
}

func ParseRealtimeVoiceListArgs(args map[string]any) (RealtimeVoiceListArgs, error) {
	if err := RejectUnknownRealtimeVoiceArgs(args, "kind", "lifecycleState", "parentAssetId", "parentTitle", "locationTitle", "parentScope", "limit"); err != nil {
		return RealtimeVoiceListArgs{}, err
	}
	kind, err := RealtimeVoiceOptionalAssetKind(args["kind"])
	if err != nil {
		return RealtimeVoiceListArgs{}, err
	}
	parentTitle, err := OptionalRealtimeVoiceTitle(args["parentTitle"])
	if err != nil {
		return RealtimeVoiceListArgs{}, err
	}
	locationTitle, err := OptionalRealtimeVoiceTitle(args["locationTitle"])
	if err != nil {
		return RealtimeVoiceListArgs{}, err
	}
	parentAssetID := strings.TrimSpace(stringArg(args["parentAssetId"]))
	if parentAssetID != "" {
		if _, ok := asset.NewID(parentAssetID); !ok {
			return RealtimeVoiceListArgs{}, ports.ErrInvalidProviderInput
		}
	}
	limit, err := RealtimeVoiceToolLimit(args["limit"])
	if err != nil {
		return RealtimeVoiceListArgs{}, err
	}
	lifecycleState, err := RealtimeVoiceOptionalLifecycleState(args["lifecycleState"])
	if err != nil {
		return RealtimeVoiceListArgs{}, err
	}
	parentScope, err := RealtimeVoiceOptionalParentScope(args["parentScope"])
	if err != nil {
		return RealtimeVoiceListArgs{}, err
	}
	if parentScope == RealtimeVoiceParentScopeRoot && (parentAssetID != "" || parentTitle != "" || locationTitle != "") {
		return RealtimeVoiceListArgs{}, ports.ErrInvalidProviderInput
	}
	if parentAssetID != "" && (parentTitle != "" || locationTitle != "") {
		return RealtimeVoiceListArgs{}, ports.ErrInvalidProviderInput
	}
	return RealtimeVoiceListArgs{Kind: kind, LifecycleState: lifecycleState, ParentAssetID: parentAssetID, ParentTitle: parentTitle, LocationTitle: locationTitle, ParentScope: parentScope, Limit: limit}, nil
}

func ParseRealtimeVoiceAssetAuditHistoryArgs(args map[string]any) (RealtimeVoiceAssetAuditHistoryArgs, error) {
	if err := RejectUnknownRealtimeVoiceArgs(args, "assetId", "limit"); err != nil {
		return RealtimeVoiceAssetAuditHistoryArgs{}, err
	}
	assetID := strings.TrimSpace(stringArg(args["assetId"]))
	if _, ok := asset.NewID(assetID); !ok {
		return RealtimeVoiceAssetAuditHistoryArgs{}, ports.ErrInvalidProviderInput
	}
	limit, err := RealtimeVoiceToolLimit(args["limit"])
	if err != nil {
		return RealtimeVoiceAssetAuditHistoryArgs{}, err
	}
	return RealtimeVoiceAssetAuditHistoryArgs{AssetID: assetID, Limit: limit}, nil
}

func ParseRealtimeVoiceAssetDetailArgs(args map[string]any) (RealtimeVoiceAssetDetailArgs, error) {
	if err := RejectUnknownRealtimeVoiceArgs(args, "assetId"); err != nil {
		return RealtimeVoiceAssetDetailArgs{}, err
	}
	assetID := strings.TrimSpace(stringArg(args["assetId"]))
	if _, ok := asset.NewID(assetID); !ok {
		return RealtimeVoiceAssetDetailArgs{}, ports.ErrInvalidProviderInput
	}
	return RealtimeVoiceAssetDetailArgs{AssetID: assetID}, nil
}

func ParseRealtimeVoiceCheckedOutAssetsArgs(args map[string]any) (RealtimeVoiceCheckedOutAssetsArgs, error) {
	if err := RejectUnknownRealtimeVoiceArgs(args, "limit"); err != nil {
		return RealtimeVoiceCheckedOutAssetsArgs{}, err
	}
	limit, err := RealtimeVoiceToolLimit(args["limit"])
	if err != nil {
		return RealtimeVoiceCheckedOutAssetsArgs{}, err
	}
	return RealtimeVoiceCheckedOutAssetsArgs{Limit: limit}, nil
}

func ParseRealtimeVoiceAssetCheckoutHistoryArgs(args map[string]any) (RealtimeVoiceAssetCheckoutHistoryArgs, error) {
	if err := RejectUnknownRealtimeVoiceArgs(args, "assetId", "limit"); err != nil {
		return RealtimeVoiceAssetCheckoutHistoryArgs{}, err
	}
	assetID := strings.TrimSpace(stringArg(args["assetId"]))
	if _, ok := asset.NewID(assetID); !ok {
		return RealtimeVoiceAssetCheckoutHistoryArgs{}, ports.ErrInvalidProviderInput
	}
	limit, err := RealtimeVoiceToolLimit(args["limit"])
	if err != nil {
		return RealtimeVoiceAssetCheckoutHistoryArgs{}, err
	}
	return RealtimeVoiceAssetCheckoutHistoryArgs{AssetID: assetID, Limit: limit}, nil
}

func RejectUnknownRealtimeVoiceArgs(args map[string]any, allowed ...string) error {
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

func OptionalRealtimeVoiceTitle(raw any) (string, error) {
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

func RealtimeVoiceOptionalParentScope(raw any) (string, error) {
	if raw == nil {
		return "", nil
	}
	value, ok := raw.(string)
	if !ok {
		return "", ports.ErrInvalidProviderInput
	}
	value = strings.TrimSpace(value)
	switch value {
	case "", RealtimeVoiceParentScopeAny, RealtimeVoiceParentScopeRoot:
		return value, nil
	default:
		return "", ports.ErrInvalidProviderInput
	}
}

const (
	RealtimeVoiceParentScopeAny  = "any"
	RealtimeVoiceParentScopeRoot = "root"
)

type RealtimeVoiceSearchArgs struct {
	Query          string
	LifecycleState string
	Limit          int
}

type RealtimeVoiceListArgs struct {
	Kind           asset.Kind
	LifecycleState string
	ParentAssetID  string
	ParentTitle    string
	LocationTitle  string
	ParentScope    string
	Limit          int
}

type RealtimeVoiceAssetAuditHistoryArgs struct {
	AssetID string
	Limit   int
}

type RealtimeVoiceAssetDetailArgs struct {
	AssetID string
}

type RealtimeVoiceCheckedOutAssetsArgs struct {
	Limit int
}

type RealtimeVoiceAssetCheckoutHistoryArgs struct {
	AssetID string
	Limit   int
}

func RealtimeVoiceToolLimit(raw any) (int, error) {
	if raw == nil {
		return 10, nil
	}
	switch value := raw.(type) {
	case float64:
		if math.IsNaN(value) || value != math.Trunc(value) || value < 1 {
			return 0, ports.ErrInvalidProviderInput
		}
		if value > RealtimeVoiceToolMaxResults {
			return RealtimeVoiceToolMaxResults, nil
		}
		return int(value), nil
	case int:
		if value < 1 {
			return 0, ports.ErrInvalidProviderInput
		}
		if value > RealtimeVoiceToolMaxResults {
			return RealtimeVoiceToolMaxResults, nil
		}
		return value, nil
	default:
		return 0, ports.ErrInvalidProviderInput
	}
}

func RealtimeVoiceOptionalAssetKind(raw any) (asset.Kind, error) {
	if raw == nil {
		return "", nil
	}
	value, ok := raw.(string)
	if !ok {
		return "", ports.ErrInvalidProviderInput
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	kind, ok := asset.NewKind(value)
	if !ok {
		return "", ports.ErrInvalidProviderInput
	}
	return kind, nil
}

func RealtimeVoiceOptionalLifecycleState(raw any) (string, error) {
	if raw == nil {
		return "active", nil
	}
	value, ok := raw.(string)
	if !ok {
		return "", ports.ErrInvalidProviderInput
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "active", nil
	}
	switch value {
	case "active", "archived", "all":
		return value, nil
	default:
		return "", ports.ErrInvalidProviderInput
	}
}

const RealtimeVoiceToolMaxResults = 20

func stringArg(raw any) string { value, _ := raw.(string); return value }
