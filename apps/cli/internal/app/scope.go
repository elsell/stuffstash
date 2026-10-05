package app

// missingResourceScope identifies commands that could need a saved account's
// resource selection. Server-wide label resolution needs no inventory scope.
func missingResourceScope(o Options) bool {
	if isAccountCommand(o) {
		return false
	}
	if len(o.Command) >= 2 && o.Command[0] == "labels" && o.Command[1] == "resolve" {
		return false
	}
	if !requiresInventory(o) {
		return o.Scope.Tenant == ""
	}
	return o.Scope.Tenant == "" || o.Scope.Inventory == ""
}
