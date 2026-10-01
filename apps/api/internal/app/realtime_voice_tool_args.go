package app

import (
	"github.com/stuffstash/stuff-stash/internal/app/agentmodel/tools"
)

type realtimeVoiceSearchArgs = tools.RealtimeVoiceSearchArgs
type realtimeVoiceListArgs = tools.RealtimeVoiceListArgs
type realtimeVoiceAssetAuditHistoryArgs = tools.RealtimeVoiceAssetAuditHistoryArgs
type realtimeVoiceAssetDetailArgs = tools.RealtimeVoiceAssetDetailArgs
type realtimeVoiceCheckedOutAssetsArgs = tools.RealtimeVoiceCheckedOutAssetsArgs
type realtimeVoiceAssetCheckoutHistoryArgs = tools.RealtimeVoiceAssetCheckoutHistoryArgs

const (
	realtimeVoiceParentScopeAny  = tools.RealtimeVoiceParentScopeAny
	realtimeVoiceParentScopeRoot = tools.RealtimeVoiceParentScopeRoot
)

func parseRealtimeVoiceSearchArgs(args map[string]any) (realtimeVoiceSearchArgs, error) {
	return tools.ParseRealtimeVoiceSearchArgs(args)
}

func parseRealtimeVoiceListArgs(args map[string]any) (realtimeVoiceListArgs, error) {
	return tools.ParseRealtimeVoiceListArgs(args)
}

func parseRealtimeVoiceAssetAuditHistoryArgs(args map[string]any) (realtimeVoiceAssetAuditHistoryArgs, error) {
	return tools.ParseRealtimeVoiceAssetAuditHistoryArgs(args)
}

func parseRealtimeVoiceAssetDetailArgs(args map[string]any) (realtimeVoiceAssetDetailArgs, error) {
	return tools.ParseRealtimeVoiceAssetDetailArgs(args)
}

func parseRealtimeVoiceCheckedOutAssetsArgs(args map[string]any) (realtimeVoiceCheckedOutAssetsArgs, error) {
	return tools.ParseRealtimeVoiceCheckedOutAssetsArgs(args)
}

func parseRealtimeVoiceAssetCheckoutHistoryArgs(args map[string]any) (realtimeVoiceAssetCheckoutHistoryArgs, error) {
	return tools.ParseRealtimeVoiceAssetCheckoutHistoryArgs(args)
}

func rejectUnknownRealtimeVoiceArgs(args map[string]any, allowed ...string) error {
	return tools.RejectUnknownRealtimeVoiceArgs(args, allowed...)
}

func optionalRealtimeVoiceTitle(raw any) (string, error) {
	return tools.OptionalRealtimeVoiceTitle(raw)
}

func realtimeVoiceOptionalParentScope(raw any) (string, error) {
	return tools.RealtimeVoiceOptionalParentScope(raw)
}

type realtimeVoiceAssetToolOutput struct {
	NextCursor string                       `json:"nextCursor,omitempty"`
	Tool       string                       `json:"tool"`
	Query      string                       `json:"query,omitempty"`
	Filters    map[string]string            `json:"filters,omitempty"`
	Count      int                          `json:"count"`
	HasMore    bool                         `json:"hasMore,omitempty"`
	Note       string                       `json:"note,omitempty"`
	Items      []realtimeVoiceAssetToolItem `json:"items"`
}

type realtimeVoiceAssetToolItem struct {
	CustomAssetTypeID string                             `json:"customAssetTypeId,omitempty"`
	CustomFields      map[string]any                     `json:"customFields,omitempty"`
	Expiration        *realtimeVoiceExpiration           `json:"expiration"`
	TagNames          []string                           `json:"tagNames"`
	AssetID           string                             `json:"assetId,omitempty"`
	Title             string                             `json:"title"`
	Kind              string                             `json:"kind"`
	Description       string                             `json:"description,omitempty"`
	InventoryName     string                             `json:"inventoryName"`
	LifecycleState    string                             `json:"lifecycleState"`
	ParentAssetID     string                             `json:"parentAssetId,omitempty"`
	ParentTitle       string                             `json:"parentTitle,omitempty"`
	ParentKind        string                             `json:"parentKind,omitempty"`
	LocationTitle     string                             `json:"locationTitle,omitempty"`
	ContainmentPath   []string                           `json:"containmentPath,omitempty"`
	MatchFields       []string                           `json:"matchFields,omitempty"`
	CurrentCheckout   *realtimeVoiceCurrentCheckoutEntry `json:"currentCheckout,omitempty"`
	CheckoutState     *realtimeVoiceCheckoutState        `json:"checkoutState,omitempty"`
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
