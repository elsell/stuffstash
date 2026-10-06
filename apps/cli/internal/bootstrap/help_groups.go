package bootstrap

// Group summaries explain the task before users choose a command.
var helpGroupSummaries = map[string]string{
	"access-grants":            "Manage direct access to an inventory.",
	"account":                  "Inspect your signed-in account and access.",
	"archive-jobs":             "Back up or restore inventories, photos and files.",
	"asset-types":              "Manage the types of items in your inventory.",
	"assets":                   "Find, add, edit, move and check out items.",
	"attachments":              "Upload, download and manage photos and files.",
	"completion":               "Set up command completion for your shell.",
	"connectors":               "Set up and manage printer connections.",
	"connectors print":         "Register a print worker or run it on this computer.",
	"context":                  "Choose and remember a server, household and inventory.",
	"evaluation":               "Check model behavior with recorded test cases.",
	"evaluation cases":         "Manage model test cases.",
	"evaluation runs":          "Start and inspect model test runs.",
	"field-definitions":        "Manage custom fields for your items.",
	"import-jobs":              "Preview, import and track inventory data.",
	"inventories":              "Create, inspect and manage inventories.",
	"invitations":              "Invite people and manage inventory invitations.",
	"labels":                   "Create, render and look up item labels.",
	"notification-devices":     "Manage devices that receive notifications.",
	"notification-preferences": "Choose which reminders you receive.",
	"notifications":            "Read and manage your notifications.",
	"operations":               "Inspect recorded changes and undo or redo them.",
	"print-jobs":               "Submit and track label print jobs.",
	"print-settings":           "Choose inventory printers and label defaults.",
	"printers":                 "Set up printers, choose label media and print test labels.",
	"provider-profiles":        "Configure model providers and their credentials.",
	"server":                   "Inspect server information and sign-in options.",
	"tags":                     "Organize items with tags.",
	"telemetry":                "Send an explicit batch of client measurements.",
	"tenants":                  "Create and manage households.",
	"voice-provider":           "Choose the model provider used for voice.",
	"workflows":                "Manage model workflows and their versions.",
	"workflows revisions":      "Inspect, create and publish workflow versions.",
}

func helpGroupSummary(path []string) string {
	key := ""
	for _, part := range path {
		if key != "" {
			key += " "
		}
		key += part
	}
	if summary, ok := helpGroupSummaries[key]; ok {
		return summary
	}
	return "Show available commands."
}
