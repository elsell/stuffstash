package bootstrap

func customizationHelp() []commandHelp {
	var commands []commandHelp
	for _, domain := range []string{"asset-types", "field-definitions"} {
		for _, action := range []string{"list", "show", "create", "update", "archive", "restore", "delete"} {
			c := commandHelp{Path: domain + " " + action, Scope: helpCustomization, Summary: action + " custom " + domain + ".", Options: "scope", Output: "Complete definition fields and response metadata; lists include pagination."}
			if action != "list" && action != "create" {
				c.Arguments = "ID"
			}
			if action == "list" {
				c.Options += " lifecycle limit cursor"
			}
			if action != "list" && action != "show" {
				c.Confirm = true
				c.Input = "Review the selected scope and target before confirming. No automatic retries; inspect list/show after an uncertain change."
			}
			if action == "create" || action == "update" {
				c.Options += " input name"
				c.Input = "Supply --input FILE|- for exact JSON, or use field options. Do not combine them. Interactive terminals prompt for common missing fields. JSON accepts optional $schema. "
				if action == "create" {
					c.Options += " key"
					c.Input += "Required: key and displayName. "
				} else {
					c.Input += "Update fields are optional; the server enforces compatible changes. "
				}
				if domain == "asset-types" {
					c.Options += " description expiration-enabled"
					c.Input += "Fields: displayName, description, expirationEnabled (boolean; --expiration-enabled=false disables). key is creation-only. "
				} else {
					if action == "create" {
						c.Options += " field-type"
						c.Input += "type is required: text, number, boolean, date, url, or enum. "
					}
					c.Input += "Fields: displayName, type, key, enumOptions (string array), applicability (all_assets/custom_asset_types), customAssetTypeIds (string array). Omit key and type from updates; they are immutable. The server enforces append-only options and targets and eligibility. Enum creation prompts for options, or use JSON. "
				}
				c.Input += "Scripts require --yes. Inspect current definitions before repeating an uncertain request."
				c.Example = domain + " " + action
				if action == "update" {
					c.Example += " ID"
				}
				c.Example += " --scope household --tenant HOME --input definition.json --yes"
			}
			commands = append(commands, c)
		}
	}
	return commands
}
