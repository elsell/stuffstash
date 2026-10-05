package app

// missingResourceScope identifies commands that could need a saved account's
// resource selection. Server-wide label resolution needs no inventory scope.
func missingResourceScope(o Options) bool {
	if len(o.Command) >= 2 && o.Command[0] == "labels" && o.Command[1] == "resolve" {
		return false
	}
	if len(o.Command) >= 2 && o.Command[0] == "inventories" && o.Command[1] == "list" {
		return o.Scope.Tenant == ""
	}
	return o.Scope.Tenant == "" || o.Scope.Inventory == ""
}
