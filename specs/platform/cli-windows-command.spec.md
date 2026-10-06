# Windows authenticated CLI boundary

A native Windows test must exercise a scripted command through bootstrap using
an actual OS credential-store entry and a controlled HTTP server. Use a unique
server URL, synthetic credentials and explicit household/inventory scope. Verify
that the SDK sends the stored credential to the selected server, preserves scope,
returns JSON, rejects a forbidden scope without retrying, and fails before a
request when the credential is absent. Diagnostics must not expose credentials.
Delete the test entry on completion, including failure paths.

This adds no production authentication behavior. It tests the Windows OS-store
to command to HTTP boundary. Browser/device-code sign-in and physical interactive
terminal behavior remain separate verification requirements.
