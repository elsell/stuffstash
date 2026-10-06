package bootstrap

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/stuffstash/stuff-stash/cli/internal/app"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type helpScope uint8

const (
	helpLocal helpScope = iota
	helpServer
	helpAccount
	helpHousehold
	helpInventory
	helpConnector
	helpCustomization
)

type commandHelp struct {
	Path, Arguments, Summary        string
	Scope                           helpScope
	Options, Input, Output, Example string
	Confirm                         bool
}

func helpCatalog() []commandHelp {
	var commands []commandHelp
	for _, group := range [][]commandHelp{localHelp(), inventoryHelp(), accessHelp(), administrationHelp(), printingHelp(), customizationHelp()} {
		commands = append(commands, group...)
	}
	return commands
}
func helpCommand(command []string) (commandHelp, bool) {
	for _, entry := range helpCatalog() {
		path := strings.Fields(entry.Path)
		if len(command) >= len(path) && strings.Join(command[:len(path)], " ") == entry.Path {
			return entry, true
		}
	}
	return commandHelp{}, false
}
func writeHelp(w io.Writer, command []string) error {
	var text string
	if entry, ok := helpCommand(command); ok {
		text = commandHelpText(entry)
	} else {
		var err error
		text, err = groupHelpText(command)
		if err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, text)
	return err
}
func groupHelpText(command []string) (string, error) {
	prefix := strings.Join(command, " ")
	children := map[string]string{}
	for _, entry := range helpCatalog() {
		path := strings.Fields(entry.Path)
		if len(path) <= len(command) || strings.Join(path[:len(command)], " ") != prefix {
			continue
		}
		name := path[len(command)]
		if len(path) == len(command)+1 {
			children[name] = entry.Summary
		} else if _, ok := children[name]; !ok {
			children[name] = "Browse " + name + " commands."
		}
	}
	if len(children) == 0 {
		return "", ports.Failure("usage", "Unknown command. Run stuffstash --help, or add --help to a supported group.")
	}
	var b strings.Builder
	title := "Stuff Stash CLI"
	if prefix != "" {
		title = "stuffstash " + prefix
	}
	fmt.Fprintf(&b, "%s\n\nCommands:\n", title)
	names := make([]string, 0, len(children))
	for name := range children {
		names = append(names, name)
	}
	sort.Strings(names)
	table := tabwriter.NewWriter(&b, 0, 4, 2, ' ', 0)
	for _, name := range names {
		fmt.Fprintf(table, "  %s\t%s\n", name, children[name])
	}
	table.Flush()
	fmt.Fprintf(&b, "\nUse stuffstash %sCOMMAND --help for usage, options and examples.\n", func() string {
		if prefix == "" {
			return ""
		}
		return prefix + " "
	}())
	if prefix == "" {
		b.WriteString("Help is local and never signs in or contacts a server.\n")
	}
	return b.String(), nil
}
func commandHelpText(c commandHelp) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\nUsage:\n  stuffstash %s", c.Summary, c.Path)
	if c.Arguments != "" {
		fmt.Fprintf(&b, " %s", c.Arguments)
	}
	b.WriteString(" [options]\n")
	b.WriteString("\nScope:\n")
	helpParagraph(&b, scopeHelp(c.Scope))
	b.WriteString("\nOptions:\n")
	definitions := app.HelpOptions()
	table := tabwriter.NewWriter(&b, 0, 4, 2, ' ', 0)
	for _, name := range commandHelpOptions(c) {
		option := definitions[name]
		flag := "--" + name
		if option.Value != "" {
			flag += " " + helpOptionValue(name)
		}
		fmt.Fprintf(table, "  %s\t%s\n", flag, option.Description)
	}
	table.Flush()
	input := c.Input
	if input == "" {
		input = "No request body or input file."
	}
	output := c.Output
	if output == "" {
		output = "Human-readable results by default; --json emits JSON. Diagnostics go to stderr."
	}
	confirmation := "Not required."
	if c.Confirm {
		confirmation = "Review and confirm in an interactive terminal, or supply --yes. JSON, redirected or --no-input runs require --yes."
	}
	for _, section := range []struct{ title, text string }{{"Input", input}, {"Output", output}, {"Confirmation", confirmation}} {
		fmt.Fprintf(&b, "\n%s:\n", section.title)
		helpParagraph(&b, section.text)
	}
	example := c.Example
	if example == "" {
		example = c.Path
		if c.Arguments != "" {
			example += " " + c.Arguments
		}
		if c.Confirm {
			example += " --yes"
		}
	}
	fmt.Fprintf(&b, "\nExample:\n  stuffstash %s\n", example)
	return b.String()
}
func commandHelpOptions(c commandHelp) []string {
	names := strings.Fields(c.Options)
	if c.Confirm {
		names = append(names, "yes")
	}
	switch c.Scope {
	case helpServer, helpAccount, helpHousehold, helpInventory, helpCustomization:
		names = append(names, "server", "context")
	case helpConnector:
		names = append(names, "server")
	}
	if c.Scope == helpHousehold || c.Scope == helpInventory || c.Scope == helpCustomization {
		names = append(names, "tenant")
	}
	if c.Scope == helpInventory || c.Scope == helpCustomization {
		names = append(names, "inventory")
	}
	if c.Scope != helpLocal && c.Scope != helpConnector {
		names = append(names, "request-id")
	}
	if c.Path != "connectors print run" {
		names = append(names, "json", "no-input", "color")
	}
	return append(names, "help")
}
func scopeHelp(scope helpScope) string {
	switch scope {
	case helpLocal:
		return "Local machine only; no server, sign-in or resource scope."
	case helpServer:
		return "Server from --server, STUFF_STASH_CLI_SERVER or the saved context; no household or inventory selection."
	case helpAccount:
		return "Signed-in account and server; no household or inventory selection. Use --server and --context for the destination."
	case helpHousehold:
		return "Signed-in household from --tenant, STUFF_STASH_CLI_TENANT or the same account's saved context. Interactive terminals can choose a household. Scripts must supply or remember it."
	case helpInventory:
		return "Signed-in household and inventory from --tenant/--inventory, STUFF_STASH_CLI_TENANT/STUFF_STASH_CLI_INVENTORY or the same account's saved context. Interactive terminals can choose missing scope; scripts must supply or remember it."
	case helpCustomization:
		return "Choose --scope household or inventory explicitly in scripts; terminals offer a keyboard picker. Household and inventory IDs come from --tenant/--inventory, environment or the saved context. Inventory scope requires both IDs; household scope requires only the household. Inventory lists include inherited household definitions."
	case helpConnector:
		return "Server and connector credentials, independent of user sign-in or saved resource scope. Use the OS credential store, or explicitly set STUFF_STASH_CLI_CONNECTOR_CREDENTIAL_FILE on headless hosts."
	default:
		return ""
	}
}
func helpOptionValue(name string) string {
	switch name {
	case "server":
		return "URL"
	case "input":
		return "FILE|-"
	case "output", "journal-dir":
		return "PATH"
	case "tenant", "inventory", "parent", "printer", "type-id", "tag-id", "location-id", "connector":
		return "ID"
	case "limit", "revision", "copies", "template-version":
		return "N"
	case "name", "context":
		return "NAME"
	case "cursor":
		return "CURSOR"
	case "title", "details", "query":
		return "TEXT"
	case "format":
		return "png|pdf"
	case "color":
		return "auto|always|never"
	case "kind":
		return "item|container|location"
	default:
		return "VALUE"
	}
}

// Help remains readable in redirected output without inspecting terminal state.
func helpParagraph(b *strings.Builder, text string) {
	const width = 88
	line := "  "
	for _, word := range strings.Fields(text) {
		if len(line) > 2 && len(line)+1+len(word) > width {
			b.WriteString(line + "\n")
			line = "  "
		}
		if len(line) > 2 {
			line += " "
		}
		line += word
	}
	b.WriteString(line + "\n")
}
