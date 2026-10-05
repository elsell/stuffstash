# CLI JSON input diagnostics

JSON file and stdin diagnostics distinguish empty input, a valid JSON value that
is not an object, and malformed JSON. Malformed JSON reports a one-based byte
position in the original input, including leading whitespace, so the user can
locate the error. Diagnostics give a corrective action and retain the `input`
error category. They never print input values, raw parser errors, or credentials.

Validation accepts exactly one complete object and preserves its original bytes,
including large numbers, explicit nulls and formatting. Required-field and domain
validation remain with the relevant command. Critical tests cover safe syntax
locations, invalid root/empty input, trailing values and preservation of valid
input; these do not require per-message snapshots or new dependencies.
