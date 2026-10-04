# Connected Android label acceptance

The normal application at `9323ac8d39e1c38c4180fd1dbfb8562f75bfb0a1`
passed the following software flows on the owned API 36 / Android 16 emulator
`stuffstash-audit-api36`, with Pixel 6 dimensions and normal text. This was the
production app, without fixture routes, injected sessions, or intercepted API calls.

[Hosted build 37176623392](https://github.com/elsell/stuffstash/actions/runs/37176623392)
published the APK and source/checksum files. Downloaded artifact SHA-256:
`fc7a610cf8d00e3df450270f48b7bb5fb3f73494a71e945d1d0f8928e87ac20a`.
APK SHA-256: `291f009e49d869bb4dd3ca4b08bf84317be741d3fe1ec7942dc9f78fe0a4211d`.
Both checksums and the source revision were verified before installation.

## Results

| Journey | Observed result | Evidence |
| --- | --- | --- |
| Camera permission denied | Browse → Scan label presented the actual Android permission dialog. One **Don't allow** action returned stable Settings/paste guidance without another automatic prompt. | [Denied](permission-denied.png), [fresh hierarchy](permission-denied.json) |
| Invalid label | Entering `not-a-label` in the label-link field showed the unsupported-label message and retained editable input. | [Invalid](invalid-label.png) |
| Foreign instance | A correctly shaped link with a different instance ID was rejected with a change-server explanation. | [Foreign instance](foreign-instance.png) |
| Old hostname, authorized label | Replacing the foreign ID with the configured instance's ID, while retaining `https://old.example`, opened the existing synthetic asset through the configured API. | The same destination is shown in the sign-in-return capture below. |
| Cancel and return | Reopening the scanner and choosing its native Cancel action returned to Browse with the original inventory context and tabs. | [Browse return](cancel-return-browse.png) |
| Signed-out label and real sign-in | Switch account returned to Connect. An Android VIEW intent for the asset's `stuffstash://labels/v1/...` link showed a waiting-label notice. Real Dex browser login returned to the app; the retained **Open label** action opened the intended asset. | [Waiting label](signed-out-label-waiting.png), [authorized destination](real-dex-return-asset.png) |
| Unrelated account | A second real Dex login returned to that account's separate, empty household. Opening the same pending label showed **Label unavailable**, without the owner's asset content. | [Denied](unrelated-account-denied.png), [fresh hierarchy](unrelated-account-denied.json) |

Text was entered in the label-link field using Android input events; these checks
do not claim clipboard integration or camera image decoding. No physical print
was requested. The sign-in path used an app-owned custom-scheme intent, not a
phone camera or verified HTTPS association.

## Regression and scope

The earlier normal APK at `c0acbaf2ae238a173b8bdfe76607b4ef7533307b`
([build 37175277898](https://github.com/elsell/stuffstash/actions/runs/37175277898))
repeated the OS prompt after denial and failed to settle on denial guidance.
Backgrounding for the permission prompt unmounted the permission owner, losing
its eventual result. The corrected adapter retains that request while the route
is focused, but stops actual capture whenever the app is inactive. A deferred
permission fake reproduced duplicate requests for both grant and denial before
the fix; both transitions passed afterward. The complete mobile suite passed
375 files / 2,194 tests, plus type, structural and localization checks.

The isolated Paul audit server used the pinned v0.42.0 API image
`ghcr.io/elsell/stuffstash@sha256:7742756f3716bbe3450d26dd4bc6a3036e071ed41bdd237f5ddbb2ebd0d67654`
with existing synthetic accounts, PostgreSQL and SpiceDB. Migrations and the
idempotent label bootstrap ran without replacing volumes or identity. The exact
test CA was installed in the disposable emulator's trust store; hostname checks
remained enabled. No app TLS bypass or production account/data was used.
Authentication screens, credentials and callback parameters are not retained.

This does not establish physical QR readability, camera decoding, iOS behavior,
Android enlarged-text/TalkBack behavior, universal-link association, or every
permission/recovery condition. Those remain separate acceptance work. After
sign-in, the retained-label banner still says “Sign in to open it” even though
its Open label action works. A subsequent copy-only correction uses “Your label
is ready to open” in the authenticated notice and leaves the signed-out prompt
unchanged. These native results remain pinned to the earlier APK above.

The owned emulator was stopped after testing. The unrelated emulator was not
changed. Host resource pressure interrupted initial setup; the preserved audit
AVD was moved to a task-specific temporary directory rather than wiped. The exact
operational recovery details remain in the local preparation record and must be
consulted before removing that temporary directory or restarting the audit.
