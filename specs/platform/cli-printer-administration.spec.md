# Printer administration writes

Commands: `printers create`, `printers update PRINTER_ID`, `connectors print update
CONNECTOR_ID`, and `print-settings update`. All use the existing selected household
and inventory, picker and saved-context policy. These are authenticated human
administration operations, not connector credential or physical execution flows.

All accept `--input FILE|-` with the complete REST JSON object, retaining original
bytes, optional schema, omitted fields, nulls, false values and exact integers.
Validate required structure, known fields and JSON types before authentication.
Printer updates require the supplied revision; connector updates require generation;
settings updates require revision and the complete default-printer, print-on-create
and template/options fields. Do not fetch or substitute concurrency values.

Printer creation requires an explicit --idempotency-key (1–200 characters) for
the API retry contract; uncertain creation instructs reuse of the same key and
unchanged request after inspection. Updates reject retry keys.

Printer creation additionally accepts --name, --adapter, --label-size (preset ID)
and --preset-version. On a terminal, prompt for missing required scalars through
the text-input port; scripts must supply them. Never combine field options with
JSON. Updates require JSON rather than ambiguous partial flag defaults. Preset and
template selection remain server validated; do not invent defaults or auto-select.

Show server, household, inventory, target and effect before confirmation. Scripts,
JSON and no-input mode require --yes. Connector confirmation explains that supplied
printer bindings/revocation may change access. Print settings confirmation explains
that defaults may change automatic label printing. Never retry writes or overwrite
conflicts; uncertain results direct users to existing inspection commands.

Generated SDK operations sit behind a printer administration port. Responses reuse
existing complete printer, connector and settings inspection mappings/presentation.
The generated SDK narrows unsigned REST counters and versions. Shared printer,
connector and settings response decoding therefore uses project-owned typed
envelopes behind the same SDK request methods, preserving uint64 counters and
uint32 versions. Boundary tests cover their upper ranges.

No dependency, generic raw-request command, pairing, credential rotation, physical
printer communication or pending scope-policy change is introduced.

Critical tests cover exact bodies/envelopes, omissions/null/false/concurrency,
authenticated scope and authorization denials, declined confirmation, no retry on
conflict, invalid input before credentials and guided/scalar creation. Full CLI
suite, relevant hooks and code critic precede finalization.
