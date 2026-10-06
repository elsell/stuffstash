# Complete label render request and result

Extend existing `labels render ASSET_ID --output PATH` with `--input FILE|-`
for the complete render JSON: media, template, format and optional schema.
This keeps its established file destination, private non-overwriting publication,
content bounds, signature/hash checks, scope and authorization behavior. It does
not decide the pending general download destination policy or add raw artifact
access. Existing flag/default selection remains supported.

Structured input preserves all declared fields, omitted optional fields, false
options and exact uint32 media/template versions. Validate required structure,
known fields, JSON types, PNG/PDF format and version range before credentials.
Do not combine structured input with media/template/format selection flags. The
body defines the output format. Do not fetch or replace defaults or selection
when an explicit body is supplied. Keep server-side media validation authoritative.

Use generated SDK request methods with an exact JSON body behind existing ports.
Replace signed narrowing for flag/default rendering as well. Keep generated
source unchanged. Existing automatic label-identity assignment remains part of
the render workflow; never retry render creation or downloads automatically.

Retain the complete declared render response envelope (schema, metadata, ID,
selection/media fingerprints, content type/path, checksum, expiry, dimensions and
rotation). Include it as an additional `render` property in the existing JSON file
result, retaining path, format and checksum. Human output remains a concise saved
file result; the complete render metadata is available with --json. Do not follow
returned contentPath to another host: use the generated scoped content route by
render ID and preserve redirect refusal. Preserve existing PNG/PDF signature,
checksum and bounded-read protections before publishing the private file.

Critical tests cover exact structured bytes including false/maxuint32/schema;
no selection lookup or mixed options; complete safe render metadata; authenticated
scope/denial; and existing content integrity, redirect and no-overwrite checks.
The independent render-ID download operation remains partial pending its separate
command and destination decision. Completing POST rendering must not mark that
GET operation complete by association.
