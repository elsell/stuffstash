package bootstrap

func administrationHelp() []commandHelp {
	return []commandHelp{
		{Path: "telemetry submit", Scope: helpAccount, Summary: "Send an explicit batch of existing client measurements.", Options: "input", Confirm: true, Input: "JSON requires measurements (1 to 50). Each needs platform (ios/android/web), operation (request/image), surface (application/home/list/detail/gallery/fullscreen/upload), variant (none/small/medium/large/original), outcome (success/failure/cancelled) and durationMs (0 to 60000). Optional $schema. No automatic CLI collection or retries.", Example: "telemetry submit --input measurements.json --yes"},
		{Path: "evaluation cases create", Scope: helpHousehold, Summary: "Create an evaluation case and its first revision.", Options: "input", Confirm: true, Input: evaluationDefinitionHelp, Example: "evaluation cases create --tenant HOME --input case.json --yes"},
		{Path: "evaluation revisions create", Arguments: "CASE_ID", Scope: helpHousehold, Summary: "Append an evaluation case revision at the expected revision.", Options: "input", Confirm: true, Input: evaluationDefinitionHelp + " Also supply expectedRevision as a positive 64-bit integer. No latest-revision lookup or automatic retry occurs.", Example: "evaluation revisions create CASE_ID --tenant HOME --input revision.json --yes"},
		{Path: "evaluation runs create", Scope: helpHousehold, Summary: "Queue a background text-only evaluation run.", Options: "input", Confirm: true, Input: "JSON requires workflowId, revisionId and cases (1 to 100 objects with caseId and revisionId); optional $schema is accepted. This queues provider work, does not activate a workflow, and does not retry automatically. Inspect runs list/show before repeating an uncertain request.", Example: "evaluation runs create --tenant HOME --input run.json --yes"},

		{Path: "evaluation runs cancel", Arguments: "RUN_ID", Scope: helpHousehold, Summary: "Cancel an evaluation run at its expected version.", Options: "input", Input: "JSON requires expectedVersion as a positive 64-bit integer; optional $schema is allowed. Interactive terminals read the current version before confirmation. Scripts must supply --input FILE|- and --yes. The CLI does not retry cancellation automatically.", Confirm: true, Example: "evaluation runs cancel RUN_ID --input cancellation.json --yes"},
		{Path: "provider-profiles list", Scope: helpHousehold, Summary: "List provider profiles."},
		{Path: "provider-profiles show", Arguments: "PROFILE_ID", Scope: helpHousehold, Summary: "Inspect a provider profile without credential values."},
		{Path: "provider-profiles enable", Arguments: "PROFILE_ID", Scope: helpHousehold, Summary: "Enable a provider profile.", Confirm: true},
		{Path: "provider-profiles disable", Arguments: "PROFILE_ID", Scope: helpHousehold, Summary: "Disable a provider profile.", Confirm: true},
		{Path: "provider-profiles archive", Arguments: "PROFILE_ID", Scope: helpHousehold, Summary: "Archive a provider profile.", Confirm: true},
		{Path: "provider-profiles test", Arguments: "PROFILE_ID", Scope: helpHousehold, Summary: "Contact the configured provider and record its test result.", Confirm: true},
		{Path: "voice-provider show", Scope: helpHousehold, Summary: "Inspect household voice-provider readiness and selected profiles."},
		{Path: "workflows create", Scope: helpHousehold, Summary: "Create a workflow and its first revision.", Options: "input", Input: "Required JSON: definition with name, budget {elapsedSeconds, followUpTurns, modelCalls, toolCalls}, optional instructions/providerProfileId; optional $schema. Use --input FILE or --input -.", Confirm: true, Output: "Complete created revision and response metadata.", Example: "workflows create --input workflow.json --yes"},
		{Path: "workflows revisions create", Arguments: "WORKFLOW_ID", Scope: helpHousehold, Summary: "Append a workflow revision without activating it.", Options: "input", Input: "Required JSON: definition {name, budget {elapsedSeconds, followUpTurns, modelCalls, toolCalls}, optional instructions/providerProfileId} and positive integer expectedRevision; optional $schema. Use --input FILE or --input -.", Confirm: true, Output: "Complete appended revision and response metadata.", Example: "workflows revisions create WORKFLOW_ID --input revision.json --yes"},
		{Path: "workflows activate", Arguments: "WORKFLOW_ID", Scope: helpHousehold, Summary: "Change the selected workflow to an evaluated revision.", Options: "input", Input: "Required JSON: revisionId, runId, cases [{caseId, revisionId}] (or null); optional expected {workflowId, revisionId} (or null) and $schema. Use --input FILE or --input -. The server validates evaluation evidence.", Confirm: true, Output: "Complete activated revision and response metadata.", Example: "workflows activate WORKFLOW_ID --input activation.json --yes"},
		{Path: "workflows list", Scope: helpHousehold, Summary: "List workflow heads.", Options: "limit cursor"},
		{Path: "workflows show", Arguments: "WORKFLOW_ID", Scope: helpHousehold, Summary: "Inspect the latest workflow revision."},
		{Path: "workflows revisions list", Arguments: "WORKFLOW_ID", Scope: helpHousehold, Summary: "List workflow revisions.", Options: "limit cursor"},
		{Path: "workflows revisions show", Arguments: "WORKFLOW_ID REVISION_ID", Scope: helpHousehold, Summary: "Inspect one workflow revision."},
		{Path: "workflows selection show", Scope: helpHousehold, Summary: "Show the currently selected workflow and revision."},
		{Path: "evaluation cases list", Scope: helpHousehold, Summary: "List evaluation cases.", Options: "limit cursor"},
		{Path: "evaluation cases show", Arguments: "CASE_ID", Scope: helpHousehold, Summary: "Inspect an evaluation case revision."},
		{Path: "evaluation runs list", Scope: helpHousehold, Summary: "List evaluation runs.", Options: "limit cursor"},
		{Path: "evaluation runs show", Arguments: "RUN_ID", Scope: helpHousehold, Summary: "Inspect an evaluation run."},
		{Path: "evaluation revisions list", Arguments: "CASE_ID", Scope: helpHousehold, Summary: "List evaluation case revisions.", Options: "limit cursor"},
		{Path: "evaluation revisions show", Arguments: "CASE_ID REVISION_ID", Scope: helpHousehold, Summary: "Inspect one evaluation case revision."},
	}
}

const evaluationDefinitionHelp = "JSON requires definition with title, utterance and expectations.kind. Optional assets is an array of objects with id, title, kind and optional description, parentId, tagNames. Expectations accepts referencedAssets, locations (assetId, ancestorId), proposals (operation and optional targetId, destinationId, newTitle, newKind, details), and forbiddenOperations. Optional $schema is accepted. Use --input FILE|-; scripts require --yes. Inspect current cases before repeating an uncertain request."
