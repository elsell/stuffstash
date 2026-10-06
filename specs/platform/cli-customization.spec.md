# CLI customization commands

## Approved command surface

Expose `asset-types` and `field-definitions` with `list`, `show ID`,
`create`, `update ID`, `archive ID`, `restore ID`, and `delete ID`.
Each command requires an explicit `--scope household|inventory` in scripts.
Interactive callers without scope select it with the existing keyboard picker;
absence of an inventory ID never chooses household scope. Existing context
resolution supplies the selected scope's authorized household and inventory.

Lists accept lifecycle, limit, and cursor filters. Details and mutations retain
all declared response fields and envelope metadata. Human output quotes values
and includes scope and lifecycle. JSON retains transport values without hiding
null, empty arrays, or false. All transport remains behind existing definition
ports and generated SDK adapters.

## Inputs and changes

Create and update accept bounded exact JSON using `--input FILE|-`, including
all declared request fields. Validate structure before authentication; retain
original JSON bytes for the server. Reject unknown fields and mixing JSON with
field flags. The server decides compatibility, uniqueness, deletion eligibility,
and immutable field constraints; the CLI does not silently change requests.

Common creation fields are available as named flags and interactive prompts:
key, display name, and field type. Asset types also expose description and an
explicit boolean expiration flag. Field enum options are supplied through JSON
or guided entry. Updates expose display name and asset type description and
expiration; other declared updates use JSON. Interactive creation requests
missing common required fields through keyboard/text prompts. Guided updates
without fields request a display name. Scripts require all required fields.

Every mutation shows its server, scope, action, and target and requires explicit
confirmation or `--yes`. Cancellation sends no mutation. Archive, restore, and
delete accept no body. Requests are sent once; errors preserve machine categories
and callers inspect current state before retrying an uncertain write. No local
context changes occur merely because an explicit scope flag was supplied.

## Verification

Critical CLI boundary tests cover both scopes and all seven operations for both
domains, exact input and complete results, list filters and pagination, denial,
pre-auth invalid options, missing scripted scope, and canceled confirmation.
Existing adapter tests continue to enforce generated route and wire contracts.
Run the full CLI suite and relevant structural hooks on the validation host,
then obtain code critic review before committing. These commands close only the
28 declared custom asset type and field definition operations.

### Implementation evidence

The CLI boundary suite exercises all 28 routes with a stored signed-in session,
exact request bodies, full JSON result comparison, list filters, confirmation,
and 401/403/409/500 failures without retry or response-body disclosure. Focused
application tests verify pre-auth scope/input rejection, keyboard scope and enum
creation, explicit false, and canceled deletion before API construction.

The full CLI Go suite passes on Paul with `GOWORK=off`; Go formatting, structural
and domain-import hooks pass. The coverage checker records all 28 customization
operations implemented. Code critic review found one guidance/fixture issue: update help now explicitly
omits immutable key/type fields, and the successful update fixture uses mutable
fields only. The complete suite also passes after integration onto the media
transfer branch, preserving search and transfer wiring. Combined coverage is
165 of 192 operations. This evidence does not claim server end-to-end deployment
validation.
