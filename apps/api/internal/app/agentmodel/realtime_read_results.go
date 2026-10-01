package agentmodel

type RealtimeVoiceAssetToolOutput struct {
	NextCursor string                       `json:"nextCursor,omitempty"`
	Tool       string                       `json:"tool"`
	Query      string                       `json:"query,omitempty"`
	Filters    map[string]string            `json:"filters,omitempty"`
	Count      int                          `json:"count"`
	HasMore    bool                         `json:"hasMore,omitempty"`
	Note       string                       `json:"note,omitempty"`
	Items      []RealtimeVoiceAssetToolItem `json:"items"`
}

type RealtimeVoiceAssetToolItem struct {
	CustomAssetTypeID string                             `json:"customAssetTypeId,omitempty"`
	CustomFields      map[string]any                     `json:"customFields,omitempty"`
	Expiration        *RealtimeVoiceExpiration           `json:"expiration"`
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
	CurrentCheckout   *RealtimeVoiceCurrentCheckoutEntry `json:"currentCheckout,omitempty"`
	CheckoutState     *RealtimeVoiceCheckoutState        `json:"checkoutState,omitempty"`
}

type RealtimeVoiceCurrentCheckoutEntry struct {
	ID                      string `json:"id"`
	CheckedOutAt            string `json:"checkedOutAt"`
	CheckedOutByPrincipalID string `json:"checkedOutByPrincipalId"`
}

type RealtimeVoiceCheckoutState struct {
	State        string `json:"state"`
	CheckedOut   bool   `json:"checkedOut"`
	CheckedOutAt string `json:"checkedOutAt,omitempty"`
}

type RealtimeVoiceAssetAuditHistoryToolOutput struct {
	Tool    string                                `json:"tool"`
	Asset   RealtimeVoiceAssetToolItem            `json:"asset"`
	Order   string                                `json:"order"`
	Count   int                                   `json:"count"`
	HasMore bool                                  `json:"hasMore,omitempty"`
	Note    string                                `json:"note,omitempty"`
	Entries []RealtimeVoiceAssetAuditHistoryEntry `json:"entries"`
}

type RealtimeVoiceAssetAuditHistoryEntry struct {
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

type RealtimeVoiceAssetCheckoutHistoryToolOutput struct {
	Tool    string                                   `json:"tool"`
	Asset   RealtimeVoiceAssetToolItem               `json:"asset"`
	Order   string                                   `json:"order"`
	Count   int                                      `json:"count"`
	HasMore bool                                     `json:"hasMore,omitempty"`
	Note    string                                   `json:"note,omitempty"`
	Entries []RealtimeVoiceAssetCheckoutHistoryEntry `json:"entries"`
}

type RealtimeVoiceAssetCheckoutHistoryEntry struct {
	ID                      string `json:"id"`
	State                   string `json:"state"`
	CheckedOutAt            string `json:"checkedOutAt"`
	CheckedOutByPrincipalID string `json:"checkedOutByPrincipalId"`
	CheckoutDetails         string `json:"checkoutDetails,omitempty"`
	ReturnedAt              string `json:"returnedAt,omitempty"`
	ReturnedByPrincipalID   string `json:"returnedByPrincipalId,omitempty"`
	ReturnDetails           string `json:"returnDetails,omitempty"`
}
