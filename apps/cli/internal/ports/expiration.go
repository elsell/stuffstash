package ports

type ExpirationQuery struct {
	Page                                                                        Page
	Mode, Kind, CheckoutState, Query, TypeID, LocationID, FromDate, ThroughDate string
	TagIDs                                                                      []string
}
type ExpirationCounts struct {
	All     int64 `json:"all"`
	Expired int64 `json:"expired"`
	Soon    int64 `json:"soon"`
}
type ExpirationAncestor struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
type ExpirationItem struct {
	Asset
	AncestorPath []ExpirationAncestor `json:"ancestorPath"`
}
type ExpirationWorkspace struct {
	Counts   ExpirationCounts `json:"counts"`
	Items    []ExpirationItem `json:"items"`
	Timezone string           `json:"timezone"`
}
