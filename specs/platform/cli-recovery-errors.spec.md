# CLI recovery error guidance

CLI errors explain the failed operation and a concrete next action when known. This pass improves existing context storage, credential storage, OIDC configuration, print configuration, and label file messages without changing endpoint coverage, authentication behavior, security checks, or machine error categories.

- Context storage distinguishes failures to load existing configuration from failures to save new configuration. A failure while reading existing data during an update remains a read failure. Size and format errors must not falsely suggest that saving data succeeded.
- Credential errors distinguish non-regular paths from permissions that allow other accounts to access the file. Existing ownership, file type, and permission checks remain mandatory. System credential storage failures explain how to retry sign-in or configure the existing private-file alternative.
- OIDC configuration and discovery errors direct the caller to the appropriate connection check or server administrator without exposing provider responses or inventing a cause.
- Invalid print duration messages name the environment variable, accepted units/example, and its actual supported bounds.
- Label output errors identify the output file operation and a safe next action. An existing destination receives specific guidance to choose another output path. Publication remains atomic and never replaces existing files or symbolic links.

Authored guidance uses short sentences and direct instructions consistent with the CLI's ASD-STE100 writing goal. This bounded review is not a complete dictionary audit or certification.

The available-reference review applies clear, complete sentences, active voice,
one recovery instruction per sentence, and consistent technical names. Prefer
instructions of at most 20 words. Preserve literal command and flag spelling.
Review a message in its failure context: do not invent a cause, promise a safe
retry after an uncertain mutation, or tell a user to repeat a physical print.
Sign-in failures identify the sign-in step and give an available recovery action;
invalid provider responses direct persistent failures to the server administrator.
Print protocol failures direct users to inspect the job or connection before
retrying. Error categories, exit codes, security checks, and redaction remain
unchanged. This review uses the publicly available STE writing-rule summary;
full Issue 9 dictionary and grammar conformance remains a separate unverified
requirement, not a claim made by these copy changes.

A lease timeout can occur after physical print submission. Its shared error must
not claim that no output started. It must tell the user to inspect the job and
printer before any retry. Critical verification expires the lease after a device
submission and preserves the uncertain outcome and journal without a second
submission. The safety margin can stop work before the nominal expiry, so the
message must not assert that the lease has already expired.

Reference: [Simplified Technical English writing rules](https://en.wikipedia.org/wiki/Simplified_Technical_English#Writing_rules).

Critical verification covers classification of context read/save failures, distinct credential type/permission errors with unchanged rejection, and existing label destinations remaining intact. Existing authentication, file security, and CLI tests remain required. Do not add copy-only tests for every reworded message.

Streaming input guidance must be valid for both `--file` transfers and `--input`
imports. Shared errors refer to a file path instead of naming the wrong flag.
Binary downloads distinguish an existing destination from other publication
failures. They explain that the existing file was not changed and ask for another
output path, while preserving atomic no-overwrite behavior on Unix and Windows.
Invalid file-backed connector registration explains how to select a new private
`STUFF_STASH_CLI_CONNECTOR_CREDENTIAL_FILE` path and pair again. Never delete or
replace the invalid credential store automatically. System credential storage
can replace an invalid entry through pairing, so its guidance asks users to pair
again without recommending unsupported file credentials on Windows.
