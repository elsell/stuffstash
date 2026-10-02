# Android archive acceptance — October 2, 2026

**Not accepted.** The emulator booted, but the app did not compile. None of the
three planned runtime checks ran at source `588e415326463566c63abf4e061e89663627d383`.
This is environment evidence, not a reproduced product failure.

| Check | Intended boundary | Result |
| --- | --- | --- |
| Upload integrity | Existing `ArchiveTransferFixture`, real Android file and native upload to a controlled loopback peer; 1 MiB length/hash and synthetic authentication headers | Not run |
| Rejection and cancellation | Same fixture: redirect/oversized-response rejection, subsequent successful upload and cancellation | Not run |
| Restore review and approval | Existing `InventoryArchiveFixture`, real task controls with controlled repository/picker; preview, close without approval, reopen, approve and open destination | Not run |

The existing disposable Android source tree and pinned SDK/JDK were reused. A
missing custom AVD-directory variable and insufficient root disk initially blocked
startup. Removing 5.1 GiB of stale synthetic emulator userdata allowed a fresh boot.
The separate project's emulator was untouched. No user inventory, source or retained
screenshots were deleted.

Dependency installation and Expo prebuild succeeded. Gradle then rejected modified
immutable generated cache workspaces. Removing one Groovy entry exposed a Kotlin
entry failure; regenerating both DSL cache directories exposed transform-cache
failures. The final [sanitized build output](build-failure.txt) and
[scope/result record](result.json) retain the terminal result. An older APK left
in the reused build tree was not installed or counted as current evidence.

The investigation stopped. The audit emulator was shut down and quarantined
corrupt generated caches were removed. A future attempt needs a fresh generated
Gradle cache or another known-good build environment, with an explicit resource
budget; do not repeat this attempt unchanged or continue deleting entries one at
a time. An observation timeout alone is not a failed build.

Even a future pass of these fixtures would not prove physical document-provider
behavior, TalkBack operation, or authenticated server restore. Those requirements
remain separate. This setup failure does not gate independently verified releases.
