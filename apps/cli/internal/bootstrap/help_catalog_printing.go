package bootstrap

func printingHelp() []commandHelp {
	return []commandHelp{
		{Path: "labels show", Arguments: "ASSET_ID", Scope: helpInventory, Summary: "Inspect an asset label identity."},
		{Path: "labels assign", Arguments: "ASSET_ID", Scope: helpInventory, Summary: "Assign a stable label identity without rendering or printing.", Confirm: true},
		{Path: "labels templates", Scope: helpInventory, Summary: "List available label templates."},
		{Path: "labels resolve", Arguments: "LABEL_URL", Scope: helpAccount, Summary: "Resolve a label link to an accessible asset."},
		{Path: "labels render", Arguments: "ASSET_ID", Scope: helpInventory, Summary: "Render a label into a new private file.", Options: "format output printer media-preset width-mm height-mm template template-version show-reference", Input: "Supply --output PATH. Format is png (default) or pdf. Choose one of --printer, --media-preset, or paired exact catalog --width-mm/--height-mm; otherwise the default printer is used.", Example: "labels render ASSET_ID --format png --output label.png --printer PRINTER_ID", Output: "Writes a new private file; existing paths are never overwritten. The result reports its path; --json prints that result as JSON."},
		{Path: "labels print", Arguments: "ASSET_ID", Scope: helpInventory, Summary: "Queue an asset label for printing.", Options: "printer template template-version copies show-reference idempotency-key", Input: "Use the inventory defaults or override the printer/template selection. After an uncertain request, reuse the same retry key and unchanged selection."},
		{Path: "print-settings show", Scope: helpInventory, Summary: "Inspect default printer and label settings."},
		{Path: "printers profiles", Scope: helpInventory, Summary: "List printer adapter profiles."},
		{Path: "printers list", Scope: helpInventory, Summary: "List registered printers.", Options: "limit cursor"},
		{Path: "printers show", Arguments: "PRINTER_ID", Scope: helpInventory, Summary: "Inspect a registered printer."},
		{Path: "printers configure", Arguments: "PRINTER_ID", Scope: helpInventory, Summary: "Set a supported label media preset.", Options: "label-size", Input: "Supply --label-size PRESET_ID from the supported printer catalog.", Example: "printers configure PRINTER_ID --label-size PRESET_ID"},
		{Path: "printers test", Arguments: "PRINTER_ID", Scope: helpInventory, Summary: "Queue one test label.", Options: "template template-version show-reference idempotency-key"},
		{Path: "print-jobs list", Scope: helpInventory, Summary: "List print jobs.", Options: "printer limit cursor"},
		{Path: "print-jobs show", Arguments: "JOB_ID", Scope: helpInventory, Summary: "Inspect a print job and its current revision."},
		{Path: "print-jobs cancel", Arguments: "JOB_ID", Scope: helpInventory, Summary: "Request print-job cancellation; output may already have printed.", Confirm: true},
		{Path: "print-jobs reprint", Arguments: "JOB_ID", Scope: helpInventory, Summary: "Queue another copy of a print job.", Options: "printer template template-version copies show-reference idempotency-key"},
		{Path: "print-jobs resolve", Arguments: "JOB_ID", Scope: helpInventory, Summary: "Record a report about uncertain physical output.", Options: "input", Input: "JSON requires reportedOutcome (printed, not_printed or unknown), positive revision, and acknowledgeUncertainty: true. Interactive terminals offer these choices. Scripts must use --input FILE|-. This records a report; it does not verify output or print again.", Confirm: true, Example: "print-jobs resolve JOB_ID --input resolution.json --yes"},
		{Path: "connectors print list", Scope: helpInventory, Summary: "List print connectors.", Options: "limit cursor"},
		{Path: "connectors print show", Arguments: "CONNECTOR_ID", Scope: helpInventory, Summary: "Inspect a print connector."},
		{Path: "connectors print register", Scope: helpConnector, Summary: "Register a connector through browser approval.", Options: "name", Input: "Supply --name NAME and connect a supported local printer for discovery. A browser user approves the displayed URL and short code; connector credentials are stored locally.", Example: "connectors print register --server https://stash.example --name \"Home printer\""},
		{Path: "connectors print rotate", Scope: helpConnector, Summary: "Rotate a connector credential through browser approval.", Options: "connector", Input: "Supply --connector CONNECTOR_ID. Browser approval is required; the replacement credential is stored locally.", Example: "connectors print rotate --server https://stash.example --connector CONNECTOR_ID"},
		{Path: "connectors print run", Scope: helpConnector, Summary: "Run the Linux USB print worker.", Options: "connector journal-dir", Input: "Supply --connector CONNECTOR_ID or STUFF_STASH_CLI_CONNECTOR_ID. Keep --journal-dir PATH persistent; the worker consumes assigned printers.", Example: "connectors print run --server https://stash.example --connector CONNECTOR_ID --journal-dir /var/lib/stuffstash/print", Output: "This long-running worker emits operational diagnostics until stopped. It does not produce a finite JSON result."},
	}
}
