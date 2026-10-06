package bootstrap

const archiveScopeHelp = "Household-wide by default; --inventory explicitly filters jobs. The remembered inventory does not narrow this request."
const importSourceHelp = "Use protected --input FILE|- JSON or guided terminal entry. Required sourceType: legacy_homebox or legacy_homebox_csv. Optional $schema, baseUrl, username, password, includeImages, allowInsecureTLS, allowPrivateNetwork, fileName, contentBase64. Credentials never belong in argv. Private networks and untrusted TLS require explicit opt-in. No automatic retry."

func portabilityHelp() []commandHelp {
	return []commandHelp{
		{Path: "archive-jobs list", Scope: helpHousehold, Summary: "List archive and restore jobs.", Options: "inventory limit cursor", Input: archiveScopeHelp},
		{Path: "archive-jobs show", Arguments: "JOB_ID", Scope: helpHousehold, Summary: "Inspect an archive or restore job.", Options: "inventory", Input: archiveScopeHelp},
		{Path: "archive-jobs preview", Arguments: "JOB_ID", Scope: helpHousehold, Summary: "Review restore counts and key remappings.", Options: "inventory", Input: archiveScopeHelp},
		{Path: "archive-jobs create", Scope: helpInventory, Summary: "Create an inventory archive.", Options: "input idempotency-key", Confirm: true, Input: "JSON requires inventoryId, photos and otherFiles (explicit booleans), plus optional $schema. A terminal can choose scope and file inclusion. Preserve the reported retry key after an uncertain response.", Example: "archive-jobs create --tenant HOME --input archive.json --yes"},
		{Path: "archive-jobs upload", Scope: helpHousehold, Summary: "Upload a ZIP archive for restore review.", Options: "file idempotency-key", Confirm: true, Input: "Supply --file PATH|- with ZIP bytes. No local extraction. Upload does not approve restoration. Preserve the reported retry key after an uncertain response.", Example: "archive-jobs upload --file backup.zip --yes"},
		{Path: "archive-jobs download", Arguments: "JOB_ID", Scope: helpHousehold, Summary: "Download archive ZIP bytes.", Options: "inventory output", Input: archiveScopeHelp + " Supply --output PATH|-; paths are private and never overwritten.", Output: "A path produces download status. --output - writes only ZIP bytes to stdout, diagnostics stay on stderr. Do not combine --output - with --json.", Example: "archive-jobs download JOB_ID --output backup.zip"},
		{Path: "archive-jobs approve", Arguments: "JOB_ID", Scope: helpHousehold, Summary: "Restore a reviewed archive into a new inventory.", Options: "inventory input", Confirm: true, Input: archiveScopeHelp + " JSON requires name and accepts optional $schema. A terminal can ask for the new inventory name. Existing inventories are not replaced.", Example: "archive-jobs approve JOB_ID --input destination.json --yes"},
		{Path: "archive-jobs retry", Arguments: "JOB_ID", Scope: helpHousehold, Summary: "Request one server-controlled archive retry.", Options: "inventory", Confirm: true, Input: archiveScopeHelp + " Inspect the job first; the CLI does not poll or repeat the request."},
		{Path: "archive-jobs delete", Arguments: "JOB_ID", Scope: helpHousehold, Summary: "Delete an archive job and retained content.", Options: "inventory", Confirm: true, Input: archiveScopeHelp},
		{Path: "import-jobs preview", Scope: helpInventory, Summary: "Create an import preview from a supplied source.", Options: "input", Confirm: true, Input: importSourceHelp, Example: "import-jobs preview --input source.json --yes"},
		{Path: "import-jobs start", Arguments: "JOB_ID", Scope: helpInventory, Summary: "Start a reviewed import into this inventory.", Options: "input", Confirm: true, Input: importSourceHelp + " Supply the same source and security options as preview. The server checks their fingerprint.", Example: "import-jobs start JOB_ID --input source.json --yes"},
	}
}
