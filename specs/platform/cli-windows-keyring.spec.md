# Windows CLI credential-store verification

Windows uses the OS credential store for human sessions. Credential-store read
and save failures must not recommend the unsupported credential-file fallback.
On Unix, the explicit private-file fallback remains available. Messages preserve
safe, actionable recovery without including backend errors or credential values.

The native Windows workflow verifies the real credential adapter with a unique,
random test server identity and synthetic session. It must preserve the session,
isolate other server keys, reject malformed or mismatched stored sessions, and
support deletion and missing-entry deletion. Always clean up the test entry.
No live identity provider, user account or real token is used. The Windows file
credential guard remains unchanged. This proves the OS storage boundary only;
authenticated bootstrap and real terminal/browser interaction remain outstanding.
