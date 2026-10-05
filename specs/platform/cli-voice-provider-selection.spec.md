# Voice provider selection update

`voice-provider update --input FILE|-` replaces the household's three provider
selections using the complete API request. `speechToTextProfileId`,
`languageInferenceProfileId`, `textToSpeechProfileId` and optional `$schema` are the
only fields. IDs are strings or null. Preserve the submitted bytes. Reject malformed
objects, unknown fields and incorrect types before authentication.

This is replacement, not patch: omitted, null or empty IDs remove an explicit
selection and use the server's automatic selection for that capability. Automatic
does not mean disabled. Do not fill omitted script fields from current state.
Help and the pre-write review explain this and show all three resulting choices.
The API has no version/CAS field; warn that the full replacement can replace another
administrator's recent selections. Never retry PUT, invent concurrency fields or
claim conflict protection that the API does not provide.

Without input, a terminal user can choose with the existing keyboard picker.
Read current configuration and household provider profiles after scope/authentication.
For each of speech input, language inference and spoken output, offer cancellation,
keep the current explicit selection (or keep automatic), automatic selection, and
non-archived matching profiles. Show profile lifecycle/credential state; disabled
profiles remain selectable because the API allows them. Preserve implicit current
selections as automatic, rather than silently pinning the effective profile. Do not
create profiles, change credentials, infer defaults, or execute voice/chat. If the
API reports an invalid explicit selection without its ID, guided replacement must
stop and request complete explicit input: it cannot safely preserve a hidden ID.
An unavailable explicit ID that is returned remains keepable with a warning.

Show server, household and all three resulting selections before confirmation.
JSON/no-input/redirected use requires explicit --input and --yes. Guided input must
remain an explicit user choice even with --yes; no input in scripts fails early.
Use the existing household selection policy, generated SDK behind ports, complete
show response mapping, safe human/JSON output and injected observability.

Critical tests cover byte-exact input/null/omission semantics, authentication and
wrong household, no retry on conflict/uncertain result, invalid input before auth,
guided keep/automatic/filtering and cancellation without PUT. Full CLI suite,
relevant hooks and critic review precede completion.
