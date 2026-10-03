# Android onboarding and archive acceptance

The normal release-mode APK and isolated server are identified in
[runtime-result.json](runtime-result.json). The hosted build source and rebased
integration commit have the same tree. No fixture routes, injected sessions or
intercepted responses were used. Test certificate and hostname setup follows
[the connected Android record](../android-connected/README.md).

Fresh browser sign-in and household creation now return directly to Home
(`setup-home.png`), resolving the previously recorded Containing location fallback.
Owner sign-in also returned to Home.

Authenticated export reached Ready and offered the downloaded ZIP through the
Android share sheet (`export-ready.png`, `export-share.png`). This emulator offered
no local-save recipient. ADB copied the **actual downloaded app-cache ZIP** into
Downloads; this bridge is not evidence of recipient saving. The system document
picker then selected that ZIP for upload. Validation showed six items and four
photos (`restore-review.png`). Closing and reopening the staged review retained
its counts. Renaming the destination, approving restore and opening the resulting
inventory succeeded (`restored-home.png`). Opening a restored original photo
rendered its image (`restored-photo.png`). This fixture had no tags, custom fields,
custom types or non-photo files; richer API/browser coverage is separate evidence.

TalkBack 16.0 was enabled and visible accessibility focus was observed
(`talkback-partial.png`). A complete focus/activation journey was **not verified**.
The emulator had no audio, so spoken output was not assessed. Later Home/Browse
refresh errors appeared during this inspection; their cause remains undiagnosed.
Do not infer either a TalkBack defect or a passing accessibility journey from this
partial check. Original disabled accessibility settings were restored afterward.

Physical recipient saving, iPhone, VoiceOver and complete TalkBack acceptance
remain open. This evidence supports shipping the verified onboarding fix without
turning unavailable evidence into a passing release gate.
