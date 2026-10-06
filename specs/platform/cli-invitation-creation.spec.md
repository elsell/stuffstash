# CLI invitation creation

`invitations create --email ADDRESS --role viewer|editor` creates an invitation
in the resolved inventory. On a terminal, missing email and role use text input
and a keyboard role picker. Scripts can use flags or `--input FILE|-` containing
email and relationship. Do not combine these input methods. Preserve exact JSON
through the generated SDK. Reject unknown roles before sending a request.

Show the resolved scope and email before confirming this access-related action;
scripts require `--yes`. Successful output includes the complete invitation
response and its one-time inviteUrl. Human output explains that the link is shown
only at creation and must be saved or shared now. List/show continue to omit it.
No logs or notices copy the link, and no email or browser is opened by the CLI.

Creation has no retry key. Definite authorization/validation failures retain safe
error categories; ambiguous transport failure tells users to inspect invitations
before repeating the command. Never retry automatically. Critical command-boundary
tests cover exact fields, authentication and inventory scoping, confirmation,
non-interactive missing input, one-time link output and denied requests.

A successful response without an invitation ID or one-time link is an unknown outcome. The CLI must not report success or encourage sharing an empty link. It directs the user to inspect invitations before retrying.

## Verification evidence

The combined CLI suite passes on Paul after integration with search, media,
customization, archive/import, and provider setup. Critical command tests cover
confirmation, missing scripted input, one-time link output, authorization denial,
and a malformed success response without its required link. The adapter rejects
missing invitation ID/link; the command reports an uncertain result rather than
claiming usable invitation creation. Go structural/domain-import checks and the
API coverage check pass. The integrated manifest records 180 of 192 operations
implemented; this batch is not a claim of complete parity or a release.
